// Package defaults make the list of Defaulter implementations available
// so projects extending GoReleaser are able to use it, namely, GoDownloader.
package defaults

import (
	"fmt"

	"github.com/litebitus/go-publisher/internal/pipe/archive"
	"github.com/litebitus/go-publisher/internal/pipe/artifactory"
	"github.com/litebitus/go-publisher/internal/pipe/aur"
	"github.com/litebitus/go-publisher/internal/pipe/aursources"
	"github.com/litebitus/go-publisher/internal/pipe/blob"
	"github.com/litebitus/go-publisher/internal/pipe/bluesky"
	"github.com/litebitus/go-publisher/internal/pipe/brew"
	"github.com/litebitus/go-publisher/internal/pipe/build"
	"github.com/litebitus/go-publisher/internal/pipe/cask"
	"github.com/litebitus/go-publisher/internal/pipe/changelog"
	"github.com/litebitus/go-publisher/internal/pipe/checksums"
	"github.com/litebitus/go-publisher/internal/pipe/chocolatey"
	"github.com/litebitus/go-publisher/internal/pipe/discord"
	"github.com/litebitus/go-publisher/internal/pipe/discourse"
	"github.com/litebitus/go-publisher/internal/pipe/dist"
	"github.com/litebitus/go-publisher/internal/pipe/docker"
	dockerv2 "github.com/litebitus/go-publisher/internal/pipe/docker/v2"
	"github.com/litebitus/go-publisher/internal/pipe/dockerdigest"
	"github.com/litebitus/go-publisher/internal/pipe/flatpak"
	"github.com/litebitus/go-publisher/internal/pipe/gomod"
	"github.com/litebitus/go-publisher/internal/pipe/ko"
	"github.com/litebitus/go-publisher/internal/pipe/krew"
	"github.com/litebitus/go-publisher/internal/pipe/linkedin"
	"github.com/litebitus/go-publisher/internal/pipe/makeself"
	"github.com/litebitus/go-publisher/internal/pipe/mastodon"
	"github.com/litebitus/go-publisher/internal/pipe/mattermost"
	"github.com/litebitus/go-publisher/internal/pipe/mcp"
	"github.com/litebitus/go-publisher/internal/pipe/milestone"
	"github.com/litebitus/go-publisher/internal/pipe/nfpm"
	"github.com/litebitus/go-publisher/internal/pipe/nix"
	"github.com/litebitus/go-publisher/internal/pipe/notary"
	"github.com/litebitus/go-publisher/internal/pipe/opencollective"
	"github.com/litebitus/go-publisher/internal/pipe/project"
	"github.com/litebitus/go-publisher/internal/pipe/reddit"
	"github.com/litebitus/go-publisher/internal/pipe/release"
	"github.com/litebitus/go-publisher/internal/pipe/sbom"
	"github.com/litebitus/go-publisher/internal/pipe/scoop"
	"github.com/litebitus/go-publisher/internal/pipe/sign"
	"github.com/litebitus/go-publisher/internal/pipe/slack"
	"github.com/litebitus/go-publisher/internal/pipe/smtp"
	"github.com/litebitus/go-publisher/internal/pipe/snapcraft"
	"github.com/litebitus/go-publisher/internal/pipe/snapshot"
	"github.com/litebitus/go-publisher/internal/pipe/sourcearchive"
	"github.com/litebitus/go-publisher/internal/pipe/srpm"
	"github.com/litebitus/go-publisher/internal/pipe/teams"
	"github.com/litebitus/go-publisher/internal/pipe/telegram"
	"github.com/litebitus/go-publisher/internal/pipe/twitter"
	"github.com/litebitus/go-publisher/internal/pipe/universalbinary"
	"github.com/litebitus/go-publisher/internal/pipe/upload"
	"github.com/litebitus/go-publisher/internal/pipe/upx"
	"github.com/litebitus/go-publisher/internal/pipe/webhook"
	"github.com/litebitus/go-publisher/internal/pipe/winget"
	"github.com/litebitus/go-publisher/pkg/context"
)

// Defaulter can be implemented by a Piper to set default values for its
// configuration.
type Defaulter interface {
	fmt.Stringer

	// Default sets the configuration defaults
	Default(ctx *context.Context) error
}

// Defaulters is the list of defaulters.
//
//nolint:gochecknoglobals
var Defaulters = []Defaulter{
	dist.Pipe{},
	snapshot.Pipe{},
	release.Pipe{},
	project.Pipe{},
	changelog.Pipe{},
	gomod.Pipe{},
	build.Pipe{},
	universalbinary.Pipe{},
	upx.Pipe{},
	sign.BinaryPipe{},
	notary.MacOS{},
	sourcearchive.Pipe{},
	archive.Pipe{},
	makeself.Pipe{},
	nfpm.Pipe{},
	srpm.Pipe{},
	snapcraft.Pipe{},
	flatpak.Pipe{},
	checksums.Pipe{},
	sign.Pipe{},
	sign.DockerPipe{},
	sbom.Pipe{},
	docker.Pipe{},
	dockerv2.Base{},
	docker.ManifestPipe{},
	dockerdigest.Pipe{},
	artifactory.Pipe{},
	blob.Pipe{},
	upload.Pipe{},
	aur.Pipe{},
	aursources.Pipe{},
	nix.Pipe{},
	winget.Pipe{},
	brew.Pipe{},
	cask.Pipe{},
	krew.Pipe{},
	ko.Pipe{},
	scoop.Pipe{},
	mcp.Pipe{},
	discord.Pipe{},
	reddit.Pipe{},
	slack.Pipe{},
	teams.Pipe{},
	twitter.Pipe{},
	smtp.Pipe{},
	mastodon.Pipe{},
	mattermost.Pipe{},
	milestone.Pipe{},
	linkedin.Pipe{},
	telegram.Pipe{},
	webhook.Pipe{},
	chocolatey.Pipe{},
	opencollective.Pipe{},
	bluesky.Pipe{},
	discourse.Pipe{},
}
