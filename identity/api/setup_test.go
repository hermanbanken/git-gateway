package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTemplates(t *testing.T) {
	LoadTemplates()
	assert.Contains(t, templates, "setup.html")
}
