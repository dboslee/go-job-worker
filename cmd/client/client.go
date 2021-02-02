package main

import (
	"context"
	"log"
	"os"

	"github.com/dboslee/job-worker/pkg/api/proto"
	"github.com/dboslee/job-worker/pkg/auth"
	"github.com/dboslee/job-worker/pkg/cli"

	"google.golang.org/grpc"
)

func init() {
	log.SetFlags(0)
}

// Bootstrap client
func main() {
	// TODO: Make certs and port configurable
	tlsCreds, err := auth.LoadClientTLS("certs/client1.pem", "certs/client1.key", "certs/ca.pem")
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}
	conn, err := grpc.Dial("0.0.0.0:8888", grpc.WithTransportCredentials(tlsCreds))
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}

	ctx := context.Background()
	jobServiceClient := proto.NewJobServiceClient(conn)
	cliClient := cli.NewClient(ctx, jobServiceClient)

	err = cliClient.HandleArgs(os.Args)
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}
	os.Exit(0)
}
