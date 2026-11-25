// Package publish contains the publishing pipe.
package publish

import (
	"fmt"

	"github.com/litebitus/go-publisher/internal/middleware/errhandler"
	"github.com/litebitus/go-publisher/internal/middleware/logging"
	"github.com/litebitus/go-publisher/internal/middleware/skip"
	"github.com/litebitus/go-publisher/internal/pipe/artifactory"
	"github.com/litebitus/go-publisher/internal/pipe/aur"
	"github.com/litebitus/go-publisher/internal/pipe/aursources"
	"github.com/litebitus/go-publisher/internal/pipe/blob"
	"github.com/litebitus/go-publisher/internal/pipe/brew"
	"github.com/litebitus/go-publisher/internal/pipe/cask"
	"github.com/litebitus/go-publisher/internal/pipe/chocolatey"
	"github.com/litebitus/go-publisher/internal/pipe/custompublishers"
	"github.com/litebitus/go-publisher/internal/pipe/docker"
	dockerv2 "github.com/litebitus/go-publisher/internal/pipe/docker/v2"
	"github.com/litebitus/go-publisher/internal/pipe/dockerdigest"
	"github.com/litebitus/go-publisher/internal/pipe/iru"
	"github.com/litebitus/go-publisher/internal/pipe/ko"
	"github.com/litebitus/go-publisher/internal/pipe/krew"
	"github.com/litebitus/go-publisher/internal/pipe/mcp"
	"github.com/litebitus/go-publisher/internal/pipe/milestone"
	"github.com/litebitus/go-publisher/internal/pipe/nix"
	"github.com/litebitus/go-publisher/internal/pipe/release"
	"github.com/litebitus/go-publisher/internal/pipe/scoop"
	"github.com/litebitus/go-publisher/internal/pipe/sign"
	"github.com/litebitus/go-publisher/internal/pipe/snapcraft"
	"github.com/litebitus/go-publisher/internal/pipe/upload"
	"github.com/litebitus/go-publisher/internal/pipe/winget"
	"github.com/litebitus/go-publisher/internal/skips"
	"github.com/litebitus/go-publisher/internal/summary"
	"github.com/litebitus/go-publisher/pkg/context"
)

// Publisher should be implemented by pipes that want to publish artifacts.
type Publisher interface {
	fmt.Stringer

	// Publish publishes the configured artifacts.
	Publish(ctx *context.Context) error
}

// New publish pipeline.
func New() Pipe {
	return Pipe{
		pipeline: []Publisher{
			blob.Pipe{},
			upload.Pipe{},
			artifactory.Pipe{},
			iru.Pipe{},
			docker.Pipe{},
			docker.ManifestPipe{},
			dockerv2.Publish{},
			ko.Pipe{},
			dockerdigest.Pipe{},
			sign.DockerPipe{},
			snapcraft.Pipe{},
			// This should be one of the last steps
			release.Pipe{},
			// brew et al use the release URL, so, they should be last
			nix.New(),
			winget.Pipe{},
			brew.Pipe{},
			cask.Pipe{},
			aur.Pipe{},
			aursources.Pipe{},
			krew.Pipe{},
			scoop.Pipe{},
			chocolatey.Pipe{},
			mcp.New(),
			milestone.Pipe{},
			custompublishers.Pipe{},
		},
	}
}

// Pipe that publishes artifacts.
type Pipe struct {
	pipeline []Publisher
}

func (Pipe) String() string                 { return "publishing" }
func (Pipe) Skip(ctx *context.Context) bool { return skips.Any(ctx, skips.Publish) }

func (p Pipe) Run(ctx *context.Context) error {
	summary.Begin(ctx.Config.ProjectName, ctx.Version)
	memo := errhandler.Memo{}
	for _, publisher := range p.pipeline {
		if err := skip.Maybe(
			publisher,
			logging.PadLog(
				publisher.String(),
				errhandler.Handle(publisher.Publish),
			),
		)(ctx); err != nil {
			if ig, ok := publisher.(Continuable); ok && ig.ContinueOnError() && !ctx.FailFast {
				memo.Memorize(fmt.Errorf("%s: %w", publisher.String(), err))
				continue
			}
			return fmt.Errorf("%s: failed to publish artifacts: %w", publisher.String(), err)
		}
	}
	return memo.Error()
}

type Continuable interface {
	ContinueOnError() bool
}
