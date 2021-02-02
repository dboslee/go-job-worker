package auth_test

import (
	"context"
	"log"
	"net"
	"testing"

	"github.com/dboslee/job-worker/pkg/api"
	"github.com/dboslee/job-worker/pkg/api/proto"
	"github.com/dboslee/job-worker/pkg/auth"
	"github.com/dboslee/job-worker/pkg/core"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/test/bufconn"
)

var listener *bufconn.Listener

func bufDial(ctx context.Context, _ string) (net.Conn, error) {
	return listener.Dial()
}

// NewClient sets up a client using bufconn
func NewClient(ctx context.Context, clientCreds credentials.TransportCredentials, serverCreds credentials.TransportCredentials) (proto.JobServiceClient, func(), error) {

	listener = bufconn.Listen(1024 * 1024)

	service := api.NewJobService(core.NewJobStore())
	s := grpc.NewServer(
		grpc.Creds(serverCreds),
		grpc.UnaryInterceptor(api.AuthUnary),
		grpc.StreamInterceptor(api.AuthStream),
	)
	proto.RegisterJobServiceServer(s, service)

	// Start server in background
	go func() {
		if err := s.Serve(listener); err != nil {
			log.Print(err)
		}
	}()
	close := func() {
		listener.Close()
		s.Stop()
	}

	conn, err := grpc.Dial(
		"0.0.0.0",
		grpc.WithContextDialer(bufDial),
		grpc.WithTransportCredentials(clientCreds),
	)
	if err != nil {
		close()
		return nil, nil, err
	}
	client := proto.NewJobServiceClient(conn)
	return client, close, nil
}

func TestMTLS(t *testing.T) {
	clientCreds, err := auth.LoadClientTLS("test_certs/client1.pem", "test_certs/client1.key", "test_certs/ca.pem")
	if err != nil {
		t.Errorf("error loading creds %v", err)
		return
	}
	serverCreds, err := auth.LoadServerTLS("test_certs/server.pem", "test_certs/server.key", "test_certs/ca.pem")
	if err != nil {
		t.Errorf("error loading creds %v", err)
		return
	}

	ctx := context.Background()
	client, close, err := NewClient(ctx, clientCreds, serverCreds)
	if err != nil {
		t.Errorf("error starting test server %v", err)
		return
	}
	defer close()

	_, err = client.Exec(ctx, &proto.ExecRequest{})
	if err != nil {
		t.Errorf("expected no error got: %v", err)
	}

}

func TestClient(t *testing.T) {
	clientCreds, err := auth.LoadClientTLS("test_certs/client1.pem", "test_certs/client1.key", "test_certs/ca2.pem")
	if err != nil {
		t.Errorf("error loading creds %v", err)
		return
	}
	serverCreds, err := auth.LoadServerTLS("test_certs/server.pem", "test_certs/server.key", "test_certs/ca2.pem")
	if err != nil {
		t.Errorf("error loading creds %v", err)
		return
	}

	ctx := context.Background()
	client, close, err := NewClient(ctx, clientCreds, serverCreds)
	if err != nil {
		t.Errorf("error starting test server %v", err)
		return
	}
	defer close()

	_, err = client.Exec(ctx, &proto.ExecRequest{})
	if err == nil {
		t.Errorf("expected connection error with ca2")
	}

}
