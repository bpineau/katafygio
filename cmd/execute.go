package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/bpineau/katafygio/config"
	"github.com/bpineau/katafygio/pkg/client"
	"github.com/bpineau/katafygio/pkg/log"
	"github.com/bpineau/katafygio/pkg/run"
)

const appName = "katafygio"

var (
	restcfg client.Interface

	// RootCmd is our main entry point, launching pkg/run.Run()
	RootCmd = &cobra.Command{
		Use:   appName,
		Short: "Backup Kubernetes cluster as yaml files",
		Long: "Backup Kubernetes cluster as yaml files in a git repository.\n" +
			"--exclude-kind (-x) and --exclude-object (-y) may be specified several times.",

		RunE: func(cmd *cobra.Command, args []string) (err error) {
			resync := time.Duration(viper.GetInt("resync-interval")) * time.Second
			logger := log.New(viper.GetString("log.level"),
				viper.GetString("log.server"),
				viper.GetString("log.output"))

			if restcfg == nil {
				restcfg, err = client.New(viper.GetString("api-server"),
					viper.GetString("kube-config"))
				if err != nil {
					return fmt.Errorf("failed to create a client: %v", err)
				}
			}

			conf := &config.KfConfig{
				DryRun:        viper.GetBool("dry-run"),
				DumpMode:      viper.GetBool("dump-only"),
				Logger:        logger,
				LocalDir:      viper.GetString("local-dir"),
				GitURL:        viper.GetString("git-url"),
				Filter:        viper.GetString("filter"),
				ExcludeKind:   viper.GetStringSlice("exclude-kind"),
				ExcludeObject: viper.GetStringSlice("exclude-object"),
				HealthPort:    viper.GetInt("healthcheck-port"),
				Client:        restcfg,
				ResyncIntv:    resync,
			}

			run.Run(conf) // <- this is where things happen
			return nil
		},
	}
)

// Execute adds all child commands to the root command and sets their flags.
func Execute() error {
	return RootCmd.Execute()
}
