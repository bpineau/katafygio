// Package log initialize and configure a logrus logger.
package log

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

const (
	outputStdout = "stdout"
	outputStderr = "stderr"
	outputTest   = "test"
	outputSyslog = "syslog"
)

// New initialize logrus and return a new logger.
func New(logLevel string, logServer string, logOutput string) (*logrus.Logger, error) {
	if logLevel == "" {
		logLevel = "info"
	}

	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		return nil, err
	}

	output, hook, err := getOutput(logServer, logOutput)
	if err != nil {
		return nil, fmt.Errorf("failed to init logger: %v", err)
	}

	formatter := &logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	}

	log := &logrus.Logger{
		Out:       output,
		Formatter: formatter,
		Hooks:     make(logrus.LevelHooks),
		Level:     level,
	}

	if logOutput == outputSyslog || logOutput == outputTest {
		log.Hooks.Add(hook)
	}

	return log, nil
}
