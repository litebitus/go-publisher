package base

import (
	"fmt"
	"path/filepath"

	"github.com/litebitus/go-publisher/internal/tmpl"
	"github.com/litebitus/go-publisher/pkg/build"
	"github.com/litebitus/go-publisher/pkg/config"
	"github.com/litebitus/go-publisher/pkg/context"
)

// OptionsForTarget resolves the binary name and output path for a parsed target.
func OptionsForTarget(ctx *context.Context, cfg config.Build, target build.Target, ext string) (*build.Options, error) {
	opts := build.Options{Target: target, Ext: ext}
	bin, err := tmpl.New(ctx).WithBuildOptions(opts).Apply(cfg.Binary)
	if err != nil {
		return nil, err
	}

	opts.Name = bin + ext
	dir := fmt.Sprintf("%s_%s", cfg.ID, target)
	noUnique, err := tmpl.New(ctx).Bool(cfg.NoUniqueDistDir)
	if err != nil {
		return nil, err
	}
	if noUnique {
		dir = ""
	}
	opts.Path, err = filepath.Abs(filepath.Join(ctx.Config.Dist, dir, opts.Name))
	if err != nil {
		return nil, err
	}
	return &opts, nil
}
