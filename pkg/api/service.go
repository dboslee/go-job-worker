package api

import (
	"context"
	"io"
	"log"
	"time"

	"github.com/dboslee/job-worker/pkg/api/proto"
	"github.com/dboslee/job-worker/pkg/core"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type key int

// KeyClientID is the key used to store a client id in a context
const KeyClientID key = iota

// JobNotFound raised when a job is not found
var JobNotFound = status.Error(codes.NotFound, "job not found")

// PermissionDenied raised when a client is not authorized
var PermissionDenied = status.Error(codes.PermissionDenied, "permission denied")

// JobService implements the grpc server interface
type JobService struct {
	jobStore *core.JobStore
}

// NewJobService creats a new JobService instance
func NewJobService(jobStore *core.JobStore) *JobService {
	return &JobService{
		jobStore: jobStore,
	}
}

// getJob only returns jobs for an authorized client
func (js *JobService) getJob(ctx context.Context, id string) (*core.Job, error) {
	job, ok := js.jobStore.Get(id)
	if !ok {
		return nil, JobNotFound
	}
	cID := ctx.Value(KeyClientID)
	if cID == nil || cID.(string) != job.ClientID {
		return nil, PermissionDenied
	}
	return job, nil
}

// Exec handles the grpc ExecRequest
func (js *JobService) Exec(ctx context.Context, req *proto.ExecRequest) (resp *proto.ExecResponse, err error) {
	cID := ctx.Value(KeyClientID)
	if cID == nil || cID.(string) == "" {
		return nil, PermissionDenied
	}
	job, err := core.NewJob(cID.(string), req.Command, req.Args...)
	if err != nil {
		return nil, status.Error(codes.Aborted, "failed to create job")
	}

	js.jobStore.Add(job)
	go job.Start()

	resp = &proto.ExecResponse{Id: job.ID}
	return resp, nil
}

// Stop handles interupting a job
func (js *JobService) Stop(ctx context.Context, req *proto.StopRequest) (resp *proto.StopResponse, err error) {
	job, err := js.getJob(ctx, req.GetId())
	if err != nil {
		return nil, err
	}

	if job.Status() != core.Running {
		return nil, status.Error(codes.FailedPrecondition, "unable to stop job thats not running")
	}

	resp = &proto.StopResponse{Success: true}
	err = job.Interrupt()
	if err != nil {
		log.Printf("error during interrupt %v", err)
		resp.Success = false
		return resp, status.Error(codes.Internal, "failed to stop job")
	}
	return resp, nil
}

// Status gets the status of a given job
func (js *JobService) Status(ctx context.Context, req *proto.StatusRequest) (resp *proto.StatusResponse, err error) {
	job, err := js.getJob(ctx, req.GetId())
	if err != nil {
		return nil, err
	}

	s := job.Status()
	resp = &proto.StatusResponse{
		Status: s.String(),
	}
	if s <= core.Running {
		return resp, nil
	}

	resp.ExitCode = int64(job.ExitCode())

	// TODO: Use a useful error message for the user here
	jobErr := job.Error()
	if jobErr != nil {
		resp.Error = jobErr.Error()
	}
	return resp, nil
}

// Logs streams the logs of a given job
func (js *JobService) Logs(req *proto.LogRequest, serv proto.JobService_LogsServer) error {
	job, err := js.getJob(serv.Context(), req.GetId())
	if err != nil {
		return err
	}

	readErr := status.Error(codes.Internal, "failed to read logs")

	r, err := job.OutputBuf.NewReader()
	if err != nil {
		log.Printf("error getting reader %v", err)
		return readErr
	}
	defer r.Close()

	tick := time.NewTicker(time.Millisecond * 100)
	resp := &proto.LogResponse{}

	// Loop until the job is done running or an error
	b := make([]byte, 1024*4)
	for {
		status := job.Status()

		// If there is a context error return
		if err = serv.Context().Err(); err != nil {
			return err
		}

		// Read the output until we hit io.EOF and the job has exited
		n, err := r.Read(b)
		if err == io.EOF && status > core.Running {
			return nil
		} else if err == io.EOF {
			<-tick.C
			continue
		} else if err != nil {
			log.Printf("error reading logs %v", err)
			return readErr
		}
		resp.Log = b[:n]

		err = serv.Send(resp)
		if err != nil {
			return err
		}
	}
}
