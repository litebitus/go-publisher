package testlib

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEmbeddedSchemaIDsNamesEveryNestedIDButTheRoot(t *testing.T) {
	var document any
	require.NoError(t, json.Unmarshal([]byte(`{
		"$id": "https://mock.example/root.json",
		"$defs": {
			"plain": {"type": "string"},
			"embedded": {"$id": "https://mock.example/embedded.json"}
		},
		"oneOf": [{"type": "string"}, {"$id": "https://mock.example/in-list.json"}]
	}`), &document))
	require.Equal(t, []string{"/$defs/embedded", "/oneOf/1"}, EmbeddedSchemaIDs(document))
}
