package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/spf13/afero"
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
	appFs   = afero.NewOsFs()

	// RootCmd is our main entry point, launching runE()
	RootCmd = &cobra.Command{
		Use:   appName,
		Short: "Backup Kubernetes cluster as yaml files",
		Long: "Backup Kubernetes cluster as yaml files in a git repository.\n" +
			"--exclude-kind (-x) and --exclude-object (-y) may be specified several times,\n" +
			"or once with several comma separated values.",
		SilenceUsage:  true,
		SilenceErrors: true,
		PreRun:        bindConf,
		RunE:          runE,
	}
)

func runE(cmd *cobra.Command, args []string) (err error) {
	logger, err := log.New(logLevel, logServer, logOutput)
	if err != nil {
		return fmt.Errorf("failed to create a logger: %v", err)
	}
	logger.Info(appName, " starting")

	if restcfg == nil {
		restcfg, err = client.New(apiServer, context, kubeConf)
		if err != nil {
			return fmt.Errorf("failed to create a client: %v", err)
		}
	}

	err = appFs.MkdirAll(filepath.Clean(localDir), 0700)
	if err != nil {
		return fmt.Errorf("can't create directory %s: %v", localDir, err)
	}

	http := health.New(logger, healthP).Start()

	var repo *git.Store
	if !noGit {
		repo, err = git.New(logger, dryRun, localDir, gitURL, gitAuthor, gitEmail, gitTimeout, time.Duration(checkInt)*time.Second).Start()
	}
	if err != nil {
		return fmt.Errorf("failed to start git repo handler: %v", err)
	}

	evts := event.New()
	fact := controller.NewFactory(logger, filter, resyncInt, exclobj)
	reco := recorder.New(logger, evts, localDir, resyncInt*2, dryRun).Start()
	obsv := observer.New(logger, restcfg, evts, fact, exclkind, namespace).Start()

	logger.Info(appName, " started")
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGTERM)
	signal.Notify(sigterm, syscall.SIGINT)
	if !dumpMode {
		<-sigterm
	}

	logger.Info(appName, " stopping")
	obsv.Stop()
	reco.Stop()
	http.Stop()
	if !noGit {
		repo.Stop()
	}
	logger.Info(appName, " stopped")

	return nil
}

// Execute adds all child commands to the root command and sets their flags.
func Execute() error {
	return RootCmd.Execute()
}
