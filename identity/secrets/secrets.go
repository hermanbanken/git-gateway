package secrets

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"cloud.google.com/go/compute/metadata"
	secretmanager "cloud.google.com/go/secretmanager/apiv1beta1"
	"github.com/netlify/git-gateway/identity/models"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1beta1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var ErrNotAvailable = errors.New("secret does not exist or version is not available")
var ErrNotFilled = errors.New("secret not filled")

func GetApp(ctx context.Context, secretName string) (out *models.App, err error) {
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return nil, err
	}

	versionName, _, err := fullNameOfSecretVersion(secretName, "latest")
	if err != nil {
		return nil, err
	}
	data, err := client.AccessSecretVersion(ctx, &secretmanagerpb.AccessSecretVersionRequest{Name: versionName})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, ErrNotAvailable
		}
		return nil, err
	}
	if len(data.GetPayload().GetData()) == 0 {
		return nil, ErrNotFilled
	}
	out = new(models.App)
	err = json.Unmarshal(data.GetPayload().GetData(), &out)
	return
}

func SetApp(ctx context.Context, secretName string, out *models.App) (err error) {
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return err
	}

	_, secretNameParts, err := fullNameOfSecretVersion(secretName, "latest")
	if err != nil {
		return err
	}

	_, err = client.GetSecret(ctx, &secretmanagerpb.GetSecretRequest{
		Name: secretName,
	})
	if status.Code(err) == codes.NotFound {
		_, err = client.CreateSecret(ctx, &secretmanagerpb.CreateSecretRequest{
			Parent:   strings.Join(secretNameParts[0:2], "/"),
			SecretId: secretNameParts[3],
			Secret:   &secretmanagerpb.Secret{Replication: &secretmanagerpb.Replication{Replication: &secretmanagerpb.Replication_Automatic_{Automatic: &secretmanagerpb.Replication_Automatic{}}}},
		})
	}
	if err != nil {
		return err
	}
	data, err := json.Marshal(out)
	_, err = client.AddSecretVersion(ctx, &secretmanagerpb.AddSecretVersionRequest{
		Parent: strings.Join(secretNameParts[0:4], "/"),
		Payload: &secretmanagerpb.SecretPayload{
			Data: data,
		},
	})
	return err
}

// Same value as firebase.DetectProjectID
const DetectProjectID = "*detect-project-id*"

func fullNameOfSecretVersion(inp string, defaultVersion string) (out string, parts []string, err error) {
	parts = strings.Split(inp, "/")

	// Add project if not set
	if metadata.OnGCE() {
		if parts[0] != "projects" {
			if project, err := metadata.ProjectID(); err == nil {
				if !(len(parts) == 2 && parts[0] == "secrets") {
					parts = append([]string{"secrets"}, parts...)
				}
				parts = append([]string{"projects", project}, parts...)
			}
		} else if len(parts) > 2 && parts[1] == DetectProjectID {
			if parts[1], err = metadata.ProjectID(); err != nil {
				return out, nil, err
			}
		}
	}

	// Check validity
	valid := (len(parts) == 4 || len(parts) == 6 && parts[4] == "versions") &&
		parts[0] == "projects" && parts[2] == "secrets"
	if !valid {
		return "", nil, fmt.Errorf("use the format 'projects/*/secrets/*' for the name of the secret")
	}

	// Set version if needed
	if len(parts) == 4 {
		parts = append(parts, "versions", defaultVersion)
	}
	return strings.Join(parts, "/"), parts, nil
}
