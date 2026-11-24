package main

import (
	"os"

	goversion "github.com/caarlos0/go-version"
	"github.com/litebitus/go-publisher/cmd"
)

//nolint:gochecknoglobals
var (
	version   = ""
	commit    = ""
	treeState = ""
	date      = ""
	builtBy   = ""
)

func main() {
	cmd.Execute(
		buildVersion(version, commit, date, builtBy, treeState),
		os.Exit,
		os.Args[1:],
	)
}

const website = "https://github.com/litebitus/go-publisher"

func buildVersion(version, commit, date, builtBy, treeState string) goversion.Info {
	return goversion.GetVersionInfo(
		goversion.WithAppDetails("go-publisher", "Builds and publishes releases of Go projects.", website),
		func(i *goversion.Info) {
			if commit != "" {
				i.GitCommit = commit
			}
			if treeState != "" {
				i.GitTreeState = treeState
			}
			if date != "" {
				i.BuildDate = date
			}
			if version != "" {
				i.GitVersion = version
			}
			if builtBy != "" {
				i.BuiltBy = builtBy
			}
		},
	)
}
