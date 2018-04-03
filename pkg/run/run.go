// Package run implements the main katafygio's loop, by
// launching the healthcheck service and all known controllers.
package run

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/bpineau/katafygio/config"
	"github.com/bpineau/katafygio/pkg/controller"
	"github.com/bpineau/katafygio/pkg/health"
	"github.com/bpineau/katafygio/pkg/recorder"
	"github.com/bpineau/katafygio/pkg/store/git"
)

// Run launchs the effective controllers goroutines
func Run(config *config.KdnConfig) {
	repos, err := git.New(config).Start()
	if err != nil {
		config.Logger.Fatalf("failed to start git repo handler: %v", err)
	}

	evchan := make(chan controller.Event)

	rec := recorder.New(config, evchan).Start()
	ctl := controller.NewObserver(config, evchan).Start()

	go func() {
		if err := health.HeartBeatService(config); err != nil {
			config.Logger.Warningf("Healtcheck service failed: %s", err)
		}
	}()

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGTERM)
	signal.Notify(sigterm, syscall.SIGINT)
	<-sigterm

	config.Logger.Infof("Stopping all controllers")
	repos.Stop()
	ctl.Stop()
	rec.Stop()
}
