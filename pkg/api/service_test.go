package api_test

import (
	"context"
	"testing"

	"github.com/dboslee/job-worker/pkg/api"
	"github.com/dboslee/job-worker/pkg/api/proto"
	"github.com/dboslee/job-worker/pkg/core"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func mockService() *api.JobService {
	return api.NewJobService(
		core.NewJobStore(),
	)
}

func TestAuthorizedAccess(t *testing.T) {
	service := mockService()
	ctx := context.WithValue(context.Background(), api.KeyClientID, "client1")
	execReq := &proto.ExecRequest{Command: ""}

	resp, _ := service.Exec(ctx, execReq)
	statusReq := &proto.StatusRequest{Id: resp.GetId()}
	_, err := service.Status(ctx, statusReq)
	if err != nil {
		t.Errorf("expected no error got: %v", err)
	}

	ctx = context.WithValue(context.Background(), api.KeyClientID, "client2")
	_, err = service.Status(ctx, statusReq)
	if e, _ := status.FromError(err); e.Code() != codes.PermissionDenied {
		t.Errorf("expected permission denied got: %v", e.Code())
	}
}

func TestStatusNotfound(t *testing.T) {
	service := mockService()
	ctx := context.WithValue(context.Background(), api.KeyClientID, "client1")

	req := &proto.StatusRequest{
		Id: "",
	}
	_, err := service.Status(ctx, req)
	if e, _ := status.FromError(err); e.Code() != codes.NotFound {
		t.Errorf("expected not found error got: %v", e.Code())
	}
}
