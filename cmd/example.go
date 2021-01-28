package main

import (
	"io"
	"log"
	"time"

	"github.com/dboslee/job-worker/pkg/core"
)

func init() {
	log.SetFlags(0)
}

// This is an example showing how the core package can be used.
func main() {
	store := core.NewJobStore()
	job, err := core.NewJob("test-client-1", "ping", "-c", "5", "8.8.8.8")
	if err != nil {
		log.Fatal(err)
	}

	store.Add(job)
	job, _ = store.Get(job.ID)

	go job.Start()

	// Buffer is nil at first so wait for it to be created
	timer := time.NewTicker(time.Millisecond * 100)

	r, err := job.OutputBuf.NewReader()
	if err != nil {
		log.Fatal(err)
	}

	b := make([]byte, 1024)
	for {
		status := job.Status()
		n, err := r.Read(b)
		if err == io.EOF && status > core.Running {
			break
		} else if err == io.EOF {
			<-timer.C
			continue
		} else if err != nil {
			log.Fatal(err)
		}
		log.Print(string(b[:n]))
	}
}
