# Job Worker

This project implements a client and server for executing arbitrary linux commands. 

## Prerequisites
- linux amd64
- [go 1.15](https://golang.org/doc/install#download)
- [openssl](https://www.openssl.org/source/)


## Install Guide
```
git clone https://github.com/dboslee/go-job-worker.git
cd go-job-worker
make build
make certs
```

## Starting the server
```./server```

## Client usage
```
./client exec <command> <args>  # Execute a command with optional arguments
./client status <id>            # Get the status of a given job ID
./client stop <id>              # Stop a given job ID
./client logs <id>              # Stream the output of a job
```

Note: 

## Testing
```make test```

## Additional Notes
The server and client are hardcoded to communicate on port 8888.

Although clients are only authorized to access their own jobs through the api this project does not provide any further isolation. Namespaces could be used for better isolation.

This project does not impose any limit on the number of concurrent jobs, the duration of a job, or the resources consumed by a job. A fixed size worker pool, deadlines, and cgroups could be used respectively to enforce these limits.