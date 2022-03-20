package cmd

import (
	"context"
	"fmt"

	"github.com/netlify/git-gateway/conf"
	"github.com/netlify/git-gateway/identity/api"
	"github.com/netlify/git-gateway/identity/storage/dial"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var bootstrapCmd = cobra.Command{
	Use:  "bootstrap",
	Long: "Bootstrap mode waits for installation to commence. It guides the user to setup a GitHub App.",
	Run: func(cmd *cobra.Command, args []string) {
		bootstrapIdentity()
	},
}

func bootstrapIdentity() {
	globalConfig, err := conf.LoadGlobal(configFile)
	if err != nil {
		logrus.Fatalf("Failed to load configuration: %+v", err)
	}

	db, err := dial.Dial(globalConfig)
	if err != nil {
		logrus.Fatalf("Error opening database: %+v", err)
	}
	defer db.Close()

	api := api.NewAPIWithVersion(context.TODO(), globalConfig, db, Version)
	l := fmt.Sprintf("%v:%v", globalConfig.API.Host, globalConfig.API.Port)
	logrus.Infof("git-gateway API started on: %s", l)
	api.ListenAndServe(l)
}
