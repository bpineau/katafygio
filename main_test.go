package main

import (
	"bytes"
	"testing"

	"github.com/bpineau/katafygio/cmd"
)

func TestMain(t *testing.T) {
	var ok = true

	privateExitHandler = func(c int) {
		ok = false
	}

	cmd.FakeCS = true
	cmd.RootCmd.SetOutput(new(bytes.Buffer))

	// test with normal exit
	cmd.RootCmd.SetArgs([]string{"--help"})
	main()

	if !ok {
		t.Errorf("main() failed")
	}

	// test with failure
	cmd.RootCmd.SetArgs([]string{"--unexpected-arg"})
	main()

	if ok {
		t.Errorf("main() should fail with unexpected arguments")
	}
}

func TestExitWrapper(t *testing.T) {
	var ok = false

	privateExitHandler = func(c int) {
		ok = true
	}

	ExitWrapper(1)

	if !ok {
		t.Errorf("Error in ExitWrapper()")
	}
}
