package secrets

import (
	"context"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAcceptsVersioned(t *testing.T) {
	var name string
	var err error
	_, _, err = fullNameOfSecretVersion("foobar", "latest")
	assert.Error(t, err)
	name, _, err = fullNameOfSecretVersion("projects/foo-bar/secrets/baz", "latest")
	assert.NoError(t, err)
	assert.Equal(t, "projects/foo-bar/secrets/baz/versions/latest", name)
	name, _, err = fullNameOfSecretVersion("projects/foo-bar/secrets/baz/versions/10", "latest")
	assert.NoError(t, err)
	assert.Equal(t, "projects/foo-bar/secrets/baz/versions/10", name)
}

func ExampleGetApp() {
	app, err := GetApp(context.Background(), "projects/yourproject/secrets/yoursecret")
	if err != nil {
		log.Println(err)
	} else {
		log.Println("Got", app.Name)
	}
}
