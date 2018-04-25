package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/bpineau/katafygio/config"
	"github.com/bpineau/katafygio/pkg/client"
	"github.com/bpineau/katafygio/pkg/controller"
	"github.com/bpineau/katafygio/pkg/event"
	"github.com/bpineau/katafygio/pkg/health"
	"github.com/bpineau/katafygio/pkg/log"
	"github.com/bpineau/katafygio/pkg/observer"
	"github.com/bpineau/katafygio/pkg/recorder"
	"github.com/bpineau/katafygio/pkg/store/git"
)

const appName = "katafygio"

func runE(cmd *cobra.Command, args []string) (err error) {

	resync := time.Duration(resyncInt) * time.Second
	logger := log.New(logLevel, logServer, logOutput)

	if restcfg == nil {
		restcfg, err = client.New(apiServer, kubeConf)
		if err != nil {
			return fmt.Errorf("failed to create a client: %v", err)
		}
	}

	conf := &config.KfConfig{
		DryRun:        dryRun,
		DumpMode:      dumpMode,
		Logger:        logger,
		LocalDir:      localDir,
		GitURL:        gitURL,
		Filter:        filter,
		ExcludeKind:   exclkind,
		ExcludeObject: exclobj,
		HealthPort:    healthP,
		Client:        restcfg,
		ResyncIntv:    resync,
	}

	repo, err := git.New(conf).Start()
	if err != nil {
		conf.Logger.Fatalf("failed to start git repo handler: %v", err)
	}

	evts := event.New()
	reco := recorder.New(conf, evts).Start()
	obsv := observer.New(conf, evts, &controller.Factory{}).Start()
	http := health.New(conf).Start()

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGTERM)
	signal.Notify(sigterm, syscall.SIGINT)
	if !conf.DumpMode {
		<-sigterm
	}

	obsv.Stop()
	repo.Stop()
	reco.Stop()
	http.Stop()

	return nil
}

var (
	restcfg client.Interface

	// RootCmd is our main entry point, launching runE()
	RootCmd = &cobra.Command{
		Use:   appName,
		Short: "Backup Kubernetes cluster as yaml files",
		Long: "Backup Kubernetes cluster as yaml files in a git repository.\n" +
			"--exclude-kind (-x) and --exclude-object (-y) may be specified several times.",
		PreRun: bindConf,
		RunE:   runE,
	}
)

// Execute adds all child commands to the root command and sets their flags.
func Execute() error {
	return RootCmd.Execute()
}
