package core

import (
	"os"
	"os/exec"
	"sync"

	uuid "github.com/satori/go.uuid"
)

// Job provides a simple interface for job access and management
type Job struct {
	ID        string
	ClientID  string
	Cmd       *exec.Cmd
	OutStream *Stream
	ErrStream *Stream
	output    []byte
	err       error
	mu        sync.RWMutex
}

// NewJob creates a new job instance
func NewJob(clientID string, command string, args ...string) *Job {
	id := uuid.NewV4().String()

	return &Job{
		ID:        id,
		ClientID:  clientID,
		Cmd:       exec.Command(command, args...),
		OutStream: NewStream(),
		ErrStream: NewStream(),
	}
}

// Exited returns whether the job has exited
func (j *Job) Exited() bool {
	if j.Cmd.ProcessState == nil {
		return false
	}
	return j.Cmd.ProcessState.Exited()
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

// Interrupt sends a SIGINT to the process
func (j *Job) Interrupt() error {
	return j.Cmd.Process.Signal(os.Interrupt)
}

// Kill sends a SIGKILL to the process.
func (j *Job) Kill() error {
	return j.Cmd.Process.Signal(os.Kill)
}

// Start runs a job and handles errors
func (j *Job) Start() error {
	done := make(chan error)
	go func() {
		done <- j.run()
	}()
	err := <-done
	j.UpdateError(err)
	return err
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

	// Setup streams for output
	go j.OutStream.Pipe(stdout)
	go j.ErrStream.Pipe(stderr)

	mergedOutput := make(chan []byte, 64)
	go MergeStreams(mergedOutput, j.OutStream, j.ErrStream)

	// Start the job
	err = cmd.Start()
	if err != nil {
		return err
	}

	// Send on done when cmd exits
	done := make(chan error)
	go func() {
		done <- cmd.Wait()
	}()

	// Read output while the channel is open
	// Return when done
	for {
		select {
		case data, ok := <-mergedOutput:
			if !ok {
				mergedOutput = nil
				break
			}
			j.AppendOutput(data)
		case err := <-done:
			return err
		}
	}

}

// AppendOutput appends bytes to a jobs output
func (j *Job) AppendOutput(b []byte) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.output = append(j.output, b...)
}

// Output returns a jobs output and a channel to stream output from
// If a job has already finished the channel will be closed
func (j *Job) Output() ([]byte, chan []byte) {
	j.mu.RLock()
	defer j.mu.RUnlock()
	ch := make(chan []byte, 64)
	MergeStreams(ch, j.OutStream, j.ErrStream)

	return j.output, ch
}
