package main

import (
	"log"
	"os"

	"github.com/netlify/git-gateway/cmd"
)

func main() {
	// Detect the Cloud Run url from the environment
	if _, hasValue := os.LookupEnv("GITGATEWAY_API_ENDPOINT"); !hasValue {
		url, _ := getCloudRunApiUrl()
		if url != nil {
			os.Setenv("GITGATEWAY_API_ENDPOINT", url.Hostname())
		}
	}
	// Use firestore by default
	_, hasDbDriver := os.LookupEnv("GITGATEWAY_DB_DRIVER")
	_, hasDbUrl := os.LookupEnv("DATABASE_URL")
	if !hasDbDriver && !hasDbUrl {
		os.Setenv("GITGATEWAY_DB_DRIVER", "firestore")
	}

	if err := cmd.RootCommand().Execute(); err != nil {
		log.Fatal(err)
	}
}
