package testlib

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestModuleRootHoldsGoMod(t *testing.T) {
	root := ModuleRoot(t)
	require.FileExists(t, filepath.Join(root, "go.mod"))
	require.DirExists(t, filepath.Join(root, "internal", "testlib"))
}
