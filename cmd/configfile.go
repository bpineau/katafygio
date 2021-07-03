package cmd

import (
	"log"
	"os"
	"strings"

	"github.com/spf13/viper"
	"k8s.io/client-go/util/homedir"
)

func loadConfigFile() {
	viper.SetConfigType("yaml")
	viper.SetConfigName(appName + ".yaml")

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

	if err := viper.ReadInConfig(); err != nil {
		_, ok := err.(viper.ConfigFileNotFoundError)
		if !ok {
			log.Fatalf("Can't read config file: %v\n", err)
		}
	}
}
