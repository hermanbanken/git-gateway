package object

import (
	// this is where we do the connections

	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	// import drivers we might need
	"cloud.google.com/go/firestore"
	_ "github.com/GoogleCloudPlatform/cloudsql-proxy/proxy/dialers/mysql"
	_ "github.com/GoogleCloudPlatform/cloudsql-proxy/proxy/dialers/postgres"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/netlify/git-gateway/conf"
	"github.com/netlify/git-gateway/models"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type logger struct {
	entry *logrus.Entry
}

func (l logger) Print(v ...interface{}) {
	l.entry.Print(v...)
}

// Connection represents a sql connection.
type Connection struct {
	client     *firestore.Client
	collection string
}

// Automigrate creates any missing tables and/or columns.
func (conn *Connection) Automigrate() error {
	// conn.db = conn.db.AutoMigrate(&models.Instance{})
	// return conn.db.Error
	return nil
}

// Close closes the database connection.
func (conn *Connection) Close() error {
	return conn.client.Close()
}

func (conn *Connection) ctx() context.Context {
	ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
	return ctx
}

type firestoreInstance struct {
	// Netlify UUID
	UUID          string              `firestore:"uuid,omitempty"`
	RawBaseConfig string              `firestore:"raw_config"`
	BaseConfig    *conf.Configuration `json:"config"`
}

func (instance *firestoreInstance) buildModel(snp *firestore.DocumentSnapshot) *models.Instance {
	return &models.Instance{
		ID:            snp.Ref.ID,
		UUID:          instance.UUID,
		BaseConfig:    instance.BaseConfig,
		RawBaseConfig: instance.RawBaseConfig,
		CreatedAt:     snp.CreateTime,
		UpdatedAt:     snp.UpdateTime,
		DeletedAt:     nil,
	}
}

// GetInstance finds an instance by ID
func (conn *Connection) GetInstance(instanceID string) (*models.Instance, error) {
	snp, err := conn.client.Collection(conn.collection).Doc(instanceID).Get(conn.ctx())
	instance := firestoreInstance{}
	if err == nil {
		err = snp.DataTo(&instance)
	}
	if err != nil {
		if grpc.Code(err) == codes.NotFound {
			return nil, models.InstanceNotFoundError{}
		}
		return nil, errors.Wrap(err, "error finding instance")
	}
	return instance.buildModel(snp), nil
}

func (conn *Connection) GetInstanceByUUID(uuid string) (*models.Instance, error) {
	it := conn.client.Collection(conn.collection).WherePath(firestore.FieldPath{"uuid"}, "==", uuid).Documents(conn.ctx())
	defer it.Stop()
	snp, err := it.Next()
	instance := firestoreInstance{}
	if err == nil {
		err = snp.DataTo(&instance)
	}
	if err != nil {
		if grpc.Code(err) == codes.NotFound || err == iterator.Done {
			return nil, models.InstanceNotFoundError{}
		}
		return nil, errors.Wrap(err, "error finding instance")
	}
	return instance.buildModel(snp), nil
}

func (conn *Connection) CreateInstance(instance *models.Instance) error {
	_, err := conn.client.Collection(conn.collection).Doc(instance.ID).Create(conn.ctx(), firestoreInstance{
		UUID:          instance.UUID,
		RawBaseConfig: instance.RawBaseConfig,
		BaseConfig:    instance.BaseConfig,
	})
	if err != nil {
		return errors.Wrap(err, "Error creating instance")
	}
	return nil
}

func (conn *Connection) UpdateInstance(instance *models.Instance) error {
	_, err := conn.client.Collection(conn.collection).Doc(instance.ID).Set(conn.ctx(), firestoreInstance{
		UUID:          instance.UUID,
		RawBaseConfig: instance.RawBaseConfig,
		BaseConfig:    instance.BaseConfig,
	})
	if err != nil {
		return errors.Wrap(err, "Error updating instance")
	}
	return nil
}

func (conn *Connection) DeleteInstance(instance *models.Instance) error {
	_, err := conn.client.Collection(conn.collection).Doc(instance.ID).Delete(conn.ctx())
	if err != nil {
		return errors.Wrap(err, "Error deleting instance")
	}
	return nil
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
	client, err := firestore.NewClient(ctx, project, debugOpt...)
	if err != nil {
		return nil, err
	}
	namespace := config.DB.Namespace
	if namespace == "" {
		namespace = "gitgateway"
	}
	return &Connection{
		client:     client,
		collection: namespace,
	}, nil
}

func IsFirestore(db conf.DBConfiguration) bool {
	if db.Driver == "firestore" {
		return true
	}
	return strings.HasPrefix(db.URL, "gcp://firestore")
}

var debugOpt []option.ClientOption

// TODO enable blocking connect when debugging
func init() {
	if false {
		debugOpt = []option.ClientOption{option.WithGRPCDialOption(grpc.WithBlock())}
	}
}
