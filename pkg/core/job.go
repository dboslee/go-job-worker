package core

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"sync"

	uuid "github.com/satori/go.uuid"
)

// JobStatus represents the current status of a job
type JobStatus int

const (
	// Pending is the initial status when a job is created
	Pending JobStatus = iota
	// Running is assigned once a job is started
	Running
	// Complete is the status when a job exits without errors
	Complete
	// Error is the status when an error occurs
	Error
)

// String is a convienient way to convert a job status to string
func (js JobStatus) String() string {
	switch js {
	case Running:
		return "running"
	case Complete:
		return "complete"
	case Error:
		return "error"
	default:
		return "pending"
	}
}

// Job provides a simple interface for job access and management
type Job struct {
	ID        string
	ClientID  string
	Cmd       *exec.Cmd
	OutputBuf *OutputBuffer
	status    JobStatus
	err       error
	mu        sync.RWMutex
}

// NewJob creates a new job instance
func NewJob(clientID string, command string, args ...string) (*Job, error) {
	id := uuid.NewV4().String()
	outputBuf, err := NewOutputBuffer()
	if err != nil {
		return nil, err
	}

	return &Job{
		ID:        id,
		ClientID:  clientID,
		Cmd:       exec.Command(command, args...),
		status:    Pending,
		OutputBuf: outputBuf,
	}, nil
}

// ExitCode returns a jobs exit code
func (j *Job) ExitCode() int {
	if j.Cmd.ProcessState == nil {
		return -1
	}
	// This also default to -1 if not exited
	return j.Cmd.ProcessState.ExitCode()
}

// Error returns an error if an error occurred running the job
func (j *Job) Error() error {
	j.mu.RLock()
	defer j.mu.RUnlock()
	return j.err
}

// UpdateError updates a jobs err
func (j *Job) UpdateError(err error) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.err = err
}

// Status returns the status of a job
func (j *Job) Status() JobStatus {
	j.mu.RLock()
	defer j.mu.RUnlock()
	return j.status
}

// UpdateStatus updates a jobs err
func (j *Job) UpdateStatus(status JobStatus) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.status = status
}

// Interrupt sends a SIGINT to the process
func (j *Job) Interrupt() error {
	if j.Cmd.Process == nil {
		return fmt.Errorf("unable to interrupt nil process")
	}
	return j.Cmd.Process.Signal(os.Interrupt)
}

// Kill sends a SIGKILL to the process.
func (j *Job) Kill() error {
	if j.Cmd.Process == nil {
		return fmt.Errorf("unable to kill nil process")
	}
	return j.Cmd.Process.Signal(os.Kill)
}

// Start runs a job and handles errors
func (j *Job) Start() error {
	err := j.run()

	if err != nil {
		log.Print(err)
		j.UpdateError(err)
		j.UpdateStatus(Error)
		return err
	}

	j.UpdateStatus(Complete)
	return nil
}

// Run executes the job and updates its state
func (j *Job) run() error {
	cmd := j.Cmd

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	output := io.MultiReader(stdout, stderr)
	err = cmd.Start()
	if err != nil {
		return err
	}
	j.UpdateStatus(Running)

	w, err := j.OutputBuf.NewWriter()
	if err != nil {
		log.Printf("unable to open log writer %v", err)
	}
	_, err = io.Copy(w, output)
	if err != nil {
		log.Printf("unable to copy output to log %v", err)
	}

	return cmd.Wait()
}
