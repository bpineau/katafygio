// +build !windows

package log

import (
	"fmt"
	"io"
	"io/ioutil"
	"log/syslog"
	"os"

	"github.com/sirupsen/logrus"
	ls "github.com/sirupsen/logrus/hooks/syslog"
	"github.com/sirupsen/logrus/hooks/test"
)

func getOutput(logServer string, logOutput string) (io.Writer, logrus.Hook, error) {
	var output io.Writer
	var hook logrus.Hook
	var err error

	switch logOutput {
	case outputStdout:
		output = os.Stdout
	case outputStderr:
		output = os.Stderr
	case outputTest:
		output = ioutil.Discard
		_, hook = test.NewNullLogger()
	case outputSyslog:
		output = os.Stderr
		if logServer == "" {
			return nil, nil, fmt.Errorf("syslog output needs a log server (ie. 127.0.0.1:514)")
		}
		hook, err = ls.NewSyslogHook("udp", logServer, syslog.LOG_INFO, "katafygio")
		if err != nil {
			return nil, nil, fmt.Errorf("failed to hook syslog output")
		}
	default:
		output = os.Stderr
	}

	return output, hook, nil
}
