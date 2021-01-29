package core

import (
	"io"
	"io/ioutil"
	"os"
)

// OutputBuffer provides
type OutputBuffer struct {
	name string
}

// NewOutputBuffer creates a OutputBuffer instance
func NewOutputBuffer() (*OutputBuffer, error) {

	// TODO: Swap out tempdir for a permanent solution
	// This is easier to setup/cleanup
	dir, err := ioutil.TempDir("", "job-worker-output-*")
	if err != nil {
		return nil, err
	}
	f, err := os.Create(dir + "/log")
	if err != nil {
		return nil, err
	}

	return &OutputBuffer{name: f.Name()}, nil
}

// NewReader opens the file read only
func (o *OutputBuffer) NewReader() (io.ReadCloser, error) {
	return os.Open(o.name)
}

// NewWriter opens the file write only
func (o *OutputBuffer) NewWriter() (io.WriteCloser, error) {
	return os.OpenFile(o.name, os.O_WRONLY, os.ModeAppend)
}

// Remove the file
func (o *OutputBuffer) Remove() error {
	return os.Remove(o.name)
}
