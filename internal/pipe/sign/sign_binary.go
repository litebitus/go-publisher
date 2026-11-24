package sign

import (
	"fmt"

	"github.com/litebitus/go-publisher/internal/artifact"
	"github.com/litebitus/go-publisher/internal/git"
	"github.com/litebitus/go-publisher/internal/ids"
	"github.com/litebitus/go-publisher/internal/pipe"
	"github.com/litebitus/go-publisher/internal/semerrgroup"
	"github.com/litebitus/go-publisher/internal/skips"
	"github.com/litebitus/go-publisher/pkg/config"
	"github.com/litebitus/go-publisher/pkg/context"
)

// defaultSignatureName tells apart binaries that differ only by target
// variant. Without the variant the signature files themselves are fine, as
// they are written to per-target directories, but the artifact name is
// ambiguous, so only one of them survives as a release asset.
const defaultSignatureName = `${artifact}_{{ .Os }}_{{ .Arch }}{{ targetVariant . }}`

// BinaryPipe signs binaries before archiving.
type BinaryPipe struct{}

func (BinaryPipe) String() string { return "signing binaries" }

func (BinaryPipe) Skip(ctx *context.Context) bool {
	return skips.Any(ctx, skips.Sign) || len(ctx.Config.BinarySigns) == 0
}

func (BinaryPipe) Dependencies(ctx *context.Context) []string {
	var cmds []string
	for _, s := range ctx.Config.BinarySigns {
		cmds = append(cmds, s.Cmd)
	}
	return cmds
}

// Default sets the Pipes defaults.
func (BinaryPipe) Default(ctx *context.Context) error {
	gpgPath, _ := git.Clean(git.Run(ctx, "config", "gpg.program"))
	if gpgPath == "" {
		gpgPath = defaultGpg
	}

	ids := ids.New("binary_signs")
	for i := range ctx.Config.BinarySigns {
		cfg := &ctx.Config.BinarySigns[i]
		if cfg.Cmd == "" {
			// gpgPath is either "gpg" (default) or the user's git config gpg.program value
			cfg.Cmd = gpgPath
		}
		if cfg.Signature == "" {
			cfg.Signature = defaultSignatureName
		}
		if len(cfg.Args) == 0 {
			cfg.Args = []string{"--output", "$signature", "--detach-sig", "$artifact"}
		}
		if cfg.Artifacts == "" {
			cfg.Artifacts = "binary"
		}
		if cfg.ID == "" {
			cfg.ID = "default"
		}
		ids.Inc(cfg.ID)
	}
	return ids.Validate()
}

// Run signs the selected binaries and registers their signature artifacts.
func (BinaryPipe) Run(ctx *context.Context) error {
	// skip-aware so that a config with `artifacts: none` doesn't mask the
	// errors of the other configs.
	g := semerrgroup.NewSkipAware(semerrgroup.New(ctx.Parallelism))
	for i := range ctx.Config.BinarySigns {
		cfg := ctx.Config.BinarySigns[i]
		g.Go(func() error {
			switch cfg.Artifacts {
			case "binary":
				// do nothing
			case "none":
				return pipe.ErrSkipSignEnabled
			default:
				return fmt.Errorf("invalid list of artifacts to sign: %s", cfg.Artifacts)
			}
			filters := []artifact.Filter{artifact.ByTypes(
				artifact.Binary,
				artifact.UniversalBinary,
			)}
			if len(cfg.IDs) > 0 {
				filters = append(filters, artifact.ByIDs(cfg.IDs...))
			}
			return sign(ctx, config.Sign(cfg), ctx.Artifacts.Filter(artifact.And(filters...)).List())
		})
	}
	return g.Wait()
}
