package cmd

import (
	"log"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile    string
	apiServer  string
	context    string
	namespace  string
	kubeConf   string
	dryRun     bool
	dumpMode   bool
	logLevel   string
	logOutput  string
	logServer  string
	filter     string
	localDir   string
	gitURL     string
	gitTimeout time.Duration
	healthP    int
	resyncInt  int
	checkInt   int
	exclkind   []string
	exclobj    []string
	noGit      bool
	gitAuthor  string
	gitEmail   string
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

	RootCmd.PersistentFlags().StringVarP(&context, "context", "q", "", "Kubernetes configuration context")
	bindPFlag("context", "context")

	RootCmd.PersistentFlags().StringVarP(&namespace, "namespace", "a", "", "Only dump objects from this namespace")
	bindPFlag("namespace", "namespace")

	RootCmd.PersistentFlags().StringVarP(&kubeConf, "kube-config", "k", "", "Kubernetes configuration path")
	bindPFlag("kube-config", "kube-config")

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

	RootCmd.PersistentFlags().DurationVarP(&gitTimeout, "git-timeout", "t", 300*time.Second, "Git operations timeout")
	bindPFlag("git-timeout", "git-timeout")

	RootCmd.PersistentFlags().StringSliceVarP(&exclkind, "exclude-kind", "x", nil, "Ressource kind to exclude. Eg. 'deployment'")
	bindPFlag("exclude-kind", "exclude-kind")

	RootCmd.PersistentFlags().StringSliceVarP(&exclobj, "exclude-object", "y", nil, "Object to exclude. Eg. 'configmap:kube-system/kube-dns'")
	bindPFlag("exclude-object", "exclude-object")

	RootCmd.PersistentFlags().StringVarP(&filter, "filter", "l", "", "Label filter. Select only objects matching the label")
	bindPFlag("filter", "filter")

	RootCmd.PersistentFlags().IntVarP(&healthP, "healthcheck-port", "p", 0, "Port for answering healthchecks on /health url")
	bindPFlag("healthcheck-port", "healthcheck-port")

	RootCmd.PersistentFlags().IntVarP(&resyncInt, "resync-interval", "i", 900, "Full resync interval in seconds (0 to disable)")
	bindPFlag("resync-interval", "resync-interval")

	RootCmd.PersistentFlags().IntVarP(&checkInt, "check-interval", "j", 10, "Check interval in seconds")
	bindPFlag("check-interval", "check-interval")

	RootCmd.PersistentFlags().BoolVarP(&noGit, "no-git", "n", false, "Don't version with git")
	bindPFlag("no-git", "no-git")

	RootCmd.PersistentFlags().StringVarP(&gitAuthor, "git-author", "b", "Katafygio", "Author for git commits")
	bindPFlag("git-author", "git-author")

	RootCmd.PersistentFlags().StringVarP(&gitEmail, "git-email", "f", "katafygio@localhost", "Email address for git commits")
	bindPFlag("git-email", "git-email")
}

// for whatever the reason, viper don't auto bind values from config file so we have to tell him
func bindConf(cmd *cobra.Command, args []string) {
	apiServer = viper.GetString("api-server")
	context = viper.GetString("context")
	namespace = viper.GetString("namespace")
	kubeConf = viper.GetString("kube-config")
	dryRun = viper.GetBool("dry-run")
	dumpMode = viper.GetBool("dump-only")
	logLevel = viper.GetString("log-level")
	logOutput = viper.GetString("log-output")
	logServer = viper.GetString("log-server")
	filter = viper.GetString("filter")
	localDir = viper.GetString("local-dir")
	gitURL = viper.GetString("git-url")
	gitTimeout = viper.GetDuration("git-timeout")
	healthP = viper.GetInt("healthcheck-port")
	resyncInt = viper.GetInt("resync-interval")
	checkInt = viper.GetInt("check-interval")
	exclkind = viper.GetStringSlice("exclude-kind")
	exclobj = viper.GetStringSlice("exclude-object")
	noGit = viper.GetBool("no-git")
	gitAuthor = viper.GetString("git-author")
	gitEmail = viper.GetString("git-email")
}
