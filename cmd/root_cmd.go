package cmd

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/netlify/git-gateway/conf"
)

var configFile = ""

var rootCmd = cobra.Command{
	Use: "git-gateway",
	Run: func(cmd *cobra.Command, args []string) {
		if _, allowBootstrap := os.LookupEnv("BOOTSTRAP_IDENTITY"); allowBootstrap {
			bootstrapIdentity()
		}
		execWithConfig(cmd, serve)
	},
}

// RootCommand will setup and return the root command
func RootCommand() *cobra.Command {
	rootCmd.AddCommand(&serveCmd, &migrateCmd, &multiCmd, &bootstrapCmd, &versionCmd)
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "the config file to use")

	return &rootCmd
}

func execWithConfig(cmd *cobra.Command, fn func(globalConfig *conf.GlobalConfiguration, config *conf.Configuration)) {
	globalConfig, err := conf.LoadGlobal(configFile)
	if err != nil {
		logrus.Fatalf("Failed to load configuration: %+v", err)
	}
	config, err := conf.LoadConfig(configFile)
	if err != nil {
		logrus.Fatalf("Failed to load configuration: %+v", err)
	}

	fn(globalConfig, config)
}

func execWithConfigAndArgs(cmd *cobra.Command, fn func(globalConfig *conf.GlobalConfiguration, config *conf.Configuration, args []string), args []string) {
	globalConfig, err := conf.LoadGlobal(configFile)
	if err != nil {
		logrus.Fatalf("Failed to load configuration: %+v", err)
	}
	config, err := conf.LoadConfig(configFile)
	if err != nil {
		logrus.Fatalf("Failed to load configuration: %+v", err)
	}

	fn(globalConfig, config, args)
}
