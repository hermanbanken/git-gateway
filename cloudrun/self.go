package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"strings"

	"cloud.google.com/go/compute/metadata"
	"golang.org/x/oauth2/google"
)

var ErrNotCloudRun = errors.New("Not running on CloudRun")

// https://medium.com/google-cloud/3-solutions-to-mitigate-the-cold-starts-on-cloud-run-8c60f0ae7894
func getCloudRunApiUrl() (*url.URL, error) {
	if service, has := os.LookupEnv("K_SERVICE"); has {
		projectNr, region, err := getProjectAndRegion()
		if err != nil {
			return nil, err
		}
		uri, err := getCloudRunUrl(region, projectNr, service)
		if err != nil {
			return nil, err
		}
		return url.Parse(uri)
	}
	return nil, ErrNotCloudRun
}

func getProjectAndRegion() (prNb string, region string, err error) {
	resp, err := metadata.Get("/instance/region")
	if err != nil {
		return
	}
	// response pattern is projects/<projectNumber>/regions/<region>
	r := strings.Split(resp, "/")
	prNb = r[1]
	region = r[3]
	return
}

func getCloudRunUrl(region string, projectNumber string, service string) (url string, err error) {
	ctx := context.Background()
	client, err := google.DefaultClient(ctx)

	cloudRunApi := fmt.Sprintf("https://%s-run.googleapis.com/apis/serving.knative.dev/v1/namespaces/%s/services/%s", region, projectNumber, service)
	resp, err := client.Get(cloudRunApi)

	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err

	}
	cloudRunResp := &CloudRunAPIUrlOnly{}
	json.Unmarshal(body, cloudRunResp)
	url = cloudRunResp.Status.URL
	return
}

// Minimal type to get only the interesting part in the answer
type CloudRunAPIUrlOnly struct {
	Status struct {
		URL string `json:"url"`
	} `json:"status"`
}
