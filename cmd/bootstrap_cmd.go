package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	gitApi "github.com/netlify/git-gateway/api"
	"github.com/netlify/git-gateway/conf"
	"github.com/netlify/git-gateway/identity/api"
	"github.com/netlify/git-gateway/identity/models"
	"github.com/netlify/git-gateway/identity/secrets"
	"github.com/netlify/git-gateway/identity/storage/dial"
	gitDial "github.com/netlify/git-gateway/storage/dial"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var bootstrapCmd = cobra.Command{
	Use:  "bootstrap",
	Long: "Bootstrap mode waits for installation to commence. It guides the user to setup a GitHub App.",
	Run: func(cmd *cobra.Command, args []string) {
		gcpMode()
	},
}

func gcpMode() {
	globalConfig, err := conf.LoadGlobal(configFile)
	if err != nil {
		logrus.Fatalf("Failed to load configuration: %+v", err)
	}

	gitDb, err := gitDial.Dial(globalConfig)
	if err != nil {
		logrus.Fatalf("Error opening database: %+v", err)
	}
	defer gitDb.Close()

	db, err := dial.Dial(globalConfig)
	if err != nil {
		logrus.Fatalf("Error opening database: %+v", err)
	}
	defer db.Close()

	gitApi := gitApi.NewAPIWithVersion(context.TODO(), globalConfig, gitDb, Version)
	api := api.NewAPIWithVersion(context.TODO(), globalConfig, db, gitApi)

	// Check if existing app configuration is mounted on the environment
	if data, hasData := os.LookupEnv("GITGATEWAY_GCP_SECRET_DATA"); hasData {
		app := new(models.App)
		if err = json.Unmarshal([]byte(data), &app); err == nil {
			api.GetSingleApp = func() (app *models.App, err error) {
				return app, nil
			}
			goto start
		}
	}
	// Check if there is an existing app configuration
	if secretName, useSecret := os.LookupEnv("GITGATEWAY_GCP_SECRET"); !globalConfig.MultiInstanceMode && useSecret {
		var cachedApp *models.App
		api.GetSingleApp = func() (app *models.App, err error) {
			if cachedApp != nil {
				return cachedApp, nil
			}
			cachedApp, err = secrets.GetApp(context.TODO(), secretName)
			return cachedApp, err
		}

		if cachedApp, err = api.GetSingleApp(); err != nil {
			if err == secrets.ErrNotAvailable {
				api.SetupEnabled = true
			}
			panic(err)
		}
	}

start:
	l := fmt.Sprintf("%v:%v", globalConfig.API.Host, globalConfig.API.Port)
	logrus.Infof("git-gateway API started on %s in %s", l, time.Since(startTime))
	api.ListenAndServe(l)
}

var startTime = time.Now()
