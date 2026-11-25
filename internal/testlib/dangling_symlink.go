package testlib

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// DanglingSymlink is the path of a symlink, in a directory of the test's own,
// to a file that does not exist. It is made by the test rather than committed,
// since a committed one breaks the checkout of this repo as a GitHub action.
func DanglingSymlink(tb testing.TB) string {
	tb.Helper()
	dir := tb.TempDir()
	link := filepath.Join(dir, "dangling-link.txt")
	require.NoError(tb, os.Symlink(filepath.Join(dir, "missing.txt"), link))
	return link
}
