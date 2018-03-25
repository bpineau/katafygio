// Package log initialize and configure a logrus logger.
package log

import (
	"io"
	"io/ioutil"
	"os"

	"log/syslog"

	"github.com/sirupsen/logrus"
	ls "github.com/sirupsen/logrus/hooks/syslog"
	"github.com/sirupsen/logrus/hooks/test"
)

// New initialize logrus and return a new logger.
func New(logLevel string, logServer string, logOutput string) *logrus.Logger {
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		level = logrus.InfoLevel
	}

	output, hook := getOutput(logServer, logOutput)

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

	if logOutput == "syslog" || logOutput == "test" {
		log.Hooks.Add(hook)
	}

	return log
}

func getOutput(logServer string, logOutput string) (io.Writer, logrus.Hook) {
	var output io.Writer
	var hook logrus.Hook
	var err error

	switch logOutput {
	case "stdout":
		output = os.Stdout
	case "stderr":
		output = os.Stderr
	case "test":
		output = ioutil.Discard
		_, hook = test.NewNullLogger()
	case "syslog":
		output = os.Stderr // does not matter ?
		if logServer == "" {
			panic("syslog output needs a log server (ie. 127.0.0.1:514)")
		}
		hook, err = ls.NewSyslogHook("udp", logServer, syslog.LOG_INFO, "katafygio")
		if err != nil {
			panic(err)
		}
	default:
		output = os.Stderr
	}

	return output, hook
}
