// Package announce contains the announcing pipe.
package announce

import (
	"fmt"

	"github.com/litebitus/go-publisher/internal/middleware/errhandler"
	"github.com/litebitus/go-publisher/internal/middleware/logging"
	"github.com/litebitus/go-publisher/internal/middleware/skip"
	"github.com/litebitus/go-publisher/internal/pipe/bluesky"
	"github.com/litebitus/go-publisher/internal/pipe/discord"
	"github.com/litebitus/go-publisher/internal/pipe/discourse"
	"github.com/litebitus/go-publisher/internal/pipe/linkedin"
	"github.com/litebitus/go-publisher/internal/pipe/mastodon"
	"github.com/litebitus/go-publisher/internal/pipe/mattermost"
	"github.com/litebitus/go-publisher/internal/pipe/opencollective"
	"github.com/litebitus/go-publisher/internal/pipe/reddit"
	"github.com/litebitus/go-publisher/internal/pipe/slack"
	"github.com/litebitus/go-publisher/internal/pipe/smtp"
	"github.com/litebitus/go-publisher/internal/pipe/teams"
	"github.com/litebitus/go-publisher/internal/pipe/telegram"
	"github.com/litebitus/go-publisher/internal/pipe/twitter"
	"github.com/litebitus/go-publisher/internal/pipe/webhook"
	"github.com/litebitus/go-publisher/internal/skips"
	"github.com/litebitus/go-publisher/internal/tmpl"
	"github.com/litebitus/go-publisher/pkg/context"
)

// Announcer should be implemented by pipes that want to announce releases.
type Announcer interface {
	fmt.Stringer
	Announce(ctx *context.Context) error
}

//nolint:gochecknoglobals
var announcers = []Announcer{
	// XXX: keep asc sorting
	bluesky.New(),
	discord.Pipe{},
	discourse.Pipe{},
	linkedin.Pipe{},
	mastodon.Pipe{},
	mattermost.Pipe{},
	opencollective.Pipe{},
	reddit.Pipe{},
	slack.Pipe{},
	smtp.Pipe{},
	teams.Pipe{},
	telegram.Pipe{},
	twitter.Pipe{},
	webhook.Pipe{},
}

// Pipe that announces releases.
type Pipe struct{}

func (Pipe) String() string { return "announcing" }

func (Pipe) Skip(ctx *context.Context) (bool, error) {
	if skips.Any(ctx, skips.Announce) {
		return true, nil
	}
	return tmpl.New(ctx).Bool(ctx.Config.Announce.Skip)
}

// Run the pipe.
func (Pipe) Run(ctx *context.Context) error {
	memo := errhandler.Memo{}
	for _, announcer := range announcers {
		if err := skip.Maybe(
			announcer,
			logging.PadLog(announcer.String(), errhandler.Handle(announcer.Announce)),
		)(ctx); err != nil {
			memo.Memorize(fmt.Errorf("%s: %w", announcer.String(), err))
		}
	}
	return memo.Error()
}
