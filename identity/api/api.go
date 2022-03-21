package api

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	"github.com/netlify/git-gateway/conf"
	"github.com/netlify/git-gateway/identity/models"
	"github.com/netlify/git-gateway/identity/storage"
	"github.com/netlify/git-gateway/static"
	"github.com/sirupsen/logrus"
)

type API struct {
	handler http.Handler
	db      storage.Connection
	config  *conf.GlobalConfiguration
	version string

	SetupEnabled bool
	GetSingleApp func() (*models.App, error)
}

const AppSetupRedirectPath = "/identity/github/setup-redirect"
const AppOAuthCallback = "/identity/github/callback"
const AppHook = "/identity/github/events"

// ListenAndServe starts the REST API
func (a *API) ListenAndServe(hostAndPort string) {
	log := logrus.WithField("component", "identity-api")

	server := &http.Server{
		Addr:    hostAndPort,
		Handler: a.handler,
	}
	done := make(chan struct{})
	defer close(done)
	go func() {
		waitForTermination(log, done)
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		server.Shutdown(ctx)
	}()

	if err := server.ListenAndServe(); err != nil {
		log.WithError(err).Fatal("Identity API server failed")
	}
}

// waitForShutdown blocks until the system signals termination or done has a value
func waitForTermination(log logrus.FieldLogger, done <-chan struct{}) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	select {
	case sig := <-signals:
		log.Infof("Triggering shutdown from signal %s", sig)
	case <-done:
		log.Infof("Shutting down...")
	}
}

// NewAPIWithVersion creates a new REST API using the specified version
func NewAPIWithVersion(ctx context.Context, globalConfig *conf.GlobalConfiguration, db storage.Connection, version string) *API {
	api := &API{config: globalConfig, db: db, version: version}

	r := chi.NewRouter()
	// r.Use(addRequestID)
	// r.UseBypass(newStructuredLogger(logrus.StandardLogger()))
	// r.Use(recoverer)

	if api.SetupEnabled {
		r.Get(AppSetupRedirectPath, withError(api.setup))
	}
	r.Post(AppHook, withError(api.hook))
	r.Get(AppOAuthCallback, withError(api.callback))
	r.Get("/select-app", func(rw http.ResponseWriter, r *http.Request) {})

	r.Route("/", func(r chi.Router) {
		r.Get("/", withError(api.withAuthentication(api.home)))
		r.Mount("/*", http.FileServer(http.FS(static.Files)))
	})

	// TODO mount at the right place
	r.Route("/identity/", func(r chi.Router) {
		r.Post("/token", withError(api.Token))
		r.Get("/user", withError(api.withAuthentication(api.User)))
	})

	// TODO add proxy to git-gateway

	api.handler = chi.ServerBaseContext(r, ctx)
	LoadTemplates()
	return api
}
