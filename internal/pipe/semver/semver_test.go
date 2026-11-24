package semver

import (
	"testing"

	"github.com/litebitus/go-publisher/internal/skips"
	"github.com/litebitus/go-publisher/internal/testctx"
	"github.com/litebitus/go-publisher/internal/testlib"
	"github.com/litebitus/go-publisher/pkg/config"
	"github.com/litebitus/go-publisher/pkg/context"
	"github.com/stretchr/testify/require"
)

func TestDescription(t *testing.T) {
	require.NotEmpty(t, Pipe{}.String())
}

func TestValidSemver(t *testing.T) {
	ctx := testctx.Wrap(t.Context(), testctx.WithCurrentTag("v1.5.2-rc1"))
	require.NoError(t, Pipe{}.Run(ctx))
	require.Equal(t, context.Semver{
		Major:      1,
		Minor:      5,
		Patch:      2,
		Prerelease: "rc1",
	}, ctx.Semver)
}

// With a tag prefix, the version is parsed from the tag without the prefix;
// a tag that does not start with it is parsed whole.
func TestPrefixedTagSemver(t *testing.T) {
	cfg := config.Project{Git: config.Git{TagPrefix: "mock-module/"}}
	ctx := testctx.WrapWithCfg(t.Context(), cfg, testctx.WithCurrentTag("mock-module/v1.5.2-rc1"))
	require.NoError(t, Pipe{}.Run(ctx))
	require.Equal(t, context.Semver{
		Major:      1,
		Minor:      5,
		Patch:      2,
		Prerelease: "rc1",
	}, ctx.Semver)

	unprefixed := testctx.WrapWithCfg(t.Context(), cfg, testctx.WithCurrentTag("other-module/v1.5.2"))
	require.ErrorContains(t, Pipe{}.Run(unprefixed), "failed to parse tag 'other-module/v1.5.2' as semver")
}

func TestInvalidSemver(t *testing.T) {
	ctx := testctx.Wrap(t.Context(), testctx.WithCurrentTag("aaaav1.5.2-rc1"))
	err := Pipe{}.Run(ctx)
	require.ErrorContains(t, err, "failed to parse tag 'aaaav1.5.2-rc1' as semver")
}

func TestInvalidSemverSkipValidate(t *testing.T) {
	ctx := testctx.Wrap(t.Context(),
		testctx.WithCurrentTag("aaaav1.5.2-rc1"),
		testctx.Skip(skips.Validate))

	testlib.AssertSkipped(t, Pipe{}.Run(ctx))
	require.Equal(t, context.Semver{
		Major:      0,
		Minor:      0,
		Patch:      0,
		Prerelease: "",
	}, ctx.Semver)
}
