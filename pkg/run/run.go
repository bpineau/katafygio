// Package run implements the main katafygio's loop, by
// launching the healthcheck service and all known controllers.
package run

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/bpineau/katafygio/config"
	"github.com/bpineau/katafygio/pkg/controllers"
	"github.com/bpineau/katafygio/pkg/health"
	"github.com/bpineau/katafygio/pkg/recorder"
)

// Run launchs the effective controllers goroutines
func Run(config *config.KdnConfig) {
	wg := sync.WaitGroup{}
	wg.Add(len(controllers.AllControllers))
	defer wg.Wait()

	var chans []chan controllers.Event

	for _, c := range controllers.AllControllers {
		ch := make(chan controllers.Event, 100)
		chans = append(chans, ch)
		ctrl := c(config, ch)

		go ctrl.Start(&wg)
		defer func(cont controllers.Controller) {
			go cont.Stop()
		}(ctrl)
	}

	go recorder.Watch(config, chans)

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
}
