package core

import (
	"io"
	"io/ioutil"
	"os"
)

// OutputBuffer provides
type OutputBuffer struct {
	f *os.File
}

// NewOutputBuffer creates a OutputBuffer instance
func NewOutputBuffer() (*OutputBuffer, error) {

	// TODO: Swap out tempfile for a permanent solution
	// This is easier to setup/cleanup
	f, err := ioutil.TempFile("", "job-worker-output-*")
	if err != nil {
		return nil, err
	}

	// TempFile opens the file by default so we will close immediately
	err = f.Close()
	if err != nil {
		return nil, err
	}

	return &OutputBuffer{f: f}, nil
}

// NewReader opens the file read only
func (o *OutputBuffer) NewReader() (io.ReadCloser, error) {
	return os.Open(o.f.Name())
}

// NewWriter opens the file write only
func (o *OutputBuffer) NewWriter() (io.WriteCloser, error) {
	return os.OpenFile(o.f.Name(), os.O_WRONLY, os.ModeAppend)
}

// Remove the file
func (o *OutputBuffer) Remove() error {
	return os.Remove(o.f.Name())
}
