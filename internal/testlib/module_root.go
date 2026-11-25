package testlib

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// ModuleRoot is the directory holding the go.mod of the package under test,
// found from the directory the test runs in.
func ModuleRoot(tb testing.TB) string {
	tb.Helper()
	dir, err := os.Getwd()
	require.NoError(tb, err)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		require.NotEqual(tb, dir, parent, "no go.mod above the test's directory")
		dir = parent
	}
}
