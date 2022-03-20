package object

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/netlify/git-gateway/conf"
	"github.com/netlify/git-gateway/identity/models"
	"github.com/pkg/errors"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Connection is the interface a storage provider must implement.
type Connection struct {
	client *firestore.Client

	appCollection          string
	installationCollection string
}

func (conn *Connection) ctx() context.Context {
	ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
	return ctx
}

func (conn *Connection) Close() error {
	return conn.client.Close()
}

func (conn *Connection) CreateApp(app *models.App) error {
	docId := strconv.FormatInt(app.ID, 10)
	_, err := conn.client.Collection(conn.appCollection).Doc(docId).Create(conn.ctx(), app)
	if err != nil {
		return errors.Wrap(err, "Error creating app")
	}
	return nil
}

func (conn *Connection) GetApp(appID int64) (*models.App, error) {
	docId := strconv.FormatInt(appID, 10)
	snp, err := conn.client.Collection(conn.appCollection).Doc(docId).Get(conn.ctx())
	model := models.App{}
	if err == nil {
		err = snp.DataTo(&model)
	}
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, models.NotFoundError{Table: conn.appCollection, ID: docId}
		}
		return nil, errors.Wrap(err, "error finding installation")
	}
	return &model, nil
}

func (conn *Connection) CreateInstallation(install *models.Installation) error {
	docId := strconv.FormatInt(install.InstallationID, 10)
	_, err := conn.client.Collection(conn.installationCollection).Doc(docId).Create(conn.ctx(), install)
	if err != nil {
		return errors.Wrap(err, "Error creating installation")
	}
	return nil

}
func (conn *Connection) UpdateInstallation(install *models.Installation) error {
	docId := strconv.FormatInt(install.InstallationID, 10)
	_, err := conn.client.Collection(conn.installationCollection).Doc(docId).Set(conn.ctx(), install)
	if err != nil {
		return errors.Wrap(err, "Error updating installation")
	}
	return nil

}

func (conn *Connection) GetInstallation(id int64) (*models.Installation, error) {
	docId := strconv.FormatInt(id, 10)
	snp, err := conn.client.Collection(conn.installationCollection).Doc(docId).Get(conn.ctx())
	model := models.Installation{}
	if err == nil {
		err = snp.DataTo(&model)
	}
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, models.NotFoundError{Table: conn.installationCollection, ID: docId}
		}
		return nil, errors.Wrap(err, "error finding installation")
	}
	return &model, nil
}
func (conn *Connection) DeleteInstallation(id int64) error {
	docId := strconv.FormatInt(id, 10)
	_, err := conn.client.Collection(conn.installationCollection).Doc(docId).Delete(conn.ctx())
	if err != nil && status.Code(err) == codes.NotFound {
		return models.NotFoundError{Table: conn.installationCollection, ID: docId}
	}
	return errors.Wrap(err, "error deleting installation")
}

// Dial will connect to that storage engine
func Dial(config *conf.GlobalConfiguration) (*Connection, error) {
	ctx := context.Background()

	var uri *url.URL
	var err error
	project := firestore.DetectProjectID
	if config.DB.URL != "" {
		uri, err = url.Parse(config.DB.URL)
		if err != nil {
			return nil, fmt.Errorf("invalid db url (format gcp://firestore?project=projectid): %w", err)
		}
		if proj := uri.Query().Get("project"); proj != "" {
			project = proj
		}
	}

	// make connection (WithBlock to ensure it works)
	client, err := firestore.NewClient(ctx, project, option.WithGRPCDialOption(grpc.WithBlock()))
	if err != nil {
		return nil, err
	}
	namespace := config.DB.Namespace
	if namespace == "" {
		namespace = "gitgateway"
	}
	return &Connection{
		client:                 client,
		appCollection:          namespace + "Apps",
		installationCollection: namespace + "Installations",
	}, nil
}

func IsFirestore(db conf.DBConfiguration) bool {
	if db.Driver == "firestore" {
		return true
	}
	return strings.HasPrefix(db.URL, "gcp://firestore")
}
