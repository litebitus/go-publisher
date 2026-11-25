package summary

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAppendNoGitHub(t *testing.T) {
	t.Setenv(envSummary, "")
	require.NotPanics(t, func() { Append("Pushed foo to AUR") })
}

func TestAppend(t *testing.T) {
	path := filepath.Join(t.TempDir(), "summary.md")
	t.Setenv(envSummary, path)

	Append("Published 1.2.3 to GitHub with 43 assets: https://github.com/foo/bar/releases/tag/v1.2.3")
	Appendf("Pushed %s to AUR", "foo")

	bts, err := os.ReadFile(path)
	require.NoError(t, err)
	out := string(bts)
	require.Contains(t, out, "- Published 1.2.3 to GitHub with 43 assets: https://github.com/foo/bar/releases/tag/v1.2.3\n")
	require.Contains(t, out, "- Pushed foo to AUR\n")
}

func TestAppendAppends(t *testing.T) {
	path := filepath.Join(t.TempDir(), "summary.md")
	require.NoError(t, os.WriteFile(path, []byte("existing\n"), 0o644))
	t.Setenv(envSummary, path)

	Append("Pushed foo to AUR")

	bts, err := os.ReadFile(path)
	require.NoError(t, err)
	out := string(bts)
	require.Contains(t, out, "existing\n")
	require.Contains(t, out, "- Pushed foo to AUR\n")
}

func TestAppendConcurrent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "summary.md")
	t.Setenv(envSummary, path)

	var wg sync.WaitGroup
	for i := range 50 {
		wg.Go(func() {
			Appendf("line %d", i)
		})
	}
	wg.Wait()

	bts, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, 50, strings.Count(string(bts), "\n"))
}

func beginWithCleanup(t *testing.T, project, version string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "summary.md")
	t.Setenv(envSummary, path)
	t.Cleanup(func() {
		mu.Lock()
		defer mu.Unlock()
		publishedHeading = ""
		headingWritten = false
	})
	Begin(project, version)
	return path
}

func TestBeginWritesHeadingOnceBeforeEntries(t *testing.T) {
	path := beginWithCleanup(t, "mock-project", "1.2.3")

	Append("GitHub release", "`checksums.txt`", "`mock-project_1.2.3_linux_arm64.tar.gz`")
	Append("homebrew cask `mock-cask`")

	bts, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, "### Go Publisher Summary\n\n"+
		"Published `mock-project` `1.2.3`\n\n"+
		"- GitHub release\n"+
		"  - `checksums.txt`\n"+
		"  - `mock-project_1.2.3_linux_arm64.tar.gz`\n"+
		"- homebrew cask `mock-cask`\n", string(bts))
}

func TestBeginWithoutEntriesWritesNothing(t *testing.T) {
	path := beginWithCleanup(t, "mock-project", "1.2.3")

	_, err := os.Stat(path)
	require.ErrorIs(t, err, os.ErrNotExist)
}
