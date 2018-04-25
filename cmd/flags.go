package cmd

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
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
	resyncInt int
	exclkind  []string
	exclobj   []string
)

func bindPFlag(key string, cmd string) {
	if err := viper.BindPFlag(key, RootCmd.PersistentFlags().Lookup(cmd)); err != nil {
		log.Fatal("Failed to bind cli argument:", err)
	}
}

func init() {
	cobra.OnInitialize(loadConfigFile)
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
	bindPFlag("log-level", "log-level")

	RootCmd.PersistentFlags().StringVarP(&logOutput, "log-output", "o", "stderr", "Log output")
	bindPFlag("log-output", "log-output")

	RootCmd.PersistentFlags().StringVarP(&logServer, "log-server", "r", "", "Log server (if using syslog)")
	bindPFlag("log-server", "log-server")

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

	RootCmd.PersistentFlags().IntVarP(&resyncInt, "resync-interval", "i", 900, "Full resync interval in seconds (0 to disable)")
	bindPFlag("resync-interval", "resync-interval")
}

// for whatever the reason, viper don't auto bind values from config file so we have to tell him
func bindConf(cmd *cobra.Command, args []string) {
	apiServer = viper.GetString("api-server")
	kubeConf = viper.GetString("kube-config")
	dryRun = viper.GetBool("dry-run")
	dumpMode = viper.GetBool("dump-only")
	logLevel = viper.GetString("log-level")
	logOutput = viper.GetString("log-output")
	logServer = viper.GetString("log-server")
	filter = viper.GetString("filter")
	localDir = viper.GetString("local-dir")
	gitURL = viper.GetString("git-url")
	healthP = viper.GetInt("healthcheck-port")
	resyncInt = viper.GetInt("resync-interval")
	exclkind = viper.GetStringSlice("exclude-kind")
	exclobj = viper.GetStringSlice("exclude-object")
}
