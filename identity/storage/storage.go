package storage

import "github.com/netlify/git-gateway/identity/models"

// Connection is the interface a storage provider must implement.
type Connection interface {
	Close() error

	ListApps(count int, start string) ([]*models.App, error)
	CreateApp(app *models.App) error
	GetApp(appID int64) (*models.App, error)

	CreateInstallation(app *models.Installation) error
	UpdateInstallation(app *models.Installation) error
	GetInstallation(id int64) (*models.Installation, error)
	DeleteInstallation(id int64) error
}
