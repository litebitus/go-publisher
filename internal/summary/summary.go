// Package summary appends one-line entries describing what a run published
// (releases, opened pull requests, pushed packages, uploaded artifacts) to the
// GitHub Actions job summary, under a heading naming the project and version
// published. Outside GitHub Actions it does nothing — the regular logs already
// cover the same ground.
package summary

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/caarlos0/log"
)

// envSummary is the file GitHub Actions renders the job summary from.
const envSummary = "GITHUB_STEP_SUMMARY"

var (
	mu sync.Mutex
	// publishedHeading is the heading Begin names, written before the first entry.
	publishedHeading string
	// headingWritten is whether publishedHeading is in the summary already.
	headingWritten bool
)

// Begin starts the summary of a run publishing project at version. Its heading
// is written before the run's first entry, so a run that publishes nothing
// writes none.
func Begin(project, version string) {
	mu.Lock()
	defer mu.Unlock()
	publishedHeading = fmt.Sprintf("### Go Publisher Summary\n\nPublished `%s` `%s`\n\n", project, version)
	headingWritten = false
}

// Append writes a markdown bullet to the GitHub Actions job summary, with each
// item a bullet nested under it. It is a no-op when not running in GitHub
// Actions, and is safe for concurrent use.
func Append(line string, items ...string) {
	path := os.Getenv(envSummary)
	if path == "" {
		return
	}
	mu.Lock()
	defer mu.Unlock()
	var entry strings.Builder
	if !headingWritten {
		entry.WriteString(publishedHeading)
	}
	entry.WriteString("- ")
	entry.WriteString(line)
	entry.WriteString("\n")
	for _, item := range items {
		entry.WriteString("  - ")
		entry.WriteString(item)
		entry.WriteString("\n")
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err == nil {
		_, werr := f.WriteString(entry.String())
		err = errors.Join(werr, f.Close())
	}
	if err != nil {
		log.WithError(err).Warn("failed to write github actions summary")
		return
	}
	headingWritten = true
}

// Appendf is [Append] with formatting.
func Appendf(format string, a ...any) {
	Append(fmt.Sprintf(format, a...))
}
