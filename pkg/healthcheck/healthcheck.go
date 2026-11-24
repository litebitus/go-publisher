// Package healthcheck checks for missing binaries that the user needs to
// install.
package healthcheck

import (
	"fmt"

	"github.com/litebitus/go-publisher/internal/pipe/chocolatey"
	"github.com/litebitus/go-publisher/internal/pipe/docker"
	dockerv2 "github.com/litebitus/go-publisher/internal/pipe/docker/v2"
	"github.com/litebitus/go-publisher/internal/pipe/flatpak"
	"github.com/litebitus/go-publisher/internal/pipe/makeself"
	"github.com/litebitus/go-publisher/internal/pipe/sbom"
	"github.com/litebitus/go-publisher/internal/pipe/sign"
	"github.com/litebitus/go-publisher/internal/pipe/snapcraft"
	"github.com/litebitus/go-publisher/internal/pipe/upx"
	"github.com/litebitus/go-publisher/pkg/build"
	"github.com/litebitus/go-publisher/pkg/context"
)

// HealthChecker should be implemented by pipes that want checks.
type HealthChecker interface {
	fmt.Stringer

	// Healthcheck does checks that are more complex than [DependencyChecker],
	// and returns an error if they fail.
	Healthcheck(ctx *context.Context) error
}

// DependencyChecker should be implemented by pipes that want checks.
type DependencyChecker interface {
	fmt.Stringer

	// Dependencies return the binaries of the dependencies needed.
	Dependencies(ctx *context.Context) []string
}

// HealthCheckers is the list of health checkers.
//
//nolint:gochecknoglobals
var HealthCheckers = []HealthChecker{
	dockerv2.Base{},
	docker.Pipe{},
}

// DependencyCheckers is the list of dependency checkers.
//
//nolint:gochecknoglobals
var DependencyCheckers = []DependencyChecker{
	system{},
	builds{},
	snapcraft.Pipe{},
	sign.Pipe{},
	sign.BinaryPipe{},
	sign.DockerPipe{},
	sbom.Pipe{},
	docker.Pipe{},
	docker.ManifestPipe{},
	dockerv2.Base{},
	chocolatey.Pipe{},
	flatpak.Pipe{},
	makeself.Pipe{},
	upx.Pipe{},
}

type system struct{}

func (system) String() string                         { return "system" }
func (system) Dependencies(*context.Context) []string { return []string{"git"} }

type builds struct{}

func (builds) String() string                             { return "build" }
func (builds) Dependencies(ctx *context.Context) []string { return build.Dependencies(ctx) }
