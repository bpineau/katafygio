package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

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

func runE(cmd *cobra.Command, args []string) (err error) {
	logger := log.New(logLevel, logServer, logOutput)

	if restcfg == nil {
		restcfg, err = client.New(apiServer, kubeConf)
		if err != nil {
			return fmt.Errorf("failed to create a client: %v", err)
		}
	}

	repo, err := git.New(logger, dryRun, localDir, gitURL).Start()
	if err != nil {
		return fmt.Errorf("failed to start git repo handler: %v", err)
	}

	evts := event.New()
	fact := controller.NewFactory(logger, filter, resyncInt, exclobj)
	reco := recorder.New(logger, evts, localDir, resyncInt*2, dryRun).Start()
	obsv := observer.New(logger, restcfg, evts, fact, exclkind).Start()
	http := health.New(logger, healthP).Start()

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGTERM)
	signal.Notify(sigterm, syscall.SIGINT)
	if !dumpMode {
		<-sigterm
	}

	obsv.Stop()
	repo.Stop()
	reco.Stop()
	http.Stop()

	return nil
}

// Execute adds all child commands to the root command and sets their flags.
func Execute() error {
	return RootCmd.Execute()
}
