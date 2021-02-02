package cli

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/dboslee/job-worker/pkg/api/proto"
)

// Client provides a way to interact with the grpc server through the cli
type Client struct {
	ctx        context.Context
	jobService proto.JobServiceClient
}

// NewClient creates a Client instance
func NewClient(ctx context.Context, jobService proto.JobServiceClient) *Client {
	return &Client{
		ctx:        ctx,
		jobService: jobService,
	}
}

// HandleArgs expects the command-line arguments starting with the program name
// and calls the corresponding subcommand or returns an error
func (c *Client) HandleArgs(args []string) error {
	switch len(args) {
	case 0:
		return fmt.Errorf("unknown command")
	case 1:
		return fmt.Errorf("subcommand required")
	}

	subcommand := args[1]
	switch subcommand {
	case "exec":
		return c.exec(args[2:])
	case "status":
		return c.status(args[2:])
	case "stop":
		return c.stop(args[2:])
	case "logs":
		return c.logs(args[2:])
	default:
		return fmt.Errorf("unknown subcommand %v", subcommand)
	}

}

// exec calls the exec rpc and outputs the job id
func (c *Client) exec(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("must provide a command to execute")
	}
	req := &proto.ExecRequest{
		Command: args[0],
		Args:    args[1:],
	}
	resp, err := c.jobService.Exec(c.ctx, req)
	if err != nil {
		return err
	}
	log.Print(resp.GetId())
	return nil
}

// status calls the status rpc and outputs the status
func (c *Client) status(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("must provide a job ID")
	}
	req := &proto.StatusRequest{
		Id: args[0],
	}
	resp, err := c.jobService.Status(c.ctx, req)
	if err != nil {
		return err
	}
	log.Printf("Status: %v", resp.GetStatus())
	log.Printf("ExitCode: %v", resp.GetExitCode())
	log.Printf("Error: %v", resp.GetError())
	return nil
}

// stop calls the stop rpc
func (c *Client) stop(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("must provide a job ID")
	}
	req := &proto.StopRequest{
		Id: args[0],
	}

	// There should be an error if anything goes wrong so we don't really need to check the response
	_, err := c.jobService.Stop(c.ctx, req)
	if err != nil {
		return err
	}
	return nil
}

// logs calls the logs rpc to stream the output of a job
func (c *Client) logs(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("must provide a job ID")
	}
	req := &proto.LogRequest{
		Id: args[0],
	}
	stream, err := c.jobService.Logs(c.ctx, req)
	if err != nil {
		return err
	}

	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}
		log.Print(string(resp.GetLog()))
	}

}
