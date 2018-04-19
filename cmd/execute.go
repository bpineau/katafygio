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
	version = "0.3.0"

	cfgFile   string
	apiServer string
	kubeConf  string
	dryRun    bool
	dumpMode  bool
	logLevel  string
	logOutput string
	logServer string
	filter    string
	localDir  string
	gitURL    string
	healthP   int
	resync    int
	exclkind  []string
	exclobj   []string

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
		Long: "Backup Kubernetes cluster as yaml files in a git repository.\n" +
			"--exclude-kind (x) and --exclude-object (-y) may be specified several times.",

		RunE: func(cmd *cobra.Command, args []string) error {
			conf := &config.KfConfig{
				DryRun:        viper.GetBool("dry-run"),
				DumpMode:      viper.GetBool("dump-only"),
				Logger:        klog.New(viper.GetString("log.level"), viper.GetString("log.server"), viper.GetString("log.output")),
				LocalDir:      viper.GetString("local-dir"),
				GitURL:        viper.GetString("git-url"),
				Filter:        viper.GetString("filter"),
				ExcludeKind:   viper.GetStringSlice("exclude-kind"),
				ExcludeObject: viper.GetStringSlice("exclude-object"),
				HealthPort:    viper.GetInt("healthcheck-port"),
				ResyncIntv:    time.Duration(viper.GetInt("resync-interval")) * time.Second,
			}

			err := conf.Init(viper.GetString("api-server"), viper.GetString("kube-config"))
			if err != nil {
				return fmt.Errorf("Failed to initialize the configuration: %v", err)
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
	RootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", defaultCfg, "Configuration file")

	RootCmd.PersistentFlags().StringVarP(&apiServer, "api-server", "s", "", "Kubernetes api-server url")
	bindPFlag("api-server", "api-server")

	RootCmd.PersistentFlags().StringVarP(&kubeConf, "kube-config", "k", "", "Kubernetes config path")
	bindPFlag("kube-config", "kube-config")
	if err := viper.BindEnv("kube-config", "KUBECONFIG"); err != nil {
		log.Fatal("Failed to bind cli argument:", err)
	}

	RootCmd.PersistentFlags().BoolVarP(&dryRun, "dry-run", "d", false, "Dry-run mode: don't store anything")
	bindPFlag("dry-run", "dry-run")

	RootCmd.PersistentFlags().BoolVarP(&dumpMode, "dump-only", "m", false, "Dump mode: dump everything once and exit")
	bindPFlag("dump-only", "dump-only")

	RootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "v", "info", "Log level")
	bindPFlag("log.level", "log-level")

	RootCmd.PersistentFlags().StringVarP(&logOutput, "log-output", "o", "stderr", "Log output")
	bindPFlag("log.output", "log-output")

	RootCmd.PersistentFlags().StringVarP(&logServer, "log-server", "r", "", "Log server (if using syslog)")
	bindPFlag("log.server", "log-server")

	RootCmd.PersistentFlags().StringVarP(&localDir, "local-dir", "e", "./kubernetes-backup", "Where to dump yaml files")
	bindPFlag("local-dir", "local-dir")

	RootCmd.PersistentFlags().StringVarP(&gitURL, "git-url", "g", "", "Git repository URL")
	bindPFlag("git-url", "git-url")

	RootCmd.PersistentFlags().StringSliceVarP(&exclkind, "exclude-kind", "x", nil, "Ressource kind to exclude. Eg. 'deployment'")
	bindPFlag("exclude-kind", "exclude-kind")

	RootCmd.PersistentFlags().StringSliceVarP(&exclobj, "exclude-object", "y", nil, "Object to exclude. Eg. 'configmap:kube-system/kube-dns'")
	bindPFlag("exclude-object", "exclude-object")

	RootCmd.PersistentFlags().StringVarP(&filter, "filter", "l", "", "Label filter. Select only objects matching the label.")
	bindPFlag("filter", "filter")

	RootCmd.PersistentFlags().IntVarP(&healthP, "healthcheck-port", "p", 0, "Port for answering healthchecks on /health url")
	bindPFlag("healthcheck-port", "healthcheck-port")

	RootCmd.PersistentFlags().IntVarP(&resync, "resync-interval", "i", 300, "Resync interval in seconds (0 to disable)")
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
	viper.SetEnvPrefix("KF")
	replacer := strings.NewReplacer("-", "_", ".", "_DOT_")
	viper.SetEnvKeyReplacer(replacer)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		RootCmd.Printf("Using config file: %s", viper.ConfigFileUsed())
	}
}
