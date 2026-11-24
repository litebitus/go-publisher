package config

import (
	"reflect"
	"testing"

	"github.com/litebitus/go-publisher/internal/fielddescriptions"
	"github.com/stretchr/testify/require"
)

// FieldDescriptions holds the doc comments as they are: `go generate` in
// pkg/config writes it again after a comment changes.
func TestFieldDescriptionsAreTheDocComments(t *testing.T) {
	current, err := fielddescriptions.Extract(".", reflect.TypeOf(Project{}).PkgPath())
	require.NoError(t, err)
	require.Equal(t, current, FieldDescriptions, "field_descriptions.go is out of date: run `go generate ./pkg/config`")
}
