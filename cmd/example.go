package main

import (
	"fmt"
	"time"

	"github.com/dboslee/job-worker/pkg/core"
)

// This is an example showing how the core package can be used.
func main() {
	store := core.NewJobStore()
	job := core.NewJob("test-client-1", "ping", "-c", "10", "8.8.8.8")
	store.Add(job)
	job, _ = store.Get(job.ID)

	ch := make(chan []byte, 64)
	go core.MergeStreams(ch, job.OutStream, job.ErrStream)

	go job.Start()
	go func() {
		timer := time.NewTimer(time.Second * 5)
		<-timer.C
		job.Interrupt()
	}()

	for {
		data, ok := <-ch
		if !ok {
			break
		}
		fmt.Print(string(data))
	}
	fmt.Println(job.ExitCode())

}
