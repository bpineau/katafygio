// +build ignore

// end-to-end tests.
// A kubernetes cluster must be reachable by kubectl and katafygio.
package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"testing"
	"time"
)

var (
	checkTimeout  = 30 * time.Second
	checkInterval = 1 * time.Second
	dumpPath      = "/tmp/kf-e2e-test"
)

type dumpStep struct {
	// a kubectl command to create a dumpable object
	cmd string
	// the counter-command, to purge the created object
	teardown string
	// the file that this command will eventually trigger katafygio to generate or delete
	relpath string
	// wether the file should be created or deleted
	shouldExist bool
}

var testsTable = []dumpStep{
	{"kubectl create ns kf-e2e-test-1", "kubectl delete ns kf-e2e-test-1", "namespace-kf-e2e-test-1.yaml", true},
	{"kubectl run kf-e2e-test-2 --image=gcr.io/google_containers/pause-amd64:3.0", "kubectl delete deploy kf-e2e-test-2", "default/deployment-kf-e2e-test-2.yaml", true},
	{"kubectl expose deployment kf-e2e-test-2 --port=80 --target-port=8000 --name=kf-e2e-test-3", "kubectl delete svc kf-e2e-test-3", "default/service-kf-e2e-test-3.yaml", true},
	{"kubectl delete service kf-e2e-test-3", "", "default/service-kf-e2e-test-3.yaml", false},
	{"kubectl delete namespace kf-e2e-test-1", "", "namespace-kf-e2e-test-1.yaml", false},
	{"kubectl create configmap kf-e2e-test-4 --from-literal=key1=config1", "kubectl delete configmap kf-e2e-test-4", "default/configmap-kf-e2e-test-4.yaml", true},
}

func TestE2E(t *testing.T) {
	for _, tt := range testsTable {
		t.Run(tt.cmd, func(t *testing.T) {
			err := tt.fileExists()
			if err != nil {
				t.Errorf("%s test failed (%v)", tt.cmd, err)
			}
		})

	}
}

func TestMain(m *testing.M) {
	deleteResources()
	_ = exec.Command("rm", "-rf", dumpPath)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		cmd := exec.CommandContext(ctx, "katafygio", "-e", dumpPath)
		_ = cmd.Run()
	}()

	ret := m.Run()
	cancel()
	deleteResources()
	os.Exit(ret)
}

func deleteResources() {
	for _, d := range testsTable {
		if len(d.teardown) == 0 {
			continue
		}

		command := strings.Split(d.teardown, " ")
		_ = exec.Command(command[0], command[1:]...).Run()
	}
}

func (d *dumpStep) fileExists() error {
	checkTick := time.NewTicker(checkInterval)
	timeoutTick := time.NewTicker(checkTimeout)
	defer checkTick.Stop()
	defer timeoutTick.Stop()

	command := strings.Split(d.cmd, " ")
	cmd := exec.Command(command[0], command[1:]...)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s failed with code %v: %s", d.cmd, err, out)
	}

	for {
		select {
		case <-checkTick.C:
			_, err := os.Stat(path.Join(dumpPath, d.relpath))

			if err == nil && d.shouldExist {
				return nil
			}

			if os.IsNotExist(err) && !d.shouldExist {
				return nil
			}

		case <-timeoutTick.C:
			return fmt.Errorf("timeout waiting for %s", d.relpath)
		}
	}
}
