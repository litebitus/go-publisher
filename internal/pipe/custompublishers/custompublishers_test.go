package custompublishers

import (
	"testing"

	"github.com/litebitus/go-publisher/internal/artifact"
	"github.com/litebitus/go-publisher/internal/testctx"
	"github.com/litebitus/go-publisher/internal/testlib"
	"github.com/litebitus/go-publisher/pkg/config"
	"github.com/stretchr/testify/require"
)

func TestDescription(t *testing.T) {
	require.NotEmpty(t, Pipe{}.String())
}

func TestSkip(t *testing.T) {
	t.Run("skip", func(t *testing.T) {
		require.True(t, Pipe{}.Skip(testctx.Wrap(t.Context())))
	})

	t.Run("dont skip", func(t *testing.T) {
		ctx := testctx.WrapWithCfg(t.Context(), config.Project{
			Publishers: []config.Publisher{
				{},
			},
		})

		require.False(t, Pipe{}.Skip(ctx))
	})
}

func TestPublish(t *testing.T) {
	require.NoError(t, Pipe{}.Publish(testctx.WrapWithCfg(t.Context(), config.Project{
		Publishers: []config.Publisher{
			{
				Cmd: "echo",
			},
		},
	})))
}

func TestPublishSummary(t *testing.T) {
	cfg := config.Project{
		Publishers: []config.Publisher{
			{
				Name: "custom",
				Cmd:  "echo",
			},
		},
	}

	t.Run("no artifacts runs nothing", func(t *testing.T) {
		readSummary := testlib.CaptureSummary(t)
		require.NoError(t, Pipe{}.Publish(testctx.WrapWithCfg(t.Context(), cfg)))
		require.Empty(t, readSummary())
	})

	t.Run("reports the artifacts it ran on", func(t *testing.T) {
		readSummary := testlib.CaptureSummary(t)
		ctx := testctx.WrapWithCfg(t.Context(), cfg)
		ctx.Artifacts.Add(&artifact.Artifact{
			Name: "foo",
			Path: "foo",
			Type: artifact.UploadableBinary,
		})
		require.NoError(t, Pipe{}.Publish(ctx))
		require.Equal(t, []string{"- Ran custom publisher `custom` on 1 artifacts"}, readSummary())
	})
}
