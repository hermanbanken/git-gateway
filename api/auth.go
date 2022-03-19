package api

import (
	"context"
	"errors"
	"net/http"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/netlify/git-gateway/models"
	"github.com/sirupsen/logrus"
)

// requireAuthentication checks incoming requests for tokens presented using the Authorization header
func (a *API) requireAuthentication(w http.ResponseWriter, r *http.Request) (context.Context, error) {
	logrus.Info("Getting auth token")
	token, err := a.extractBearerToken(w, r)
	if err != nil {
		return nil, err
	}

	logrus.Infof("Parsing JWT claims: %v", token)
	return a.parseJWTClaims(token, r)
}

func (a *API) extractBearerToken(w http.ResponseWriter, r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", unauthorizedError("This endpoint requires a Bearer token")
	}

	matches := bearerRegexp.FindStringSubmatch(authHeader)
	if len(matches) != 2 {
		return "", unauthorizedError("This endpoint requires a Bearer token")
	}

	return matches[1], nil
}

func (a *API) parseJWTClaims(bearer string, r *http.Request) (context.Context, error) {
	ctx := r.Context()
	p := jwt.Parser{ValidMethods: []string{jwt.SigningMethodHS256.Name}}
	claims := GatewayClaims{}
	token, err := p.ParseWithClaims(bearer, &claims, func(token *jwt.Token) (key interface{}, err error) {
		if !a.config.MultiInstanceMode {
			config := getConfig(r.Context())
			secret := config.JWT.Secret
			return []byte(secret), nil
		}

		instance, err := a.db.GetInstanceByUUID(claims.Subject)
		if err != nil {
			return nil, err
		}
		conf, err := instance.Config()
		if err != nil {
			if models.IsNotFoundError(err) {
				return nil, notFoundError("Unable to locate site configuration")
			}
			return nil, err
		}
		ctx, err = WithInstanceConfig(r.Context(), conf, instance.ID)
		return []byte(conf.JWT.Secret), err
	})
	if err != nil {
		if errors.Is(err, &HTTPError{}) {
			return nil, err
		}
		return nil, unauthorizedError("Invalid token: %v", err)
	}
	return withToken(ctx, token), err
}
