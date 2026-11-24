// Package commitauthor provides common commit author functionality.
package commitauthor

import (
	"github.com/litebitus/go-publisher/internal/tmpl"
	"github.com/litebitus/go-publisher/pkg/config"
	"github.com/litebitus/go-publisher/pkg/context"
)

const (
	defaultName  = "github-actions[bot]"
	defaultEmail = "github-actions[bot]@users.noreply.github.com"
)

// Get templates the commit author and returns a new [config.CommitAuthor].
func Get(ctx *context.Context, author config.CommitAuthor) (config.CommitAuthor, error) {
	if err := tmpl.New(ctx).ApplyAll(
		&author.Name,
		&author.Email,
		&author.Signing.Key,
		&author.Signing.Program,
		&author.Signing.Format,
	); err != nil {
		return config.CommitAuthor{}, err
	}
	return author, nil
}

// Default sets the default commit author name and email.
func Default(og config.CommitAuthor) config.CommitAuthor {
	if og.Name == "" {
		og.Name = defaultName
	}
	if og.Email == "" {
		og.Email = defaultEmail
	}

	// set default signing format if enabled but format not specified
	if og.Signing.Enabled && og.Signing.Format == "" {
		og.Signing.Format = "openpgp"
	}

	return og
}
