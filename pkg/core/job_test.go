package core_test

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"testing"

	"github.com/dboslee/job-worker/pkg/core"
)

func TestNewJobID(t *testing.T) {
	job := core.NewJob("test-client", "")

	if job.ID == "" {
		t.Errorf("expected non empty string")
	}
}

func TestError(t *testing.T) {
	job := core.NewJob("test-client", "")
	job.Start()
	if err := job.Error(); err == nil {
		fmt.Println("expected error")
	}
}

func TestExited(t *testing.T) {
	job := core.NewJob("test-client", "exit", "0")
	job.Cmd = mockExec("exit", "0")
	if job.Exited() {
		t.Errorf("unexpected exit")
	}
	job.Start()
	if !job.Exited() {
		t.Errorf("expected exit")
	}
}

func TestExitCode(t *testing.T) {
	cases := []struct {
		command string
		args    []string
		code    int
	}{
		{"exit", []string{"0"}, 0},
		{"exit", []string{"1"}, 1},
		{"exit", []string{"2"}, 2},
	}

	for _, tc := range cases {
		job := core.NewJob("test-client", tc.command, tc.args...)
		job.Cmd = mockExec(tc.command, tc.args...)
		job.Start()
		if code := job.ExitCode(); code != tc.code {
			t.Errorf("exit code want: %v got: %v", tc.code, code)
		}
	}
}

func TestOutput(t *testing.T) {
	cases := []struct {
		command string
		args    []string
		output  string
	}{
		{"echo", []string{"hello", "world"}, "hello world\n"},
	}

	for _, tc := range cases {
		job := core.NewJob("test-client", tc.command, tc.args...)
		job.Cmd = mockExec(tc.command, tc.args...)
		job.Start()
		if output, _ := job.Output(); string(output) != tc.output {
			t.Errorf("output want: %v got: %v", tc.output, string(output))
		}
	}
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	defer os.Exit(0)

	// Get args after --
	args := os.Args
	for len(args) > 0 {
		if args[0] == "--" {
			args = args[1:]
			break
		}
		args = args[1:]
	}
	cmd, args := args[0], args[1:]
	switch cmd {
	case "echo":
		iargs := []interface{}{}
		for _, s := range args {
			iargs = append(iargs, s)
		}
		fmt.Println(iargs...)
	case "exit":
		n, _ := strconv.Atoi(args[0])
		os.Exit(n)
	default:
		fmt.Fprintf(os.Stderr, "No such command %q\n", cmd)
		os.Exit(2)
	}

}

// Based off of os/exec tests
func mockExec(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}
