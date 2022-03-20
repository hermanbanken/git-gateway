package dial

import (
	"errors"

	"github.com/netlify/git-gateway/conf"
	"github.com/netlify/git-gateway/identity/storage"
	"github.com/netlify/git-gateway/identity/storage/object"
)

// Dial will connect to that storage engine
func Dial(config *conf.GlobalConfiguration) (storage.Connection, error) {
	var conn storage.Connection
	var err error
	if object.IsFirestore(config.DB) {
		conn, err = object.Dial(config)
	} else {
		return nil, errors.New("only firestore is supported at this time")
	}

	if err != nil {
		return nil, err
	}

	return conn, nil
}
