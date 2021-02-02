package main

import (
	"log"
	"net"
	"os"

	"github.com/dboslee/job-worker/pkg/api"
	"github.com/dboslee/job-worker/pkg/api/proto"
	"github.com/dboslee/job-worker/pkg/auth"
	"github.com/dboslee/job-worker/pkg/core"
	"google.golang.org/grpc"
)

// Bootstrap grpc server
func main() {
	jobStore := core.NewJobStore()
	jobService := api.NewJobService(jobStore)

	// TODO: Make certs and port configurable through env vars, config file, or cli args
	tlsCreds, err := auth.LoadServerTLS("certs/server.pem", "certs/server.key", "certs/ca.pem")
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}

	grpcServer := grpc.NewServer(
		grpc.Creds(tlsCreds),
		grpc.UnaryInterceptor(api.AuthUnary),
		grpc.StreamInterceptor(api.AuthStream),
	)
	proto.RegisterJobServiceServer(grpcServer, jobService)

	var port = "8888"
	conn, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}

	// TODO: Handle graceful shutdowns to clean up any running tasks and open connections
	log.Printf("starting job-worker on port %v", port)
	err = grpcServer.Serve(conn)
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}
}
