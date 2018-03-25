package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/client-go/util/homedir"

	"github.com/bpineau/katafygio/config"
	klog "github.com/bpineau/katafygio/pkg/log"
	"github.com/bpineau/katafygio/pkg/run"
)

const appName = "katafygio"

var (
	version = "0.1.0 (HEAD)"

	cfgFile   string
	apiServer string
	kubeConf  string
	dryRun    bool
	logLevel  string
	logOutput string
	logServer string
	filter    string
	localDir  string
	healthP   int
	resync    int

	// FakeCS uses the client-go testing clientset
	FakeCS bool

	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print the version number",
		Run: func(cmd *cobra.Command, args []string) {
			RootCmd.Printf("%s version %s\n", appName, version)
		},
	}

	// RootCmd represents the base command when called without any subcommands
	RootCmd = &cobra.Command{
		Use:   appName,
		Short: "Backup Kubernetes cluster as yaml files",
		Long:  "Backup Kubernetes cluster as yaml files",

		RunE: func(cmd *cobra.Command, args []string) error {
			conf := &config.KdnConfig{
				DryRun:     viper.GetBool("dry-run"),
				Logger:     klog.New(viper.GetString("log.level"), viper.GetString("log.server"), viper.GetString("log.output")),
				LocalDir:   viper.GetString("local-dir"),
				Filter:     viper.GetString("filter"),
				HealthPort: viper.GetInt("healthcheck-port"),
				ResyncIntv: time.Duration(viper.GetInt("resync-interval")) * time.Second,
			}
			if FakeCS {
				conf.ClientSet = config.FakeClientSet()
			}
			err := conf.Init(viper.GetString("api-server"), viper.GetString("kube-config"))
			if err != nil {
				return fmt.Errorf("Failed to initialize the configuration: %+v", err)
			}
			run.Run(conf)
			return nil
		},
	}
)

// Execute adds all child commands to the root command and sets their flags.
func Execute() error {
	return RootCmd.Execute()
}

func bindPFlag(key string, cmd string) {
	if err := viper.BindPFlag(key, RootCmd.PersistentFlags().Lookup(cmd)); err != nil {
		log.Fatal("Failed to bind cli argument:", err)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	RootCmd.AddCommand(versionCmd)

	defaultCfg := "/etc/katafygio/" + appName + ".yaml"
	RootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", defaultCfg, "configuration file")

	RootCmd.PersistentFlags().StringVarP(&apiServer, "api-server", "s", "", "kube api server url")
	bindPFlag("api-server", "api-server")

	RootCmd.PersistentFlags().StringVarP(&kubeConf, "kube-config", "k", "", "kube config path")
	bindPFlag("kube-config", "kube-config")
	if err := viper.BindEnv("kube-config", "KUBECONFIG"); err != nil {
		log.Fatal("Failed to bind cli argument:", err)
	}

	RootCmd.PersistentFlags().BoolVarP(&dryRun, "dry-run", "d", false, "dry-run mode")
	bindPFlag("dry-run", "dry-run")

	RootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "v", "info", "log level")
	bindPFlag("log.level", "log-level")

	RootCmd.PersistentFlags().StringVarP(&logOutput, "log-output", "o", "stderr", "log output")
	bindPFlag("log.output", "log-output")

	RootCmd.PersistentFlags().StringVarP(&logServer, "log-server", "r", "", "log server (if using syslog)")
	bindPFlag("log.server", "log-server")

	RootCmd.PersistentFlags().StringVarP(&localDir, "local-dir", "e", "./kubernetes-backup", "Local directory")
	bindPFlag("local-dir", "local-dir")

	RootCmd.PersistentFlags().StringVarP(&filter, "filter", "l", "", "Label filter")
	bindPFlag("filter", "filter")

	RootCmd.PersistentFlags().IntVarP(&healthP, "healthcheck-port", "p", 0, "port for answering healthchecks")
	bindPFlag("healthcheck-port", "healthcheck-port")

	RootCmd.PersistentFlags().IntVarP(&resync, "resync-interval", "i", 900, "resync interval in seconds (0 to disable)")
	bindPFlag("resync-interval", "resync-interval")
}

func initConfig() {
	viper.SetConfigType("yaml")
	viper.SetConfigName(appName)

	// all possible config file paths, by priority
	viper.AddConfigPath("/etc/katafygio/")
	if home := homedir.HomeDir(); home != "" {
		viper.AddConfigPath(home)
	}
	viper.AddConfigPath(".")

	// prefer the config file path provided by cli flag, if any
	if _, err := os.Stat(cfgFile); !os.IsNotExist(err) {
		viper.SetConfigFile(cfgFile)
	}

	// allow config params through prefixed env variables
	viper.SetEnvPrefix("KB")
	replacer := strings.NewReplacer("-", "_", ".", "_DOT_")
	viper.SetEnvKeyReplacer(replacer)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		RootCmd.Printf("Using config file: %s", viper.ConfigFileUsed())
	}
}
