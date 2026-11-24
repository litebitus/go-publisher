package cmd

import (
	"errors"
	"io/fs"
	"os"

	"github.com/caarlos0/log"
	"github.com/litebitus/go-publisher/internal/logext"
	"github.com/litebitus/go-publisher/pkg/config"
)

// configFileName is the configuration read where --config names none.
const configFileName = ".gopublisher.yml"

var proExplain = `Your configuration is for GoReleaser Pro.
go-publisher ignores its Pro-only features.`

func loadConfig(strict bool, path string) (config.Project, error) {
	p, path, err := loadConfigCheck(path)
	if err == nil {
		log.WithField("path", path).Debug("using configuration")
	}
	if errors.Is(err, config.ErrProConfig) {
		if strict {
			return p, err
		}
		log.WithField("explanation", proExplain).
			Warnf(
				"%s %s",
				logext.Warning("your configuration specifies"),
				logext.Keyword("pro: true"),
			)
		return p, nil
	}
	return p, err
}

func loadConfigCheck(path string) (config.Project, string, error) {
	if path == "-" {
		p, err := config.LoadReader(os.Stdin)
		return p, path, err
	}
	if path != "" {
		p, err := config.Load(path)
		return p, path, err
	}
	proj, err := config.Load(configFileName)
	if err == nil || !errors.Is(err, fs.ErrNotExist) {
		return proj, configFileName, err
	}
	// the user didn't specify a config file and the default one doesn't
	// exist, so, return an empty config and a nil err.
	log.Warn(logext.Warning("could not find a configuration file, using defaults..."))
	return config.Project{}, "", nil
}
