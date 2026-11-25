package nodedist

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/litebitus/go-publisher/pkg/config"
	"github.com/stretchr/testify/require"
)

const (
	mockVersion     = "v22.10.0"
	mockArchiveName = "node-v22.10.0-linux-x64.tar.gz"
)

// countingServer serves payload at the mock archive's path and counts the
// requests it answers.
func countingServer(t *testing.T, payload []byte) *atomic.Int32 {
	t.Helper()
	var requests atomic.Int32
	mux := http.NewServeMux()
	mux.HandleFunc("/"+mockVersion+"/"+mockArchiveName, func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		_, _ = w.Write(payload)
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	SetBaseURL(t, server.URL)
	return &requests
}

// archivesIn is the path of every file under cacheDir, relative to it.
func archivesIn(t *testing.T, cacheDir string) []string {
	t.Helper()
	var files []string
	require.NoError(t, filepath.WalkDir(cacheDir, func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		relative, err := filepath.Rel(cacheDir, path)
		files = append(files, filepath.ToSlash(relative))
		return err
	}))
	return files
}

func TestDownloadStoresTheArchiveInTheCache(t *testing.T) {
	payload := []byte("fake archive bytes")
	StubRelease(t, mockVersion, mockArchiveName, payload)
	countingServer(t, payload)
	cacheDir := t.TempDir()

	got, err := Download(t.Context(), cacheDir, mockVersion, mockArchiveName)
	require.NoError(t, err)
	require.Equal(t, filepath.Join(cacheDir, mockVersion, mockArchiveName), got)
	bts, err := os.ReadFile(got)
	require.NoError(t, err)
	require.Equal(t, payload, bts)
	require.Equal(t, []string{mockVersion + "/" + mockArchiveName}, archivesIn(t, cacheDir))
}

func TestCachedServesAStoredArchiveWithoutARequest(t *testing.T) {
	payload := []byte("fake archive bytes")
	StubRelease(t, mockVersion, mockArchiveName, payload)
	requests := countingServer(t, payload)
	cacheDir := t.TempDir()

	_, cached := Cached(cacheDir, mockVersion, mockArchiveName)
	require.False(t, cached)
	downloaded, err := Download(t.Context(), cacheDir, mockVersion, mockArchiveName)
	require.NoError(t, err)

	got, cached := Cached(cacheDir, mockVersion, mockArchiveName)
	require.True(t, cached)
	require.Equal(t, downloaded, got)
	require.Equal(t, int32(1), requests.Load())
}

func TestCachedRefusesAnArchiveWhoseSHADiffers(t *testing.T) {
	StubRelease(t, mockVersion, mockArchiveName, []byte("fake archive bytes"))
	cacheDir := t.TempDir()
	corrupted := filepath.Join(cacheDir, mockVersion, mockArchiveName)
	require.NoError(t, os.MkdirAll(filepath.Dir(corrupted), 0o755))
	require.NoError(t, os.WriteFile(corrupted, []byte("corrupted bytes"), 0o644))

	_, cached := Cached(cacheDir, mockVersion, mockArchiveName)
	require.False(t, cached)
}

func TestDownloadReplacesACorruptedArchive(t *testing.T) {
	payload := []byte("fake archive bytes")
	StubRelease(t, mockVersion, mockArchiveName, payload)
	countingServer(t, payload)
	cacheDir := t.TempDir()
	corrupted := filepath.Join(cacheDir, mockVersion, mockArchiveName)
	require.NoError(t, os.MkdirAll(filepath.Dir(corrupted), 0o755))
	require.NoError(t, os.WriteFile(corrupted, []byte("corrupted bytes"), 0o644))

	got, err := Download(t.Context(), cacheDir, mockVersion, mockArchiveName)
	require.NoError(t, err)
	bts, err := os.ReadFile(got)
	require.NoError(t, err)
	require.Equal(t, payload, bts)
}

func TestDownloadEvictsTheLeastRecentlyUsedArchives(t *testing.T) {
	const olderArchiveName = "node-v22.10.0-darwin-arm64.tar.gz"
	payload := []byte("12345678")
	StubRelease(t, mockVersion, mockArchiveName, payload)
	countingServer(t, payload)
	cacheDir := t.TempDir()
	older := filepath.Join(cacheDir, mockVersion, olderArchiveName)
	require.NoError(t, os.MkdirAll(filepath.Dir(older), 0o755))
	require.NoError(t, os.WriteFile(older, payload, 0o644))
	lastUsed := time.Now().Add(-time.Hour)
	require.NoError(t, os.Chtimes(older, lastUsed, lastUsed))

	previousLimit := cacheSizeLimitBytes
	cacheSizeLimitBytes = int64(len(payload)) + 1
	t.Cleanup(func() { cacheSizeLimitBytes = previousLimit })

	_, err := Download(t.Context(), cacheDir, mockVersion, mockArchiveName)
	require.NoError(t, err)
	require.Equal(t, []string{mockVersion + "/" + mockArchiveName}, archivesIn(t, cacheDir))
}

func TestConcurrentDownloadsLeaveOneCompleteArchive(t *testing.T) {
	payload := []byte("fake archive bytes")
	StubRelease(t, mockVersion, mockArchiveName, payload)
	countingServer(t, payload)
	cacheDir := t.TempDir()

	var downloads sync.WaitGroup
	for range 4 {
		downloads.Go(func() {
			_, err := Download(t.Context(), cacheDir, mockVersion, mockArchiveName)
			require.NoError(t, err)
		})
	}
	downloads.Wait()

	require.Equal(t, []string{mockVersion + "/" + mockArchiveName}, archivesIn(t, cacheDir))
	bts, err := os.ReadFile(filepath.Join(cacheDir, mockVersion, mockArchiveName))
	require.NoError(t, err)
	require.Equal(t, payload, bts)
}

func TestDownload_BadSHA(t *testing.T) {
	StubRelease(t, mockVersion, mockArchiveName, []byte("expected"))
	countingServer(t, []byte("actual"))
	cacheDir := t.TempDir()

	_, err := Download(t.Context(), cacheDir, mockVersion, mockArchiveName)
	require.Error(t, err)
	require.Contains(t, err.Error(), "SHA-256 mismatch")
	require.Empty(t, archivesIn(t, cacheDir))
}

func TestDownload_UnknownVersion(t *testing.T) {
	_, err := Download(t.Context(), t.TempDir(), "v0.0.999", "whatever.tar.gz")
	require.Error(t, err)
	require.Contains(t, err.Error(), "no embedded entry")
}

func TestDownload_RetriesOn5xx(t *testing.T) {
	payload := []byte("fake archive bytes")
	StubRelease(t, mockVersion, mockArchiveName, payload)

	var hits atomic.Int32
	mux := http.NewServeMux()
	mux.HandleFunc("/"+mockVersion+"/"+mockArchiveName, func(w http.ResponseWriter, _ *http.Request) {
		if hits.Add(1) < 2 {
			http.Error(w, "boom", http.StatusServiceUnavailable)
			return
		}
		_, _ = w.Write(payload)
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	SetBaseURL(t, server.URL)

	prevRetry := defaultRetry
	defaultRetry = config.Retry{Attempts: 4}
	t.Cleanup(func() { defaultRetry = prevRetry })

	got, err := Download(t.Context(), t.TempDir(), mockVersion, mockArchiveName)
	require.NoError(t, err)
	require.FileExists(t, got)
	require.GreaterOrEqual(t, int(hits.Load()), 2)
}
