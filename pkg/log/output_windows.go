// +build windows

package log

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
)

func getOutput(logServer string, logOutput string) (io.Writer, logrus.Hook, error) {
	var output io.Writer
	var hook logrus.Hook

	switch logOutput {
	case outputStdout:
		output = os.Stdout
	case outputStderr:
		output = os.Stderr
	case outputTest:
		output = ioutil.Discard
		_, hook = test.NewNullLogger()
	case outputSyslog:
		return nil, nil, fmt.Errorf("Syslog output isn't supported on Windows")
	default:
		output = os.Stderr
	}

	return output, hook, nil
}
