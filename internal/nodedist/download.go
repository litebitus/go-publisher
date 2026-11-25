package nodedist

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/litebitus/go-publisher/internal/retryx"
)

// downloadTempPrefix names an archive while it downloads into the cache, so
// the cache's size and eviction leave it out.
const downloadTempPrefix = ".download-"

// Cached returns the path of the archive
// https://nodejs.org/dist/<version>/<archiveName> in cacheDir, and whether it
// is there with the SHA-256 the embedded release index states. A cached
// archive counts as used now, for eviction. It belongs to the cache: the
// caller reads it and never changes or removes it.
func Cached(cacheDir, version, archiveName string) (string, bool) {
	expected, err := lookupSHA(version, archiveName)
	if err != nil {
		return "", false
	}
	archivePath := cachedArchivePath(cacheDir, version, archiveName)
	hash, err := fileSHA256(archivePath)
	if err != nil || hash != expected {
		return "", false
	}
	now := time.Now()
	if err := os.Chtimes(archivePath, now, now); err != nil {
		return "", false
	}
	return archivePath, true
}

// Download fetches https://nodejs.org/dist/<version>/<archiveName> into
// cacheDir, verifies its SHA-256 against the embedded release index, and
// returns its path, which belongs to the cache as [Cached]'s does. The
// archive is renamed into place once verified, so a reader never sees a
// partial one. The least recently used archives are then removed until the
// cache holds at most cacheSizeLimitBytes.
//
// The caller owns extraction (for tar.gz archives) and any further
// mutation (e.g. stripping a code signature), on its own copy.
func Download(ctx context.Context, cacheDir, version, archiveName string) (string, error) {
	expected, err := lookupSHA(version, archiveName)
	if err != nil {
		return "", err
	}
	archivePath := cachedArchivePath(cacheDir, version, archiveName)
	if err := os.MkdirAll(filepath.Dir(archivePath), 0o755); err != nil {
		return "", err
	}
	tmp, err := os.CreateTemp(filepath.Dir(archivePath), downloadTempPrefix+"*")
	if err != nil {
		return "", err
	}
	tmpName := tmp.Name()

	url := fmt.Sprintf("%s/%s/%s", distBaseURL, version, archiveName)
	hash, err := downloadTo(ctx, url, tmp)
	_ = tmp.Close()
	if err != nil {
		_ = os.Remove(tmpName)
		return "", err
	}
	if expected != hash {
		_ = os.Remove(tmpName)
		return "", fmt.Errorf("nodedist: SHA-256 mismatch for %s: expected %s, got %s", url, expected, hash)
	}
	if err := os.Rename(tmpName, archivePath); err != nil {
		_ = os.Remove(tmpName)
		return "", err
	}
	if err := evictLeastRecentlyUsed(cacheDir, archivePath); err != nil {
		return "", fmt.Errorf("nodedist: bound the cache %s: %w", cacheDir, err)
	}
	return archivePath, nil
}

// cachedArchivePath is where cacheDir keeps the archive of version.
func cachedArchivePath(cacheDir, version, archiveName string) string {
	return filepath.Join(cacheDir, version, filepath.FromSlash(archiveName))
}

// fileSHA256 is the SHA-256 of the file at path as a lower-case hex string.
func fileSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// cachedArchive is an archive in the cache, for eviction.
type cachedArchive struct {
	path     string
	size     int64
	lastUsed time.Time
}

// evictLeastRecentlyUsed removes the least recently used archives of
// cacheDir, never kept, until the archives total at most
// cacheSizeLimitBytes.
func evictLeastRecentlyUsed(cacheDir, kept string) error {
	var archives []cachedArchive
	var total int64
	err := filepath.WalkDir(cacheDir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || strings.HasPrefix(entry.Name(), downloadTempPrefix) {
			return nil
		}
		info, err := entry.Info()
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		if err != nil {
			return err
		}
		archives = append(archives, cachedArchive{path: path, size: info.Size(), lastUsed: info.ModTime()})
		total += info.Size()
		return nil
	})
	if err != nil {
		return err
	}
	slices.SortFunc(archives, func(a, b cachedArchive) int {
		return a.lastUsed.Compare(b.lastUsed)
	})
	for _, archive := range archives {
		if total <= cacheSizeLimitBytes {
			return nil
		}
		if archive.path == kept {
			continue
		}
		if err := os.Remove(archive.path); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return err
		}
		total -= archive.size
	}
	return nil
}

// downloadTo streams an HTTP GET into dst, returning the SHA-256 of
// the downloaded bytes as a lower-case hex string. Transient failures
// (5xx, 429, network errors) are retried per defaultRetry; on each
// attempt dst is rewound and truncated so a partially downloaded body
// from a failed attempt does not pollute the next one.
func downloadTo(ctx context.Context, url string, dst *os.File) (string, error) {
	var hash string
	err := retryx.Do(ctx, defaultRetry, func() error {
		if _, err := dst.Seek(0, io.SeekStart); err != nil {
			return retryx.Unrecoverable(err)
		}
		if err := dst.Truncate(0); err != nil {
			return retryx.Unrecoverable(err)
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return retryx.Unrecoverable(err)
		}
		resp, err := httpClient.Do(req)
		if err != nil {
			return retryx.HTTP(fmt.Errorf("nodedist: download %s: %w", url, err), resp)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return retryx.HTTP(fmt.Errorf("nodedist: download %s: unexpected status %s", url, resp.Status), resp)
		}
		h := sha256.New()
		if _, err := io.Copy(io.MultiWriter(dst, h), resp.Body); err != nil {
			return retryx.HTTP(err, resp)
		}
		hash = hex.EncodeToString(h.Sum(nil))
		return nil
	}, retryx.IsRetriable)
	return hash, err
}
