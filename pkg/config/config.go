// Package config contains the model and loader of the goreleaser configuration
// file.
package config

import (
	"fmt"
	"io/fs"
	"os"
	"strings"
	"time"
)

// Versioned is the configuration's `version` and `pro` keys, read before the
// rest of the file.
type Versioned struct {
	Version int
	Pro     bool
}

// Git configures how the current and previous tags are read from Git.
type Git struct {
	// Sort order of the tags that point at the same commit, in the form
	// `git tag --sort` takes. Default: `-version:refname`.
	TagSort string `yaml:"tag_sort,omitempty" json:"tag_sort,omitempty" jsonschema:"enum=-version:refname,enum=-version:creatordate,default=-version:refname"`
	// Suffix that marks a prerelease when sorting the tags that point at the
	// same commit, passed to Git as `versionsort.suffix`.
	PrereleaseSuffix string `yaml:"prerelease_suffix,omitempty" json:"prerelease_suffix,omitempty"`
	// Tags never taken as the current or previous tag. Templated.
	IgnoreTags []string `yaml:"ignore_tags,omitempty" json:"ignore_tags,omitempty"`
	// Prefix of the tags a module is released from: the current and previous
	// tags start with it, and the version is the tag without it.
	TagPrefix string `yaml:"tag_prefix,omitempty" json:"tag_prefix,omitempty"`
}

// TagWithoutPrefix is the tag with TagPrefix removed: the part that holds the
// version.
func (g Git) TagWithoutPrefix(tag string) string {
	return strings.TrimPrefix(tag, g.TagPrefix)
}

// GitHubURLs holds the URLs of a GitHub Enterprise instance.
type GitHubURLs struct {
	// API URL of the instance. Default: GitHub's public API. Templated.
	API string `yaml:"api,omitempty" json:"api,omitempty"`
	// Upload URL of the instance. Default: `api`. Templated.
	Upload string `yaml:"upload,omitempty" json:"upload,omitempty"`
	// Download URL of the instance. Default: `https://github.com`. Templated.
	Download string `yaml:"download,omitempty" json:"download,omitempty"`
	// Whether to skip verifying the instance's TLS certificate.
	SkipTLSVerify bool `yaml:"skip_tls_verify,omitempty" json:"skip_tls_verify,omitempty"`
}

// GitLabURLs holds the URLs of a self-hosted GitLab instance.
type GitLabURLs struct {
	// API URL of the instance. Default: GitLab's public API. Templated.
	API string `yaml:"api,omitempty" json:"api,omitempty"`
	// Download URL of the instance. Default: `https://gitlab.com`. Templated.
	Download string `yaml:"download,omitempty" json:"download,omitempty"`
	// Whether to skip verifying the instance's TLS certificate.
	SkipTLSVerify bool `yaml:"skip_tls_verify,omitempty" json:"skip_tls_verify,omitempty"`
	// Whether to upload the release's files to the project's Generic Package
	// Registry instead of as attachments.
	UsePackageRegistry bool `yaml:"use_package_registry,omitempty" json:"use_package_registry,omitempty"`
	// Whether the GitLab token is the CI job token, `CI_JOB_TOKEN`.
	UseJobToken bool `yaml:"use_job_token,omitempty" json:"use_job_token,omitempty"`
}

// GiteaURLs holds the URLs of a Gitea instance.
type GiteaURLs struct {
	// API URL of the instance. Templated.
	API string `yaml:"api,omitempty" json:"api,omitempty"`
	// Download URL of the instance. Default: `api` without its `/api/v1`
	// path. Templated.
	Download string `yaml:"download,omitempty" json:"download,omitempty"`
	// Whether to skip verifying the instance's TLS certificate.
	SkipTLSVerify bool `yaml:"skip_tls_verify,omitempty" json:"skip_tls_verify,omitempty"`
}

// Repo is a GitHub, GitLab or Gitea repository a release is created in.
type Repo struct {
	// Owner of the repository.
	Owner string `yaml:"owner,omitempty" json:"owner,omitempty"`
	// Name of the repository.
	Name   string `yaml:"name,omitempty" json:"name,omitempty"`
	RawURL string `yaml:"-" json:"-"`

	// Token for this repository, in place of the default one. Templated with
	// environment variables only.
	Token string `yaml:"token,omitempty" json:"token,omitempty"`
}

// String of the repo, e.g. owner/name.
func (r Repo) String() string {
	if r.isSCM() {
		return r.Owner + "/" + r.Name
	}
	return r.Owner
}

// CheckSCM returns an error if the given url is not a valid scm url.
func (r Repo) CheckSCM() error {
	if r.isSCM() {
		return nil
	}
	return fmt.Errorf("invalid scm url: %s", r.RawURL)
}

// isSCM returns true if the repo has both an owner and name.
func (r Repo) isSCM() bool {
	return r.Owner != "" && r.Name != ""
}

// RepoRef is a repository the generated files are pushed to, which may differ
// from the one being released and need its own authentication.
type RepoRef struct {
	// Owner of the repository. Templated.
	Owner string `yaml:"owner,omitempty" json:"owner,omitempty"`
	// Name of the repository. Templated.
	Name string `yaml:"name,omitempty" json:"name,omitempty"`
	// Token for this repository, in place of the default one. Templated.
	Token string `yaml:"token,omitempty" json:"token,omitempty"`
	// Branch to push to. Default: the repository's default branch, or
	// `{{ .ProjectName }}-{{ .Version }}` when `pull_request.enabled` is
	// `true`. Templated.
	Branch string `yaml:"branch,omitempty" json:"branch,omitempty"`

	// Plain Git repository to clone, commit to and push, used when its `url`
	// is set.
	Git GitRepoRef `yaml:"git,omitempty" json:"git,omitempty"`
	// Pull request to open with the changes instead of pushing them to the
	// branch only.
	PullRequest PullRequest `yaml:"pull_request,omitempty" json:"pull_request,omitempty"`
}

// GitRepoRef is a plain Git repository the generated files are pushed to.
type GitRepoRef struct {
	// Git URL to push to. Templated.
	URL string `yaml:"url,omitempty" json:"url,omitempty"`
	// Value of `GIT_SSH_COMMAND` when pulling and pushing. Default:
	// `ssh -i "{{ .KeyPath }}" -o StrictHostKeyChecking=accept-new -F /dev/null`.
	// Templated.
	SSHCommand string `yaml:"ssh_command,omitempty" json:"ssh_command,omitempty"`
	// SSH private key to push with, as a path or the key's contents. The key
	// must not be password-protected. Templated.
	PrivateKey string `yaml:"private_key,omitempty" json:"private_key,omitempty"`
}

// PullRequestBase is the repository and branch a pull request targets.
type PullRequestBase struct {
	// Owner of the target repository.
	Owner string `yaml:"owner,omitempty" json:"owner,omitempty"`
	// Name of the target repository.
	Name string `yaml:"name,omitempty" json:"name,omitempty"`
	// Target branch.
	Branch string `yaml:"branch,omitempty" json:"branch,omitempty"`
}

// PullRequest is a pull request opened with the generated files.
type PullRequest struct {
	// Whether to open a pull request.
	Enabled bool `yaml:"enabled,omitempty" json:"enabled,omitempty"`
	// Repository and branch the pull request targets.
	Base PullRequestBase `yaml:"base,omitempty" json:"base,omitempty"`
	// Whether to open the pull request as a draft.
	Draft bool `yaml:"draft,omitempty" json:"draft,omitempty"`
	// Token to open the pull request with. Templated.
	Token string `yaml:"token,omitempty" json:"token,omitempty"`
}

// HomebrewDependency is a package a Homebrew formula depends on.
type HomebrewDependency struct {
	// Name of the dependency.
	Name string `yaml:"name,omitempty" json:"name,omitempty"`
	// Type of the dependency. Takes precedence over `version` when both are
	// set.
	Type string `yaml:"type,omitempty" json:"type,omitempty"`
	// Version of the dependency.
	Version string `yaml:"version,omitempty" json:"version,omitempty"`
	// Operating system the dependency is required on.
	OS string `yaml:"os,omitempty" json:"os,omitempty" jsonschema:"enum=mac,enum=linux"`
}

// AUR is an Arch User Repository package of prebuilt binaries.
type AUR struct {
	// Name of the package, given a `-bin` suffix when it has none. Default:
	// the project name. Templated.
	Name string `yaml:"name,omitempty" json:"name,omitempty"`
	// IDs of the artifacts to use. Empty means all.
	IDs []string `yaml:"ids,omitempty" json:"ids,omitempty"`
	// Git author of the commit.
	CommitAuthor CommitAuthor `yaml:"commit_author,omitempty" json:"commit_author,omitempty"`
	// Commit message. Default: `Update to {{ .Tag }}`. Templated.
	CommitMessageTemplate string `yaml:"commit_msg_template,omitempty" json:"commit_msg_template,omitempty"`
	// Description of the package. Templated.
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	// Homepage of the package.
	Homepage string `yaml:"homepage,omitempty" json:"homepage,omitempty"`
	// SPDX identifier of the package's license.
	License string `yaml:"license,omitempty" json:"license,omitempty"`
	// Whether to keep the generated files in the dist directory without
	// pushing them; `auto` skips the push for a prerelease. Templated.
	SkipUpload string `yaml:"skip_upload,omitempty" json:"skip_upload,omitempty" jsonschema:"oneof_type=string;boolean"`
	// Download URL of each artifact. Default: the artifact's release URL on
	// the SCM the token is for. Templated.
	URLTemplate string `yaml:"url_template,omitempty" json:"url_template,omitempty"`
	// Maintainers of the package.
	Maintainers []string `yaml:"maintainers,omitempty" json:"maintainers,omitempty"`
	// Contributors of the package.
	Contributors []string `yaml:"contributors,omitempty" json:"contributors,omitempty"`
	// Packages whose features the package provides. Default: the project
	// name.
	Provides []string `yaml:"provides,omitempty" json:"provides,omitempty"`
	// Packages that conflict with the package. Default: the project name.
	Conflicts []string `yaml:"conflicts,omitempty" json:"conflicts,omitempty"`
	// Packages required to install the package.
	Depends []string `yaml:"depends,omitempty" json:"depends,omitempty"`
	// Packages that add features without being required, each in the format
	// `package: description`.
	OptDepends []string `yaml:"optdepends,omitempty" json:"optdepends,omitempty"`
	// Files that may hold user changes, kept across upgrades and removals,
	// relative to the package root without a leading slash.
	Backup []string `yaml:"backup,omitempty" json:"backup,omitempty"`
	// Release number of the package, the PKGBUILD's `pkgrel`. Default: `1`.
	Rel string `yaml:"rel,omitempty" json:"rel,omitempty"`
	// Package instructions of the PKGBUILD. Default: install each binary into
	// `${pkgdir}/usr/bin`. Templated.
	Package string `yaml:"package,omitempty" json:"package,omitempty"`
	// AUR Git URL of the package. The push is skipped when it is empty.
	// Templated.
	GitURL string `yaml:"git_url,omitempty" json:"git_url,omitempty"`
	// Value of `GIT_SSH_COMMAND` when pulling and pushing. Default:
	// `ssh -i "{{ .KeyPath }}" -o StrictHostKeyChecking=accept-new -F /dev/null`.
	// Templated.
	GitSSHCommand string `yaml:"git_ssh_command,omitempty" json:"git_ssh_command,omitempty"`
	// SSH private key to push with, as a path or the key's contents. The key
	// must not be password-protected. Templated.
	PrivateKey string `yaml:"private_key,omitempty" json:"private_key,omitempty"`
	// `GOAMD64` variant of the amd64 artifacts to use. Default: `v1`.
	Goamd64 string `yaml:"goamd64,omitempty" json:"goamd64,omitempty"`
	// Directory in the repository the files are written to. Default: `.`.
	// Templated.
	Directory string `yaml:"directory,omitempty" json:"directory,omitempty"`

	// Whether to skip this package. Templated.
	Disable string `yaml:"disable,omitempty" json:"disable,omitempty" jsonschema:"oneof_type=string;boolean"`
	// Path of the script added to the package as `<name>.install`.
	Install string `yaml:"install,omitempty" json:"install,omitempty"`
}

// AURSource is an Arch User Repository package built from source.
type AURSource struct {
	// Name of the package, with any `-bin` suffix removed. Default: the
	// project name. Templated.
	Name string `yaml:"name,omitempty" json:"name,omitempty"`
	// IDs of the source archives to use. Empty means all.
	IDs []string `yaml:"ids,omitempty" json:"ids,omitempty"`
	// Git author of the commit.
	CommitAuthor CommitAuthor `yaml:"commit_author,omitempty" json:"commit_author,omitempty"`
	// Commit message. Default: `Update to {{ .Tag }}`. Templated.
	CommitMessageTemplate string `yaml:"commit_msg_template,omitempty" json:"commit_msg_template,omitempty"`
	// Description of the package. Templated.
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	// Homepage of the package.
	Homepage string `yaml:"homepage,omitempty" json:"homepage,omitempty"`
	// SPDX identifier of the package's license.
	License string `yaml:"license,omitempty" json:"license,omitempty"`
	// Whether to keep the generated files in the dist directory without
	// pushing them; `auto` skips the push for a prerelease. Templated.
	SkipUpload string `yaml:"skip_upload,omitempty" json:"skip_upload,omitempty" jsonschema:"oneof_type=string;boolean"`
	// Download URL of each source archive. Default: the archive's release URL
	// on the SCM the token is for. Templated.
	URLTemplate string `yaml:"url_template,omitempty" json:"url_template,omitempty"`
	// Maintainers of the package.
	Maintainers []string `yaml:"maintainers,omitempty" json:"maintainers,omitempty"`
	// Contributors of the package.
	Contributors []string `yaml:"contributors,omitempty" json:"contributors,omitempty"`
	// Architectures the package builds for. Default: `[x86_64, aarch64]`.
	Arches []string `yaml:"arches,omitempty" json:"arches,omitempty"`
	// Packages whose features the package provides. Default: the project
	// name.
	Provides []string `yaml:"provides,omitempty" json:"provides,omitempty"`
	// Packages that conflict with the package. Default: the project name.
	Conflicts []string `yaml:"conflicts,omitempty" json:"conflicts,omitempty"`
	// Packages required to install the package.
	Depends []string `yaml:"depends,omitempty" json:"depends,omitempty"`
	// Packages that add features without being required, each in the format
	// `package: description`.
	OptDepends []string `yaml:"optdepends,omitempty" json:"optdepends,omitempty"`
	// Packages required to build the package. Default: `[go, git]`.
	MakeDepends []string `yaml:"makedepends,omitempty" json:"makedepends,omitempty"`
	// Files that may hold user changes, kept across upgrades and removals,
	// relative to the package root without a leading slash.
	Backup []string `yaml:"backup,omitempty" json:"backup,omitempty"`
	// Release number of the package, the PKGBUILD's `pkgrel`. Default: `1`.
	Rel string `yaml:"rel,omitempty" json:"rel,omitempty"`
	// Prepare instructions of the PKGBUILD.
	Prepare string `yaml:"prepare,omitempty" json:"prepare,omitempty"`
	// Build instructions of the PKGBUILD.
	Build string `yaml:"build,omitempty" json:"build,omitempty"`
	// Package instructions of the PKGBUILD. Templated.
	Package string `yaml:"package,omitempty" json:"package,omitempty"`
	// AUR Git URL of the package. The push is skipped when it is empty.
	// Templated.
	GitURL string `yaml:"git_url,omitempty" json:"git_url,omitempty"`
	// Value of `GIT_SSH_COMMAND` when pulling and pushing. Default:
	// `ssh -i "{{ .KeyPath }}" -o StrictHostKeyChecking=accept-new -F /dev/null`.
	// Templated.
	GitSSHCommand string `yaml:"git_ssh_command,omitempty" json:"git_ssh_command,omitempty"`
	// SSH private key to push with, as a path or the key's contents. The key
	// must not be password-protected. Templated.
	PrivateKey string `yaml:"private_key,omitempty" json:"private_key,omitempty"`
	// Has no effect.
	Goamd64 string `yaml:"goamd64,omitempty" json:"goamd64,omitempty"`
	// Directory in the repository the files are written to. Default: `.`.
	// Templated.
	Directory string `yaml:"directory,omitempty" json:"directory,omitempty"`

	// Whether to skip this package. Templated.
	Disable string `yaml:"disable,omitempty" json:"disable,omitempty" jsonschema:"oneof_type=string;boolean"`
	// Path of the script added to the package as `<name>.install`.
	Install string `yaml:"install,omitempty" json:"install,omitempty"`
}

// Homebrew is a Homebrew formula of the release's archives.
//
// Deprecated: use `homebrew_casks`.
type Homebrew struct {
	// Name of the formula. Default: the project name. Templated.
	Name string `yaml:"name,omitempty" json:"name,omitempty"`
	// Repository the formula is pushed to.
	Repository RepoRef `yaml:"repository,omitempty" json:"repository,omitempty"`
	// Git author of the commit.
	CommitAuthor CommitAuthor `yaml:"commit_author,omitempty" json:"commit_author,omitempty"`
	// Commit message. Default:
	// `Brew formula update for {{ .ProjectName }} version {{ .Tag }}`.
	// Templated.
	CommitMessageTemplate string `yaml:"commit_msg_template,omitempty" json:"commit_msg_template,omitempty"`
	// Directory in the repository the formula is written to.
	Directory string `yaml:"directory,omitempty" json:"directory,omitempty"`
	// Caveats shown to the user of the formula.
	Caveats string `yaml:"caveats,omitempty" json:"caveats,omitempty"`
	// Install instructions of the formula. Default: `bin.install` of the
	// binary. Templated.
	Install string `yaml:"install,omitempty" json:"install,omitempty"`
	// Install instructions added to the default ones. Templated.
	ExtraInstall string `yaml:"extra_install,omitempty" json:"extra_install,omitempty"`
	// Instructions run after the install instructions.
	PostInstall string `yaml:"post_install,omitempty" json:"post_install,omitempty"`
	// Packages the formula depends on.
	Dependencies []HomebrewDependency `yaml:"dependencies,omitempty" json:"dependencies,omitempty"`
	// Test instructions of the formula, run by `brew test`. Templated.
	Test string `yaml:"test,omitempty" json:"test,omitempty"`
	// Packages that conflict with the formula.
	Conflicts []string `yaml:"conflicts,omitempty" json:"conflicts,omitempty"`
	// Description of the formula. Templated.
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	// Homepage of the formula.
	Homepage string `yaml:"homepage,omitempty" json:"homepage,omitempty"`
	// SPDX identifier of the formula's license.
	License string `yaml:"license,omitempty" json:"license,omitempty"`
	// Whether to keep the generated formula in the dist directory without
	// pushing it; `auto` skips the push for a prerelease. Templated.
	SkipUpload string `yaml:"skip_upload,omitempty" json:"skip_upload,omitempty" jsonschema:"oneof_type=string;boolean"`
	// Download strategy of the `url` stanza, implemented in the tap
	// repository.
	DownloadStrategy string `yaml:"download_strategy,omitempty" json:"download_strategy,omitempty"`
	// Download URL of each archive. Default: the archive's release URL on the
	// SCM the token is for. Templated.
	URLTemplate string `yaml:"url_template,omitempty" json:"url_template,omitempty"`
	// Headers of the `url` stanza.
	URLHeaders []string `yaml:"url_headers,omitempty" json:"url_headers,omitempty"`
	// Path given to a `require_relative` at the top of the formula.
	CustomRequire string `yaml:"custom_require,omitempty" json:"custom_require,omitempty"`
	// Ruby block added to the formula.
	CustomBlock string `yaml:"custom_block,omitempty" json:"custom_block,omitempty"`
	// IDs of the archives to use. Empty means all.
	IDs []string `yaml:"ids,omitempty" json:"ids,omitempty"`
	// `GOARM` version of the 32-bit arm archives to use. Default: `6`.
	Goarm string `yaml:"goarm,omitempty" json:"goarm,omitempty" jsonschema:"oneof_type=string;integer"`
	// `GOAMD64` variant of the amd64 archives to use. Default: `v1`.
	Goamd64 string `yaml:"goamd64,omitempty" json:"goamd64,omitempty"`
	// Service block of the formula.
	Service string `yaml:"service,omitempty" json:"service,omitempty"`
}

// HomebrewCask is a Homebrew cask of the release's archives.
type HomebrewCask struct {
	// Name of the cask. Default: the project name. Templated.
	Name string `yaml:"name,omitempty" json:"name,omitempty"`
	// Repository the cask is pushed to.
	Repository RepoRef `yaml:"repository,omitempty" json:"repository,omitempty"`
	// Git author of the commit.
	CommitAuthor CommitAuthor `yaml:"commit_author,omitempty" json:"commit_author,omitempty"`
	// Commit message. Default:
	// `Brew cask update for {{ .ProjectName }} version {{ .Tag }}`.
	// Templated.
	CommitMessageTemplate string `yaml:"commit_msg_template,omitempty" json:"commit_msg_template,omitempty"`
	// Directory in the repository the cask is written to. Default: `Casks`.
	Directory string `yaml:"directory,omitempty" json:"directory,omitempty"`
	// Caveats shown to the user of the cask.
	Caveats string `yaml:"caveats,omitempty" json:"caveats,omitempty"`
	// Description of the cask. Templated.
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	// Homepage of the cask.
	Homepage string `yaml:"homepage,omitempty" json:"homepage,omitempty"`
	// Whether to keep the generated cask in the dist directory without
	// pushing it; `auto` skips the push for a prerelease. Templated.
	SkipUpload string `yaml:"skip_upload,omitempty" json:"skip_upload,omitempty" jsonschema:"oneof_type=string;boolean"`
	// Ruby block placed at the top of the cask.
	CustomBlock string `yaml:"custom_block,omitempty" json:"custom_block,omitempty"`
	// IDs of the archives to use. Empty means all.
	IDs []string `yaml:"ids,omitempty" json:"ids,omitempty"`
	// Relative path of a service moved into `~/Library/Services` on
	// installation.
	Service string `yaml:"service,omitempty" json:"service,omitempty"`
	// Names of the binaries in the cask. Default: `name`. Templated.
	Binaries []string `yaml:"binaries,omitempty" json:"binaries,omitempty"`
	// Paths of the manpage files. Templated.
	Manpages []string `yaml:"manpages,omitempty" json:"manpages,omitempty"`
	// Download URL of the cask and its parameters.
	URL HomebrewCaskURL `yaml:"url,omitempty" json:"url,omitempty"`
	// Paths of the shell completion files. Templated.
	Completions HomebrewCaskCompletions `yaml:"completions,omitempty" json:"completions,omitempty"`
	// Casks and formulas the cask depends on.
	Dependencies []HomebrewCaskDependency `yaml:"dependencies,omitempty" json:"dependencies,omitempty"`
	// Casks that conflict with the cask.
	Conflicts []HomebrewCaskConflict `yaml:"conflicts,omitempty" json:"conflicts,omitempty"`
	// Steps run before and after the cask's install and uninstall. Templated.
	Hooks HomebrewCaskHooks `yaml:"hooks,omitempty" json:"hooks,omitempty"`
	// Procedures to uninstall the cask.
	Uninstall HomebrewCaskUninstall `yaml:"uninstall,omitempty" json:"uninstall,omitempty"`
	// Procedures for a complete uninstall, including user files and shared
	// resources. When empty, the cask declares `zap :not_required`.
	Zap HomebrewCaskUninstall `yaml:"zap,omitempty" json:"zap,omitempty"`

	// Shell completions generated by running the binary at install time.
	GenerateCompletionsFromExecutable HomebrewCaskGeneratedCompletions `yaml:"generate_completions_from_executable,omitempty" json:"generate_completions_from_executable,omitempty"`

	// Has no effect.
	License string `yaml:"license,omitempty" json:"license,omitempty"`

	// Path of the manpage file.
	//
	// Deprecated: use `manpages`.
	Manpage string `yaml:"manpage,omitempty" json:"manpage,omitempty" jsonschema:"deprecated=true"`

	// Name of the binary in the cask.
	//
	// Deprecated: use `binaries`.
	Binary string `yaml:"binary,omitempty" json:"binary,omitempty" jsonschema:"deprecated=true"`
}

// HomebrewCaskURL is the download URL of a cask and its parameters.
type HomebrewCaskURL struct {
	// Download URL of each archive. Default: the archive's release URL on the
	// SCM the token is for. Templated.
	Template string `yaml:"template,omitempty" json:"template,omitempty"`

	// Download strategy or format of the download. Templated.
	Using string `yaml:"using,omitempty" json:"using,omitempty"`
	// HTTP cookies sent with the download request. Templated.
	Cookies map[string]string `yaml:"cookies,omitempty" json:"cookies,omitempty"`
	// HTTP referer of the download request. Templated.
	Referer string `yaml:"referer,omitempty" json:"referer,omitempty"`
	// Additional HTTP headers of the download request, written as the cask's
	// `header`. Templated.
	Headers []string `yaml:"headers,omitempty" json:"headers,omitempty"`
	// HTTP user agent of the download request. Templated.
	UserAgent string `yaml:"user_agent,omitempty" json:"user_agent,omitempty"`
	// Body of the download request when it is a POST. Templated.
	Data map[string]string `yaml:"data,omitempty" json:"data,omitempty"`

	// Has no effect.
	//
	// Deprecated: by Homebrew.
	Verified string `yaml:"verified,omitempty" json:"verified,omitempty" jsonschema:"deprecated=true"`
}

// HomebrewCaskUninstall is a set of procedures that remove a cask.
type HomebrewCaskUninstall struct {
	// Labels of the launchd jobs to remove.
	Launchctl []string `yaml:"launchctl,omitempty" json:"launchctl,omitempty"`
	// Bundle IDs of the running apps to quit.
	Quit []string `yaml:"quit,omitempty" json:"quit,omitempty"`
	// Names of the login items to remove.
	LoginItem []string `yaml:"login_item,omitempty" json:"login_item,omitempty"`
	// Paths to delete.
	Delete []string `yaml:"delete,omitempty" json:"delete,omitempty"`
	// Paths to move to the Trash.
	Trash []string `yaml:"trash,omitempty" json:"trash,omitempty"`
}

// HomebrewCaskHooks are the steps run before and after a cask's install and
// uninstall.
type HomebrewCaskHooks struct {
	// Steps run before the install and uninstall.
	Pre HomebrewCaskHook `yaml:"pre,omitempty" json:"pre,omitempty"`
	// Steps run after the install and uninstall.
	Post HomebrewCaskHook `yaml:"post,omitempty" json:"post,omitempty"`
}

// HomebrewCaskHook is the steps of a cask hook, in Homebrew's declarative
// install steps language.
type HomebrewCaskHook struct {
	// Steps run at install. Templated.
	InstallSteps string `yaml:"install_steps,omitempty" json:"install_steps,omitempty"`
	// Steps run at uninstall. Templated.
	UninstallSteps string `yaml:"uninstall_steps,omitempty" json:"uninstall_steps,omitempty"`

	// Ruby run at install. Ignored when `install_steps` is set. Templated.
	//
	// Deprecated: use `install_steps`.
	Install string `yaml:"install,omitempty" json:"install,omitempty" jsonschema:"deprecated=true"`

	// Ruby run at uninstall. Ignored when `uninstall_steps` is set.
	// Templated.
	//
	// Deprecated: use `uninstall_steps`.
	Uninstall string `yaml:"uninstall,omitempty" json:"uninstall,omitempty" jsonschema:"deprecated=true"`
}

// HomebrewCaskConflict is a package that conflicts with a cask.
type HomebrewCaskConflict struct {
	// Name of the conflicting cask.
	Cask string `yaml:"cask,omitempty" json:"cask,omitempty"`

	// Has no effect.
	//
	// Deprecated: by Homebrew.
	Formula string `yaml:"formula,omitempty" json:"formula,omitempty" jsonschema:"deprecated=true"`
}

// HomebrewCaskDependency is a package a cask depends on.
type HomebrewCaskDependency struct {
	// Name of the cask depended on.
	Cask string `yaml:"cask,omitempty" json:"cask,omitempty"`
	// Name of the formula depended on.
	Formula string `yaml:"formula,omitempty" json:"formula,omitempty"`
	// TODO: support macos, arch
	// https://github.com/Homebrew/brew/blob/master/docs/Cask-Cookbook.md#stanza-depends_on
}

// HomebrewCaskCompletions are the paths of a cask's shell completion files.
type HomebrewCaskCompletions struct {
	// Path of the bash completion file. Templated.
	Bash string `yaml:"bash,omitempty" json:"bash,omitempty"`
	// Path of the zsh completion file. Templated.
	Zsh string `yaml:"zsh,omitempty" json:"zsh,omitempty"`
	// Path of the fish completion file. Templated.
	Fish string `yaml:"fish,omitempty" json:"fish,omitempty"`
}

// HomebrewCaskGeneratedCompletions are shell completions a cask generates by
// running its binary at install time.
type HomebrewCaskGeneratedCompletions struct {
	// Path of the executable in the cask. Default: the cask's first binary.
	Executable string `yaml:"executable,omitempty" json:"executable,omitempty"`
	// Arguments passed to the executable.
	Args []string `yaml:"args,omitempty" json:"args,omitempty"`
	// Base name of the completion files. Default: the binary's name.
	BaseName string `yaml:"base_name,omitempty" json:"base_name,omitempty"`
	// How the shell's name is passed to the executable: a known format, or
	// any other string used as a prefix.
	ShellParameterFormat string `yaml:"shell_parameter_format,omitempty" json:"shell_parameter_format,omitempty"`
	// Shells to generate completions for. Default: `[bash, zsh, fish]`, and
	// `pwsh` too with the `cobra` and `typer` formats.
	Shells []string `yaml:"shells,omitempty" json:"shells,omitempty"`
}

// Nix is a Nix package of the release's archives.
type Nix struct {
	// Name of the package. Default: the project name. Templated.
	Name string `yaml:"name,omitempty" json:"name,omitempty"`
	// Path of the package's file in the repository. Default:
	// `pkgs/<name>/default.nix`. Templated.
	Path string `yaml:"path,omitempty" json:"path,omitempty"`
	// Repository the package is pushed to.
	Repository RepoRef `yaml:"repository,omitempty" json:"repository,omitempty"`
	// Git author of the commit.
	CommitAuthor CommitAuthor `yaml:"commit_author,omitempty" json:"commit_author,omitempty"`
	// Commit message. Default:
	// `{{ .ProjectName }}: {{ .PreviousTag }} -> {{ .Tag }}`. Templated.
	CommitMessageTemplate string `yaml:"commit_msg_template,omitempty" json:"commit_msg_template,omitempty"`
	// IDs of the archives to use. Empty means all.
	IDs []string `yaml:"ids,omitempty" json:"ids,omitempty"`
	// `GOAMD64` variant of the amd64 archives to use. Default: `v1`.
	Goamd64 string `yaml:"goamd64,omitempty" json:"goamd64,omitempty"`
	// Whether to keep the generated package in the dist directory without
	// pushing it; `auto` skips the push for a prerelease. Templated.
	SkipUpload string `yaml:"skip_upload,omitempty" json:"skip_upload,omitempty" jsonschema:"oneof_type=string;boolean"`
	// Download URL of each archive. Default: the archive's release URL on the
	// SCM the token is for. Templated.
	URLTemplate string `yaml:"url_template,omitempty" json:"url_template,omitempty"`
	// Install script of the package. Default: copy the binary into
	// `$out/bin`, wrapped with `makeWrapper` when `dependencies` are set.
	// Templated.
	Install string `yaml:"install,omitempty" json:"install,omitempty"`
	// Install instructions added to the default ones. Templated.
	ExtraInstall string `yaml:"extra_install,omitempty" json:"extra_install,omitempty"`
	// Instructions run after the install script. Templated.
	PostInstall string `yaml:"post_install,omitempty" json:"post_install,omitempty"`
	// Description of the package. Templated.
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	// Homepage of the package. Templated.
	Homepage string `yaml:"homepage,omitempty" json:"homepage,omitempty"`
	// Nix license name of the package.
	License string `yaml:"license,omitempty" json:"license,omitempty"`

	// Runtime dependencies of the package.
	Dependencies []NixDependency `yaml:"dependencies,omitempty" json:"dependencies,omitempty"`

	// Formatter run on the generated file.
	Formatter string `yaml:"formatter,omitempty" json:"formatter,omitempty" jsonschema:"enum=alejandra,enum=nixfmt"`

	// Name of the binary `nix run` invokes, the package's `meta.mainProgram`.
	// Templated.
	MainProgram string `yaml:"main_program,omitempty" json:"main_program,omitempty"`
}

// NixDependency is a runtime dependency of a Nix package.
type NixDependency struct {
	// Name of the dependency. Required.
	Name string `yaml:"name" json:"name"`
	// Operating system the dependency is required on.
	OS string `yaml:"os,omitempty" json:"os,omitempty" jsonschema:"enum=linux,enum=darwin"`
}

// Winget is a WinGet manifest of the release's archives.
type Winget struct {
	// Name of the manifest, also its artifact ID, so it must be unique.
	// Default: the project name. Templated.
	Name string `yaml:"name,omitempty" json:"name,omitempty"`
	// Name of the package. Default: `name`. Templated.
	PackageName string `yaml:"package_name,omitempty" json:"package_name,omitempty"`
	// Identifier of the package. Default: `publisher` without spaces, a dot,
	// and `name`. Templated.
	PackageIdentifier string `yaml:"package_identifier,omitempty" json:"package_identifier,omitempty"`
	// Locale of the manifests and of the locale manifest's file name. Default:
	// `en-US`. Templated.
	DefaultLocale string `yaml:"default_locale,omitempty" json:"default_locale,omitempty"`
	// Name of the publisher. Required. Templated.
	Publisher string `yaml:"publisher" json:"publisher"`
	// URL of the publisher. Templated.
	PublisherURL string `yaml:"publisher_url,omitempty" json:"publisher_url,omitempty"`
	// Support URL of the publisher. Templated.
	PublisherSupportURL string `yaml:"publisher_support_url,omitempty" json:"publisher_support_url,omitempty"`
	// URL of the privacy policy. Templated.
	PrivacyURL string `yaml:"privacy_url,omitempty" json:"privacy_url,omitempty"`
	// Copyright notice of the package. Templated.
	Copyright string `yaml:"copyright,omitempty" json:"copyright,omitempty"`
	// URL of the copyright notice. Templated.
	CopyrightURL string `yaml:"copyright_url,omitempty" json:"copyright_url,omitempty"`
	// Author of the package. Templated.
	Author string `yaml:"author,omitempty" json:"author,omitempty"`
	// Path of the manifests in the repository. Default: under `manifests`,
	// the lowercased first character of `package_identifier`, then
	// `package_identifier` with its dots as slashes, then the version.
	// Templated.
	Path string `yaml:"path,omitempty" json:"path,omitempty"`
	// Repository the manifests are pushed to. Required.
	Repository RepoRef `yaml:"repository" json:"repository"`
	// Git author of the commit.
	CommitAuthor CommitAuthor `yaml:"commit_author,omitempty" json:"commit_author,omitempty"`
	// Commit message. Default:
	// `New version: {{ .PackageIdentifier }} {{ .Version }}`. Templated.
	CommitMessageTemplate string `yaml:"commit_msg_template,omitempty" json:"commit_msg_template,omitempty"`
	// IDs of the archives to use. Empty means all.
	IDs []string `yaml:"ids,omitempty" json:"ids,omitempty"`
	// `GOAMD64` variant of the amd64 archives to use. Default: `v1`.
	Goamd64 string `yaml:"goamd64,omitempty" json:"goamd64,omitempty"`
	// Whether to keep the generated manifests in the dist directory without
	// pushing them; `auto` skips the push for a prerelease. Templated.
	SkipUpload string `yaml:"skip_upload,omitempty" json:"skip_upload,omitempty" jsonschema:"oneof_type=string;boolean"`
	// Download URL of each archive. Default: the archive's release URL on the
	// SCM the token is for. Templated.
	URLTemplate string `yaml:"url_template,omitempty" json:"url_template,omitempty"`
	// Short description of the package. Required. Templated.
	ShortDescription string `yaml:"short_description" json:"short_description"`
	// Long description of the package. Templated.
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	// Homepage of the package. Templated.
	Homepage string `yaml:"homepage,omitempty" json:"homepage,omitempty"`
	// License of the package. Required. Templated.
	License string `yaml:"license" json:"license"`
	// URL of the license. Templated.
	LicenseURL string `yaml:"license_url,omitempty" json:"license_url,omitempty"`
	// Release notes of the version; `{{ .Changelog }}` holds the generated
	// ones. Templated.
	ReleaseNotes string `yaml:"release_notes,omitempty" json:"release_notes,omitempty"`
	// URL of the release notes. Templated.
	ReleaseNotesURL string `yaml:"release_notes_url,omitempty" json:"release_notes_url,omitempty"`
	// Installation notes, or a URL to them. Templated.
	InstallationNotes string `yaml:"installation_notes,omitempty" json:"installation_notes,omitempty"`
	// Tags of the package, lowercased with spaces replaced by hyphens.
	Tags []string `yaml:"tags,omitempty" json:"tags,omitempty"`
	// Packages the package depends on. Templated.
	Dependencies []WingetDependency `yaml:"dependencies,omitempty" json:"dependencies,omitempty"`
	// Locale manifests that override fields for other locales.
	AdditionalLocales []WingetLocale `yaml:"additional_locales,omitempty" json:"additional_locales,omitempty"`
}

// WingetLocale is a locale manifest overriding a WinGet manifest's fields for
// one locale; a field left unset takes the default locale's value.
type WingetLocale struct {
	// Locale of the manifest, unique and other than `default_locale`.
	// Required. Templated.
	Locale string `yaml:"locale" json:"locale"`
	// Name of the publisher. Templated.
	Publisher string `yaml:"publisher,omitempty" json:"publisher,omitempty"`
	// URL of the publisher. Templated.
	PublisherURL string `yaml:"publisher_url,omitempty" json:"publisher_url,omitempty"`
	// Support URL of the publisher. Templated.
	PublisherSupportURL string `yaml:"publisher_support_url,omitempty" json:"publisher_support_url,omitempty"`
	// URL of the privacy policy. Templated.
	PrivacyURL string `yaml:"privacy_url,omitempty" json:"privacy_url,omitempty"`
	// Author of the package. Templated.
	Author string `yaml:"author,omitempty" json:"author,omitempty"`
	// Name of the package. Templated.
	PackageName string `yaml:"package_name,omitempty" json:"package_name,omitempty"`
	// Homepage of the package. Templated.
	Homepage string `yaml:"homepage,omitempty" json:"homepage,omitempty"`
	// License of the package. Templated.
	License string `yaml:"license,omitempty" json:"license,omitempty"`
	// URL of the license. Templated.
	LicenseURL string `yaml:"license_url,omitempty" json:"license_url,omitempty"`
	// Copyright notice of the package. Templated.
	Copyright string `yaml:"copyright,omitempty" json:"copyright,omitempty"`
	// URL of the copyright notice. Templated.
	CopyrightURL string `yaml:"copyright_url,omitempty" json:"copyright_url,omitempty"`
	// Short description of the package. Templated.
	ShortDescription string `yaml:"short_description,omitempty" json:"short_description,omitempty"`
	// Long description of the package. Templated.
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	// Tags of the package, lowercased with spaces replaced by hyphens.
	Tags []string `yaml:"tags,omitempty" json:"tags,omitempty"`
	// Release notes of the version; `{{ .Changelog }}` holds the generated
	// ones. Templated.
	ReleaseNotes string `yaml:"release_notes,omitempty" json:"release_notes,omitempty"`
	// URL of the release notes. Templated.
	ReleaseNotesURL string `yaml:"release_notes_url,omitempty" json:"release_notes_url,omitempty"`
	// Installation notes, or a URL to them. Templated.
	InstallationNotes string `yaml:"installation_notes,omitempty" json:"installation_notes,omitempty"`
}

// WingetDependency is a package a WinGet package depends on.
type WingetDependency struct {
	// Identifier of the package depended on. Required. Templated.
	PackageIdentifier string `yaml:"package_identifier" json:"package_identifier"`
	// Minimum version of the package depended on. Templated.
	MinimumVersion string `yaml:"minimum_version,omitempty" json:"minimum_version,omitempty"`
}

// Krew is a Krew plugin manifest of the release's archives.
type Krew struct {
	// IDs of the archives to use. Empty means all.
	IDs []string `yaml:"ids,omitempty" json:"ids,omitempty"`
	// Name of the plugin. Default: the project name. Templated.
	Name string `yaml:"name,omitempty" json:"name,omitempty"`
	// Repository the manifest is pushed to.
	Repository RepoRef `yaml:"repository,omitempty" json:"repository,omitempty"`
	// Git author of the commit.
	CommitAuthor CommitAuthor `yaml:"commit_author,omitempty" json:"commit_author,omitempty"`
	// Commit message. Default:
	// `Krew manifest update for {{ .ProjectName }} version {{ .Tag }}`.
	// Templated.
	CommitMessageTemplate string `yaml:"commit_msg_template,omitempty" json:"commit_msg_template,omitempty"`
	// Caveats shown to the user of the plugin. Templated.
	Caveats string `yaml:"caveats,omitempty" json:"caveats,omitempty"`
	// Short description of the plugin. Templated.
	ShortDescription string `yaml:"short_description,omitempty" json:"short_description,omitempty"`
	// Description of the plugin. Templated.
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	// Homepage of the plugin. Templated.
	Homepage string `yaml:"homepage,omitempty" json:"homepage,omitempty"`
	// Download URL of each archive. Default: the archive's release URL on the
	// SCM the token is for. Templated.
	URLTemplate string `yaml:"url_template,omitempty" json:"url_template,omitempty"`
	// `GOARM` version of the 32-bit arm archives to use. Default: `6`.
	Goarm string `yaml:"goarm,omitempty" json:"goarm,omitempty" jsonschema:"oneof_type=string;integer"`
	// `GOAMD64` variant of the amd64 archives to use. Default: `v1`.
	Goamd64 string `yaml:"goamd64,omitempty" json:"goamd64,omitempty"`
	// Whether to keep the generated manifest in the dist directory without
	// pushing it; `auto` skips the push for a prerelease. Templated.
	SkipUpload string `yaml:"skip_upload,omitempty" json:"skip_upload,omitempty" jsonschema:"oneof_type=string;boolean"`
}

// Ko is a container image built with ko.
type Ko struct {
	// ID of the image. Default: the project name.
	ID string `yaml:"id,omitempty" json:"id,omitempty"`
	// ID of the build whose settings the image uses. Default: `id`.
	Build string `yaml:"build,omitempty" json:"build,omitempty"`
	// Main package to build, as a relative path. Default: the build's `main`.
	Main string `yaml:"main,omitempty" json:"main,omitempty"`
	// Working directory of the build. Default: the build's `dir`.
	WorkingDir string `yaml:"working_dir,omitempty" json:"working_dir,omitempty"`
	// Base image, fetched from its registry, or else from the local Docker
	// daemon. Default: `cgr.dev/chainguard/static`. Templated.
	BaseImage string `yaml:"base_image,omitempty" json:"base_image,omitempty"`
	// Labels of the image. Values are templated.
	Labels map[string]string `yaml:"labels,omitempty" json:"labels,omitempty"`
	// Annotations of the OCI manifest. Values are templated.
	Annotations map[string]string `yaml:"annotations,omitempty" json:"annotations,omitempty"`
	// Default user the image runs as.
	User string `yaml:"user,omitempty" json:"user,omitempty"`
	// Repositories to push to: the image is built for the first and copied
	// to the others. Replaced by `KO_DOCKER_REPO` when it is set. Templated.
	Repositories []string `yaml:"repositories,omitempty" json:"repositories,omitempty"`
	// Platforms to build and publish; empty ones are ignored. Default:
	// `[linux/amd64]`. Templated.
	Platforms []string `yaml:"platforms,omitempty" json:"platforms,omitempty"`
	// Tags to push; empty ones are ignored. Default: `[latest]`. Templated.
	Tags []string `yaml:"tags,omitempty" json:"tags,omitempty"`
	// Creation time of the image, in seconds since the Unix epoch. Templated.
	CreationTime string `yaml:"creation_time,omitempty" json:"creation_time,omitempty"`
	// Creation time of the files in the `kodata` directory, in seconds since
	// the Unix epoch. Templated.
	KoDataCreationTime string `yaml:"ko_data_creation_time,omitempty" json:"ko_data_creation_time,omitempty"`
	// Format of the image's SBOM. Default: `spdx`.
	SBOM string `yaml:"sbom,omitempty" json:"sbom,omitempty" jsonschema:"enum=spdx,enum=none,default=spdx"`
	// Directory the SBOM is written to. Unset, no SBOM file is written.
	SBOMDirectory string `yaml:"sbom_directory,omitempty" json:"sbom_directory,omitempty"`
	// Linker flags of the build. Default: the build's `ldflags`. Templated.
	Ldflags []string `yaml:"ldflags,omitempty" json:"ldflags,omitempty"`
	// Flags of the build. Default: the build's `flags`. Templated.
	Flags []string `yaml:"flags,omitempty" json:"flags,omitempty"`
	// Environment of the build. Default: the build's `env`. Templated.
	Env []string `yaml:"env,omitempty" json:"env,omitempty"`
	// Whether the image is named by the repository alone, with nothing added.
	Bare bool `yaml:"bare,omitempty" json:"bare,omitempty"`
	// Whether to add the full import path after the repository.
	PreserveImportPaths bool `yaml:"preserve_import_paths,omitempty" json:"preserve_import_paths,omitempty"`
	// Whether to add the import path's base, without its MD5 hash, after the
	// repository.
	BaseImportPaths bool `yaml:"base_import_paths,omitempty" json:"base_import_paths,omitempty"`
	// Registry domain of the images a snapshot publishes to the local Docker
	// daemon. Templated.
	LocalDomain string `yaml:"local_domain,omitempty" json:"local_domain,omitempty"`

	// Whether to skip this image. Templated.
	Disable string `yaml:"disable,omitempty" json:"disable,omitempty" jsonschema:"oneof_type=string;boolean"`

	// Linux capabilities set on the binary in the image, in the form
	// Docker's `--cap-add` takes. Setting them disables ko's layer cache for
	// the image.
	LinuxCapabilities []string `yaml:"linux_capabilities,omitempty" json:"linux_capabilities,omitempty"`

	// Repository to push to. Replaced by `KO_DOCKER_REPO` when it is set.
	// Templated.
	//
	// Deprecated: use `repositories`.
	Repository string `yaml:"repository,omitempty" json:"repository,omitempty" jsonschema:"deprecated=true"`
}

// Scoop is a Scoop manifest of the release's archives.
type Scoop struct {
	// Name of the manifest. Default: the project name. Templated.
	Name string `yaml:"name,omitempty" json:"name,omitempty"`
	// IDs of the archives to use. Empty means all.
	IDs []string `yaml:"ids,omitempty" json:"ids,omitempty"`
	// Bucket repository the manifest is pushed to.
	Repository RepoRef `yaml:"repository,omitempty" json:"repository,omitempty"`
	// Directory in the repository the manifest is written to.
	Directory string `yaml:"directory,omitempty" json:"directory,omitempty"`
	// Git author of the commit.
	CommitAuthor CommitAuthor `yaml:"commit_author,omitempty" json:"commit_author,omitempty"`
	// Commit message. Default:
	// `Scoop update for {{ .ProjectName }} version {{ .Tag }}`. Templated.
	CommitMessageTemplate string `yaml:"commit_msg_template,omitempty" json:"commit_msg_template,omitempty"`
	// Homepage of the app. Templated.
	Homepage string `yaml:"homepage,omitempty" json:"homepage,omitempty"`
	// Description of the app. Templated.
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	// License of the app.
	License string `yaml:"license,omitempty" json:"license,omitempty"`
	// Download URL of each archive. Default: the archive's release URL on the
	// SCM the token is for. Templated.
	URLTemplate string `yaml:"url_template,omitempty" json:"url_template,omitempty"`
	// Data kept between updates of the app.
	Persist []string `yaml:"persist,omitempty" json:"persist,omitempty"`
	// Whether to keep the generated manifest in the dist directory without
	// pushing it; `auto` skips the push for a prerelease. Templated.
	SkipUpload string `yaml:"skip_upload,omitempty" json:"skip_upload,omitempty" jsonschema:"oneof_type=string;boolean"`
	// Commands run before the app is installed.
	PreInstall []string `yaml:"pre_install,omitempty" json:"pre_install,omitempty"`
	// Commands run after the app is installed.
	PostInstall []string `yaml:"post_install,omitempty" json:"post_install,omitempty"`
	// Apps the app depends on.
	Depends []string `yaml:"depends,omitempty" json:"depends,omitempty"`
	// Start menu shortcuts, each an executable and its label, optionally
	// followed by two more values.
	Shortcuts [][]string `yaml:"shortcuts,omitempty" json:"shortcuts,omitempty"`
	// `GOAMD64` variant of the amd64 archives to use. Default: `v1`.
	Goamd64 string `yaml:"goamd64,omitempty" json:"goamd64,omitempty"`
}

// CommitAuthor is the Git author of a commit.
type CommitAuthor struct {
	// Name of the author. Default: `github-actions[bot]`. Templated.
	Name string `yaml:"name,omitempty" json:"name,omitempty"`
	// Email of the author. Default:
	// `github-actions[bot]@users.noreply.github.com`. Templated.
	Email string `yaml:"email,omitempty" json:"email,omitempty"`
	// Signing of the commit, for a plain Git repository.
	Signing CommitSigning `yaml:"signing,omitempty" json:"signing,omitempty"`
	// Whether to leave the committer out of GitHub API calls, so GitHub signs
	// the commit as the GitHub App.
	UseGitHubAppToken bool `yaml:"use_github_app_token,omitempty" json:"use_github_app_token,omitempty"`
}

// CommitSigning is the signing of a Git commit.
type CommitSigning struct {
	// Whether to sign the commit.
	Enabled bool `yaml:"enabled,omitempty" json:"enabled,omitempty"`
	// Signing key: a key ID, fingerprint, email address or path to a key
	// file. Templated.
	Key string `yaml:"key,omitempty" json:"key,omitempty"`
	// Program that signs the commit. Templated.
	Program string `yaml:"program,omitempty" json:"program,omitempty"`
	// Format of the signature. Default: `openpgp`. Templated.
	Format string `yaml:"format,omitempty" json:"format,omitempty" jsonschema:"enum=openpgp,enum=x509,enum=ssh,default=openpgp"`
}

// IgnoredBuild is a combination of target values left out of a build's
// matrix.
type IgnoredBuild struct {
	// `GOOS` of the targets to leave out.
	Goos string `yaml:"goos,omitempty" json:"goos,omitempty"`
	// `GOARCH` of the targets to leave out.
	Goarch string `yaml:"goarch,omitempty" json:"goarch,omitempty"`
	// `GOAMD64` of the targets to leave out.
	Goamd64 string `yaml:"goamd64,omitempty" json:"goamd64,omitempty"`
	// `GO386` of the targets to leave out.
	Go386 string `yaml:"go386,omitempty" json:"go386,omitempty"`
	// `GOARM` of the targets to leave out, in the exact form of the build's
	// `goarm`.
	Goarm string `yaml:"goarm,omitempty" json:"goarm,omitempty" jsonschema:"oneof_type=string;integer"`
	// `GOARM64` of the targets to leave out.
	Goarm64 string `yaml:"goarm64,omitempty" json:"goarm64,omitempty"`
	// `GOMIPS` of the targets to leave out.
	Gomips string `yaml:"gomips,omitempty" json:"gomips,omitempty"`
	// `GOPPC64` of the targets to leave out.
	Goppc64 string `yaml:"goppc64,omitempty" json:"goppc64,omitempty"`
	// `GORISCV64` of the targets to leave out.
	Goriscv64 string `yaml:"goriscv64,omitempty" json:"goriscv64,omitempty"`
}

// StringArray is a list of strings; a single string is read as a list of
// one.
type StringArray []string

// FlagArray is a list of flags; a single string is read as its
// whitespace-separated words.
type FlagArray []string

// Build is a build of the project's binaries for a set of targets.
type Build struct {
	// ID of the build, unique among the builds. Default: the project name.
	ID string `yaml:"id,omitempty" json:"id,omitempty"`
	// `GOOS` values the `go` builder builds for. Default:
	// `[linux, darwin, windows]`.
	Goos []string `yaml:"goos,omitempty" json:"goos,omitempty"`
	// `GOARCH` values the `go` builder builds for. Default:
	// `[amd64, arm64, 386]`.
	Goarch []string `yaml:"goarch,omitempty" json:"goarch,omitempty"`
	// `GOAMD64` variants built for `amd64`. Default: `[v1]`.
	Goamd64 []string `yaml:"goamd64,omitempty" json:"goamd64,omitempty"`
	// `GO386` variants built for `386`. Default: `[sse2]`.
	Go386 []string `yaml:"go386,omitempty" json:"go386,omitempty"`
	// `GOARM` versions built for `arm`, each optionally with its
	// floating-point ABI, as in `7,softfloat`; one ABI per version. Default:
	// `[6]`.
	Goarm []string `yaml:"goarm,omitempty" json:"goarm,omitempty"`
	// `GOARM64` variants built for `arm64`. Default: `[v8.0]`.
	Goarm64 []string `yaml:"goarm64,omitempty" json:"goarm64,omitempty"`
	// `GOMIPS` and `GOMIPS64` variants built for the MIPS architectures.
	// Default: `[hardfloat]`.
	Gomips []string `yaml:"gomips,omitempty" json:"gomips,omitempty"`
	// `GOPPC64` variants built for `ppc64`. Default: `[power8]`.
	Goppc64 []string `yaml:"goppc64,omitempty" json:"goppc64,omitempty"`
	// `GORISCV64` variants built for `riscv64`. Default: `[rva20u64]`.
	Goriscv64 []string `yaml:"goriscv64,omitempty" json:"goriscv64,omitempty"`
	// Targets to build, in place of the matrix the `goos`, `goarch` and
	// variant fields form. Default: the builder's targets.
	Targets []string `yaml:"targets,omitempty" json:"targets,omitempty"`
	// Target combinations left out of the matrix.
	Ignore []IgnoredBuild `yaml:"ignore,omitempty" json:"ignore,omitempty"`
	// Working directory of the build. Default: `.`.
	Dir string `yaml:"dir,omitempty" json:"dir,omitempty"`
	// Main file or package to build, a package when `gomod.proxy` is on.
	// Default: `.`.
	Main string `yaml:"main,omitempty" json:"main,omitempty"`
	// Name of the binary; a path, as in `bin/app`, puts it in a directory.
	// Default: the project name.
	Binary string `yaml:"binary,omitempty" json:"binary,omitempty"`
	// Commands run before and after each target is built. Templated.
	Hooks BuildHookConfig `yaml:"hooks,omitempty" json:"hooks,omitempty"`
	// Build implementation. Default: `go`.
	Builder string `yaml:"builder,omitempty" json:"builder,omitempty" jsonschema:"enum=,enum=go,enum=rust,enum=zig,enum=bun,enum=deno,enum=node,enum=uv,enum=poetry"`
	// Modification time set on the binary; empty leaves it unchanged.
	// Templated.
	ModTimestamp string `yaml:"mod_timestamp,omitempty" json:"mod_timestamp,omitempty"`
	// Whether to skip the build. Templated.
	Skip string `yaml:"skip,omitempty" json:"skip,omitempty" jsonschema:"oneof_type=string;boolean"`
	// Build tool to run. Default: the builder's tool, `go` for the `go`
	// builder. Templated.
	Tool string `yaml:"tool,omitempty" json:"tool,omitempty"`
	// Command of `tool` that builds. Default: the builder's command, `build`
	// for the `go` builder.
	Command string `yaml:"command,omitempty" json:"command,omitempty"`
	// Whether to write the binaries into the dist directory itself instead of
	// a directory per build target. Templated.
	NoUniqueDistDir string `yaml:"no_unique_dist_dir,omitempty" json:"no_unique_dist_dir,omitempty" jsonschema:"oneof_type=string;boolean"`
	// Whether to skip checking that `main` has a main function.
	NoMainCheck   bool   `yaml:"no_main_check,omitempty" json:"no_main_check,omitempty"`
	UnproxiedMain string `yaml:"-" json:"-"` // used by gomod.proxy
	UnproxiedDir  string `yaml:"-" json:"-"` // used by gomod.proxy

	BuildDetails `yaml:",inline" json:",inline"`
	// Build settings that replace the build's own for the targets they
	// match; each sets at least `goos` and `goarch`.
	BuildDetailsOverrides []BuildDetailsOverride `yaml:"overrides,omitempty" json:"overrides,omitempty"`

	// This is used internally only.
	InternalDefaults BuildInternalDefaults `yaml:"-" json:"-"`

	// Go binary the build runs. Default: `go`. Templated.
	//
	// Deprecated: use `tool`.
	GoBinary string `yaml:"gobinary,omitempty" json:"gobinary,omitempty" jsonschema:"deprecated=true"`
}

// BuildInternalDefaults records which of a build's settings were defaulted.
type BuildInternalDefaults struct {
	// whether the pipe set the current binary.
	// this is true when the user didn't set a binary name.
	Binary bool

	// whether the pipe set the current ID.
	// this is true when the user didn't set an ID.
	ID bool
}

// BuildDetailsOverride is a set of build settings for the targets it matches.
type BuildDetailsOverride struct {
	// `GOOS` of the targets to match. Required.
	Goos string `yaml:"goos" json:"goos"`
	// `GOARCH` of the targets to match. Required.
	Goarch string `yaml:"goarch" json:"goarch"`
	// `GOAMD64` of the targets to match; empty matches the default variant.
	Goamd64 string `yaml:"goamd64,omitempty" json:"goamd64,omitempty"`
	// `GO386` of the targets to match; empty matches the default variant.
	Go386 string `yaml:"go386,omitempty" json:"go386,omitempty"`
	// `GOARM64` of the targets to match; empty matches the default variant.
	Goarm64 string `yaml:"goarm64,omitempty" json:"goarm64,omitempty"`
	// `GOARM` of the targets to match; empty matches the default version.
	Goarm string `yaml:"goarm,omitempty" json:"goarm,omitempty" jsonschema:"oneof_type=string;integer"`
	// `GOMIPS` of the targets to match; empty matches the default variant.
	Gomips string `yaml:"gomips,omitempty" json:"gomips,omitempty"`
	// `GOPPC64` of the targets to match; empty matches the default variant.
	Goppc64 string `yaml:"goppc64,omitempty" json:"goppc64,omitempty"`
	// `GORISCV64` of the targets to match; empty matches the default variant.
	Goriscv64    string `yaml:"goriscv64,omitempty" json:"goriscv64,omitempty"`
	BuildDetails `yaml:",inline" json:",inline"`
}

// BuildDetails is the compiler settings of a build.
type BuildDetails struct {
	// Build mode of the binaries, passed as `-buildmode`.
	Buildmode string `yaml:"buildmode,omitempty" json:"buildmode,omitempty" jsonschema:"enum=c-archive,enum=c-shared,enum=pie,enum=wheel,enum=sdist,enum=,default="`
	// Linker flags, joined into one `-ldflags` value. Default for a Go build:
	// `-s -w -X main.version={{.Version}} -X main.commit={{.Commit}} -X main.date={{.Date}} -X main.builtBy=go-publisher`.
	// Templated.
	Ldflags StringArray `yaml:"ldflags,omitempty" json:"ldflags,omitempty"`
	// Build tags, joined into one `-tags` value. Templated.
	Tags FlagArray `yaml:"tags,omitempty" json:"tags,omitempty"`
	// Extra flags of the build command. Templated.
	Flags FlagArray `yaml:"flags,omitempty" json:"flags,omitempty"`
	// Assembler flags, each passed as `-asmflags`. Templated.
	Asmflags StringArray `yaml:"asmflags,omitempty" json:"asmflags,omitempty"`
	// Compiler flags, each passed as `-gcflags`. Templated.
	Gcflags StringArray `yaml:"gcflags,omitempty" json:"gcflags,omitempty"`
	// Environment variables of the build, in the format `KEY=value`.
	// Templated.
	Env []string `yaml:"env,omitempty" json:"env,omitempty"`
}

// BuildHookConfig is the commands run before and after a build.
type BuildHookConfig struct {
	// Commands run before the build.
	Pre Hooks `yaml:"pre,omitempty" json:"pre,omitempty"`
	// Commands run after the build.
	Post Hooks `yaml:"post,omitempty" json:"post,omitempty"`
}

// Hooks is a list of commands; a single string is read as a list of one.
type Hooks []Hook

// Hook is a command run as a hook; a string is read as its `cmd`.
type Hook struct {
	// Working directory of the command. Templated.
	Dir string `yaml:"dir,omitempty" json:"dir,omitempty"`
	// Command to run. Templated.
	Cmd string `yaml:"cmd,omitempty" json:"cmd,omitempty"`
	// Environment variables of the command, in the format `KEY=value`.
	// Templated.
	Env []string `yaml:"env,omitempty" json:"env,omitempty"`
	// Whether to always print the command's output, otherwise shown only in
	// debug mode.
	Output bool `yaml:"output,omitempty" json:"output,omitempty"`
}

// FormatOverride is the archive formats used for one `GOOS`.
type FormatOverride struct {
	// `GOOS` the formats apply to.
	Goos string `yaml:"goos,omitempty" json:"goos,omitempty"`
	// Archive formats for that `GOOS`; `none` skips the archive.
	Formats StringArray `yaml:"formats,omitempty" json:"formats,omitempty" jsonschema:"enum=tar,enum=tgz,enum=tar.gz,enum=zip,enum=gz,enum=tar.xz,enum=txz,enum=tar.zst,enum=tzst,enum=xz,enum=binary,enum=none,default=tar.gz"`

	// Archive format for that `GOOS`.
	//
	// Deprecated: use `formats`.
	Format string `yaml:"format,omitempty" json:"format,omitempty" jsonschema:"enum=tar,enum=tgz,enum=tar.gz,enum=zip,enum=gz,enum=tar.xz,enum=txz,enum=tar.zst,enum=tzst,enum=xz,enum=binary,enum=none,default=tar.gz,deprecated=true"`
}

// File is a set of files added to an archive; a string is read as its `src`.
type File struct {
	// Glob of the files to add. Templated.
	Source string `yaml:"src,omitempty" json:"src,omitempty"`
	// Directory the files go in, their paths kept relative to the glob.
	Destination string `yaml:"dst,omitempty" json:"dst,omitempty"`
	// Whether to drop the files' parent directories, putting each file
	// directly in `dst`.
	StripParent bool `yaml:"strip_parent,omitempty" json:"strip_parent,omitempty"`
	// File info set on the files. Default: copied from each file.
	Info    FileInfo `yaml:"info,omitempty" json:"info,omitempty"`
	Default bool     `yaml:"-" json:"-"`
}

// FileInfo is the ownership, mode and modification time of a file.
type FileInfo struct {
	// Owner name of the file. Templated.
	Owner string `yaml:"owner,omitempty" json:"owner,omitempty"`
	// Group name of the file. Templated.
	Group string `yaml:"group,omitempty" json:"group,omitempty"`
	// Mode of the file, as in `0644`.
	Mode os.FileMode `yaml:"mode,omitempty" json:"mode,omitempty"`
	// Modification time of the file, in RFC 3339 format. Templated.
	MTime       string    `yaml:"mtime,omitempty" json:"mtime,omitempty"`
	ParsedMTime time.Time `yaml:"-" json:"-"`
}

// UniversalBinary is a macOS universal binary joining the darwin binaries of
// builds.
type UniversalBinary struct {
	// ID of the universal binary. Default: the project name.
	ID string `yaml:"id,omitempty" json:"id,omitempty"`
	// IDs of the builds whose binaries are joined. Default: `id`.
	IDs []string `yaml:"ids,omitempty" json:"ids,omitempty"`
	// Name of the universal binary. Default: `{{ .ProjectName }}`. Templated.
	NameTemplate string `yaml:"name_template,omitempty" json:"name_template,omitempty"`
	// Whether to remove the single-architecture binaries from the artifacts.
	Replace bool `yaml:"replace,omitempty" json:"replace,omitempty"`
	// Commands run before and after the universal binary is made. Templated.
	Hooks BuildHookConfig `yaml:"hooks,omitempty" json:"hooks,omitempty"`
	// Modification time set on the binary; empty leaves it unchanged.
	// Templated.
	ModTimestamp string `yaml:"mod_timestamp,omitempty" json:"mod_timestamp,omitempty"`
}

// UPX is a compression of built binaries with `upx`.
type UPX struct {
	// Whether to compress the binaries. Templated.
	Enabled string `yaml:"enabled,omitempty" json:"enabled,omitempty" jsonschema:"oneof_type=string;boolean"`
	// IDs of the builds whose binaries are compressed. Empty means all.
	IDs []string `yaml:"ids,omitempty" json:"ids,omitempty"`
	// `GOOS` values of the binaries to compress. Empty means all.
	Goos []string `yaml:"goos,omitempty" json:"goos,omitempty"`
	// `GOARCH` values of the binaries to compress. Empty means all.
	Goarch []string `yaml:"goarch,omitempty" json:"goarch,omitempty"`
	// `GOARM` values of the binaries to compress. Empty means all.
	Goarm []string `yaml:"goarm,omitempty" json:"goarm,omitempty"`
	// `GOAMD64` values of the binaries to compress. Empty means all.
	Goamd64 []string `yaml:"goamd64,omitempty" json:"goamd64,omitempty"`
	// `upx` executable to run; compression is skipped when it is not found.
	// Default: `upx`.
	Binary string `yaml:"binary,omitempty" json:"binary,omitempty"`
	// Compression level passed to `upx`.
	Compress string `yaml:"compress,omitempty" json:"compress,omitempty" jsonschema:"enum=1,enum=2,enum=3,enum=4,enum=5,enum=6,enum=7,enum=8,enum=9,enum=best,enum=,default="`
	// Whether to try LZMA compression.
	LZMA bool `yaml:"lzma,omitempty" json:"lzma,omitempty"`
	// Whether to try all compression methods and filters.
	Brute bool `yaml:"brute,omitempty" json:"brute,omitempty"`
}

// Archive is an archive of the binaries of builds, one per target.
type Archive struct {
	// ID of the archive, unique among the archives. Default: `default`.
	ID string `yaml:"id,omitempty" json:"id,omitempty"`
	// IDs of the builds whose binaries are archived. Empty means all.
	IDs []string `yaml:"ids,omitempty" json:"ids,omitempty"`
	// File info set on the archived binaries; their `mode` is always `0755`.
	// Default: copied from each binary.
	BuildsInfo FileInfo `yaml:"builds_info,omitempty" json:"builds_info,omitempty"`
	// Name of the archive, without its extension. Default:
	// `{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}` and the
	// target's variant, with `{{ .Binary }}` in place of the project name for
	// the `binary` format. Templated.
	NameTemplate string `yaml:"name_template,omitempty" json:"name_template,omitempty"`
	// Formats of the archive; `binary` uploads the binaries unarchived.
	// Default: `[tar.gz]`.
	Formats StringArray `yaml:"formats,omitempty" json:"formats,omitempty" jsonschema:"enum=tar,enum=tgz,enum=tar.gz,enum=zip,enum=gz,enum=tar.xz,enum=txz,enum=tar.zst,enum=tzst,enum=xz,enum=binary,default=tar.gz"`
	// Compression settings of the `tar.zst` and `tzst` formats.
	Zstd Zstd `yaml:"zstd,omitempty" json:"zstd,omitempty"`
	// Formats that replace `formats` for a `GOOS`.
	FormatOverrides []FormatOverride `yaml:"format_overrides,omitempty" json:"format_overrides,omitempty"`
	// Whether to put the archive's files in a directory named after the
	// archive; any other string names the directory. Templated.
	WrapInDirectory string `yaml:"wrap_in_directory,omitempty" json:"wrap_in_directory,omitempty" jsonschema:"oneof_type=string;boolean"`
	// Whether to drop the directories of the binaries' paths, putting each
	// binary at the archive's root.
	StripBinaryDirectory bool `yaml:"strip_binary_directory,omitempty" json:"strip_binary_directory,omitempty"`
	// Extra files added to the archive. Default:
	// `[license*, LICENSE*, readme*, README*, changelog*, CHANGELOG*]`.
	Files []File `yaml:"files,omitempty" json:"files,omitempty"`
	// Whether to make one archive holding only `files`, with no binaries.
	Meta bool `yaml:"meta,omitempty" json:"meta,omitempty"`
	// Whether to allow the archives of different targets to hold different
	// numbers of binaries.
	AllowDifferentBinaryCount bool `yaml:"allow_different_binary_count,omitempty" json:"allow_different_binary_count,omitempty"`

	// Format of the archive.
	//
	// Deprecated: use `formats`.
	Format string `yaml:"format,omitempty" json:"format,omitempty" jsonschema:"enum=tar,enum=tgz,enum=tar.gz,enum=zip,enum=gz,enum=tar.xz,enum=txz,enum=tar.zst,enum=tzst,enum=xz,enum=binary,default=tar.gz,deprecated=true"`

	// IDs of the builds whose binaries are archived.
	//
	// Deprecated: use `ids`.
	Builds []string `yaml:"builds,omitempty" json:"builds,omitempty" jsonschema:"deprecated=true"`
}

// Zstd is the compression settings of a zstd archive.
type Zstd struct {
	// Compression level. Default: `default`.
	Level string `yaml:"level,omitempty" json:"level,omitempty" jsonschema:"enum=fastest,enum=default,enum=better,enum=best"`
	// Base-2 logarithm of the window size in bytes. Default: set by `level`.
	WindowSize *int `yaml:"window_size,omitempty" json:"window_size,omitempty" jsonschema:"minimum=10,maximum=29"`
}

// ReleaseNotesMode is what a release does with the notes of a release that
// already exists.
type ReleaseNotesMode string

const (
	ReleaseNotesModeKeepExisting ReleaseNotesMode = "keep-existing"
	ReleaseNotesModeAppend       ReleaseNotesMode = "append"
	ReleaseNotesModeReplace      ReleaseNotesMode = "replace"
	ReleaseNotesModePrepend      ReleaseNotesMode = "prepend"
)

// Release is the release of the current tag on the repository's SCM.
type Release struct {
	// GitHub repository of the release. Default: the repository of the Git
	// remote.
	GitHub Repo `yaml:"github,omitempty" json:"github,omitempty"`
	// GitLab repository of the release. Default: the repository of the Git
	// remote.
	GitLab Repo `yaml:"gitlab,omitempty" json:"gitlab,omitempty"`
	// Gitea repository of the release. Default: the repository of the Git
	// remote.
	Gitea Repo `yaml:"gitea,omitempty" json:"gitea,omitempty"`
	// Whether to leave the release as a draft instead of publishing it.
	Draft bool `yaml:"draft,omitempty" json:"draft,omitempty"`
	// Whether to delete existing draft releases of the same name before
	// creating the release, when `draft` is set. GitHub only.
	ReplaceExistingDraft bool `yaml:"replace_existing_draft,omitempty" json:"replace_existing_draft,omitempty"`
	// Whether to use an existing draft release of the same name as the
	// release. GitHub only.
	UseExistingDraft bool `yaml:"use_existing_draft,omitempty" json:"use_existing_draft,omitempty"`
	// Commit the tag is created on when it does not exist on the remote yet.
	// GitHub only. Templated.
	TargetCommitish string `yaml:"target_commitish,omitempty" json:"target_commitish,omitempty"`
	// Whether to skip the release. Templated.
	Disable string `yaml:"disable,omitempty" json:"disable,omitempty" jsonschema:"oneof_type=string;boolean"`
	// Whether to create the release without uploading artifacts to it.
	// Templated.
	SkipUpload string `yaml:"skip_upload,omitempty" json:"skip_upload,omitempty" jsonschema:"oneof_type=string;boolean"`
	// Whether to mark the release as a prerelease; `auto` marks it when the
	// tag has a prerelease part.
	Prerelease string `yaml:"prerelease,omitempty" json:"prerelease,omitempty"`
	// Whether to mark the release as the latest; a prerelease never is.
	// GitHub only. Templated.
	MakeLatest string `yaml:"make_latest,omitempty" json:"make_latest,omitempty" jsonschema:"oneof_type=string;boolean"`
	// Name of the release. Default: `{{ .Tag }}`. Templated.
	NameTemplate string `yaml:"name_template,omitempty" json:"name_template,omitempty"`
	// IDs of the artifacts to upload. Empty means all.
	IDs []string `yaml:"ids,omitempty" json:"ids,omitempty"`
	// Existing files uploaded to the release.
	ExtraFiles []ExtraFile `yaml:"extra_files,omitempty" json:"extra_files,omitempty"`
	// Category of the discussion created for the release. GitHub only.
	DiscussionCategoryName string `yaml:"discussion_category_name,omitempty" json:"discussion_category_name,omitempty"`
	// Text placed before the release notes. Templated.
	Header string `yaml:"header,omitempty" json:"header,omitempty"`
	// Text placed after the release notes. Templated.
	Footer string `yaml:"footer,omitempty" json:"footer,omitempty"`

	// What to do with the notes of a release that already exists. Default:
	// `keep-existing`.
	ReleaseNotesMode ReleaseNotesMode `yaml:"mode,omitempty" json:"mode,omitempty" jsonschema:"enum=keep-existing,enum=append,enum=prepend,enum=replace,default=keep-existing"`
	// Whether to replace an artifact of the same name already on the release.
	ReplaceExistingArtifacts bool `yaml:"replace_existing_artifacts,omitempty" json:"replace_existing_artifacts,omitempty"`
	// Whether to also upload `metadata.json` and `artifacts.json`.
	IncludeMeta bool `yaml:"include_meta,omitempty" json:"include_meta,omitempty"`
	// Checks run on the release before the build.
	Preflight ReleasePreflight `yaml:"preflight,omitempty" json:"preflight,omitempty"`
}

// ReleasePreflight is the checks run on a release before the build.
type ReleasePreflight struct {
	// Whether a failed check aborts the release before the build; otherwise
	// it only logs a warning. Templated.
	FailOnError string `yaml:"fail_on_error,omitempty" json:"fail_on_error,omitempty" jsonschema:"oneof_type=string;boolean"`
}

// Milestone is a repository milestone closed on release.
type Milestone struct {
	// Repository of the milestone. Default: the repository of the Git remote.
	Repo Repo `yaml:"repo,omitempty" json:"repo,omitempty"`
	// Whether to close the milestone.
	Close bool `yaml:"close,omitempty" json:"close,omitempty"`
	// Whether a failure to close the milestone fails the release.
	FailOnError bool `yaml:"fail_on_error,omitempty" json:"fail_on_error,omitempty"`
	// Name of the milestone. Default: `{{ .Tag }}`. Templated.
	NameTemplate string `yaml:"name_template,omitempty" json:"name_template,omitempty"`
}

// ExtraFile is a set of existing files to include.
type ExtraFile struct {
	// Glob of the files to add. Templated.
	Glob string `yaml:"glob,omitempty" json:"glob,omitempty"`
	// Name of the file, in place of its base name; the glob must then match
	// one file. Templated.
	NameTemplate string `yaml:"name_template,omitempty" json:"name_template,omitempty"`
}

// NFPM is a set of system packages of the binaries of builds, one per format
// and target.
type NFPM struct {
	NFPMOverridables `yaml:",inline" json:",inline"`
	// Settings that replace the package's own for a format, keyed by format.
	Overrides map[string]NFPMOverridables `yaml:"overrides,omitempty" json:"overrides,omitempty"`

	// ID of the packages, unique among the package sets. Default: `default`.
	ID string `yaml:"id,omitempty" json:"id,omitempty"`
	// IDs of the builds whose binaries are packaged. Empty means all.
	IDs []string `yaml:"ids,omitempty" json:"ids,omitempty"`
	// Formats of the packages to build; `msix` packages the Windows binaries.
	Formats []string `yaml:"formats,omitempty" json:"formats,omitempty" jsonschema:"enum=apk,enum=deb,enum=rpm,enum=termux.deb,enum=archlinux,enum=ipk,enum=msix"`
	// Section of the package.
	Section string `yaml:"section,omitempty" json:"section,omitempty"`
	// Priority of the package.
	Priority string `yaml:"priority,omitempty" json:"priority,omitempty"`
	// Vendor of the package.
	Vendor string `yaml:"vendor,omitempty" json:"vendor,omitempty"`
	// Homepage of the package. Templated.
	Homepage string `yaml:"homepage,omitempty" json:"homepage,omitempty"`
	// Maintainer of the package. Templated.
	Maintainer string `yaml:"maintainer,omitempty" json:"maintainer,omitempty"`
	// Description of the package. Templated.
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	// License of the package.
	License string `yaml:"license,omitempty" json:"license,omitempty"`
	// Directory the binaries are installed in; the `msix` format ignores it.
	// Default: `/usr/bin`. Templated.
	Bindir string `yaml:"bindir,omitempty" json:"bindir,omitempty"`
	// Directories the libraries built with a `buildmode` are installed in.
	Libdirs Libdirs `yaml:"libdirs,omitempty" json:"libdirs,omitempty"`
	// Path of the changelog YAML file, in chglog format.
	Changelog string `yaml:"changelog,omitempty" json:"changelog,omitempty"`
	// Modification time of the package and its binaries, in RFC 3339 format.
	// Templated.
	MTime string `yaml:"mtime,omitempty" json:"mtime,omitempty" `
	// Whether to make a package with no binaries, holding only its contents
	// and dependencies.
	Meta bool `yaml:"meta,omitempty" json:"meta,omitempty"`

	ParsedMTime time.Time `yaml:"-" json:"-"`

	// `GOAMD64` variants of the amd64 binaries to package. Empty means all.
	GoAmd64 []string `yaml:"goamd64,omitempty" json:"goamd64,omitempty"`

	// IDs of the builds whose binaries are packaged.
	//
	// Deprecated: use `ids`.
	Builds []string `yaml:"builds,omitempty" json:"builds,omitempty" jsonschema:"deprecated=true"`
}

// Libdirs is the directories a package installs libraries in.
type Libdirs struct {
	// Directory of the C headers. Default: `/usr/include`. Templated.
	Header string `yaml:"header,omitempty" json:"header,omitempty"`
	// Directory of the C archive libraries. Default: `/usr/lib`. Templated.
	CArchive string `yaml:"carchive,omitempty" json:"carchive,omitempty"`
	// Directory of the C shared libraries. Default: `/usr/lib`. Templated.
	CShared string `yaml:"cshared,omitempty" json:"cshared,omitempty"`
}

// NFPMScripts is the maintainer scripts of a package.
type NFPMScripts struct {
	// Path of the script run before the package is installed. Templated.
	PreInstall string `yaml:"preinstall,omitempty" json:"preinstall,omitempty"`
	// Path of the script run after the package is installed. Templated.
	PostInstall string `yaml:"postinstall,omitempty" json:"postinstall,omitempty"`
	// Path of the script run before the package is removed. Templated.
	PreRemove string `yaml:"preremove,omitempty" json:"preremove,omitempty"`
	// Path of the script run after the package is removed. Templated.
	PostRemove string `yaml:"postremove,omitempty" json:"postremove,omitempty"`
}

// NFPMRPMSignature is the signing of an RPM package.
type NFPMRPMSignature struct {
	// Path of the PGP secret key, which may be ASCII-armored; the package is
	// signed when it is set. Templated.
	KeyFile       string `yaml:"key_file,omitempty" json:"key_file,omitempty"`
	KeyPassphrase string `yaml:"-" json:"-"` // populated from environment variable
}

// NFPMRPMScripts is the scripts of an RPM package's transactions.
type NFPMRPMScripts struct {
	// Path of the script run before all RPM transactions.
	PreTrans string `yaml:"pretrans,omitempty" json:"pretrans,omitempty"`
	// Path of the script run after all RPM transactions.
	PostTrans string `yaml:"posttrans,omitempty" json:"posttrans,omitempty"`
}

// NFPMRPM is the settings of an RPM package.
type NFPMRPM struct {
	// Summary of the package. Default: the first line of `description`.
	Summary string `yaml:"summary,omitempty" json:"summary,omitempty"`
	// Group of the package.
	Group string `yaml:"group,omitempty" json:"group,omitempty"`
	// Compression algorithm of the package. Default: `gzip`.
	Compression string `yaml:"compression,omitempty" json:"compression,omitempty"`
	// Signing of the package.
	Signature NFPMRPMSignature `yaml:"signature,omitempty" json:"signature,omitempty"`
	// Scripts run in the package's transactions.
	Scripts NFPMRPMScripts `yaml:"scripts,omitempty" json:"scripts,omitempty"`
	// Prefixes of a relocatable package.
	Prefixes []string `yaml:"prefixes,omitempty" json:"prefixes,omitempty"`
	// Organization that packaged the software. Default: `maintainer`.
	Packager string `yaml:"packager,omitempty" json:"packager,omitempty"`
	// Host name of the build machine. Default: the host name of this
	// machine.
	BuildHost string `yaml:"buildhost,omitempty" json:"buildhost,omitempty"`
}

// NFPMDebScripts is the special files of a deb package.
type NFPMDebScripts struct {
	// Path of the rules script.
	Rules string `yaml:"rules,omitempty" json:"rules,omitempty"`
	// Path of the debconf templates file.
	Templates string `yaml:"templates,omitempty" json:"templates,omitempty"`
	// Path of the debconf config script.
	Config string `yaml:"config,omitempty" json:"config,omitempty"`
}

// NFPMDebTriggers is the dpkg triggers of a deb package.
type NFPMDebTriggers struct {
	// Triggers activated by other packages that the package is interested in.
	Interest []string `yaml:"interest,omitempty" json:"interest,omitempty"`
	// Triggers the package is interested in, awaiting their processing.
	InterestAwait []string `yaml:"interest_await,omitempty" json:"interest_await,omitempty"`
	// Triggers the package is interested in, without awaiting their
	// processing.
	InterestNoAwait []string `yaml:"interest_noawait,omitempty" json:"interest_noawait,omitempty"`
	// Triggers the package activates for other packages.
	Activate []string `yaml:"activate,omitempty" json:"activate,omitempty"`
	// Triggers the package activates, awaiting their processing.
	ActivateAwait []string `yaml:"activate_await,omitempty" json:"activate_await,omitempty"`
	// Triggers the package activates, without awaiting their processing.
	ActivateNoAwait []string `yaml:"activate_noawait,omitempty" json:"activate_noawait,omitempty"`
}

// NFPMDebSignature is the signing of a deb package.
type NFPMDebSignature struct {
	// Path of the PGP secret key, which may be ASCII-armored; the package is
	// signed when it is set. Templated.
	KeyFile       string `yaml:"key_file,omitempty" json:"key_file,omitempty"`
	KeyPassphrase string `yaml:"-" json:"-"` // populated from environment variable
	// Role of the signer. Default: `origin`.
	Type string `yaml:"type,omitempty" json:"type,omitempty"`
}

// NFPMDeb is the settings of a deb package.
type NFPMDeb struct {
	// Special files of the package.
	Scripts NFPMDebScripts `yaml:"scripts,omitempty" json:"scripts,omitempty"`
	// Triggers of the package.
	Triggers NFPMDebTriggers `yaml:"triggers,omitempty" json:"triggers,omitempty"`
	// Packages that break when the package is installed; installing it is
	// blocked while any of them is installed.
	Breaks []string `yaml:"breaks,omitempty" json:"breaks,omitempty"`
	// Signing of the package.
	Signature NFPMDebSignature `yaml:"signature,omitempty" json:"signature,omitempty"`
	// Lintian tags overridden for the package.
	Lintian []string `yaml:"lintian_overrides,omitempty" json:"lintian_overrides,omitempty"`
	// Compression algorithm of the package data. Default: `gzip`.
	Compression string `yaml:"compression,omitempty" json:"compression,omitempty" jsonschema:"enum=gzip,enum=xz,enum=zstd,enum=none,default=gzip"`
	// Additional fields of the control file; empty fields are left out.
	Fields map[string]string `yaml:"fields,omitempty" json:"fields,omitempty"`
	// Packages installed completely, scripts included, before the package.
	Predepends []string `yaml:"predepends,omitempty" json:"predepends,omitempty"`
}

// NFPMAPKScripts is the upgrade scripts of an apk package.
type NFPMAPKScripts struct {
	// Path of the script run before the package is upgraded.
	PreUpgrade string `yaml:"preupgrade,omitempty" json:"preupgrade,omitempty"`
	// Path of the script run after the package is upgraded.
	PostUpgrade string `yaml:"postupgrade,omitempty" json:"postupgrade,omitempty"`
}

// NFPMAPKSignature is the signing of an apk package.
type NFPMAPKSignature struct {
	// Path of the RSA private key, in PEM format; the package is signed when
	// it is set. Templated.
	KeyFile       string `yaml:"key_file,omitempty" json:"key_file,omitempty"`
	KeyPassphrase string `yaml:"-" json:"-"` // populated from environment variable
	// Name of the signing key. Default: `<maintainer email>.rsa.pub`.
	// Templated.
	KeyName string `yaml:"key_name,omitempty" json:"key_name,omitempty"`
}

// NFPMAPK is the settings of an apk package.
type NFPMAPK struct {
	// Upgrade scripts of the package.
	Scripts NFPMAPKScripts `yaml:"scripts,omitempty" json:"scripts,omitempty"`
	// Signing of the package.
	Signature NFPMAPKSignature `yaml:"signature,omitempty" json:"signature,omitempty"`
}

// NFPMArchLinuxScripts is the upgrade scripts of an Arch Linux package.
type NFPMArchLinuxScripts struct {
	// Path of the script run before the package is upgraded.
	PreUpgrade string `yaml:"preupgrade,omitempty" json:"preupgrade,omitempty"`
	// Path of the script run after the package is upgraded.
	PostUpgrade string `yaml:"postupgrade,omitempty" json:"postupgrade,omitempty"`
}

// NFPMArchLinux is the settings of an Arch Linux package.
type NFPMArchLinux struct {
	// Name that refers to the group of packages the package belongs to.
	Pkgbase string `yaml:"pkgbase,omitempty" json:"pkgbase,omitempty"`
	// Organization that packaged the software, as against its maintainer.
	Packager string `yaml:"packager,omitempty" json:"packager,omitempty"`
	// Upgrade scripts of the package.
	Scripts NFPMArchLinuxScripts `yaml:"scripts,omitempty" json:"scripts,omitempty"`
}

// NFPMIPKAlternative is an alternative name of a file in an ipk package,
// created as a symlink.
type NFPMIPKAlternative struct {
	// Priority of the alternative link.
	Priority int `yaml:"priority,omitempty" json:"priority,omitempty"`
	// Path the alternative link points to.
	Target string `yaml:"target,omitempty" json:"target,omitempty"`
	// Path of the alternative link.
	LinkName string `yaml:"link_name,omitempty" json:"link_name,omitempty"`
}

// NFPMIPK is the settings of an ipk package.
type NFPMIPK struct {
	// ABI version of the package.
	ABIVersion string `yaml:"abi_version,omitempty" json:"abi_version,omitempty"`
	// Alternative names of files, created as symlinks.
	Alternatives []NFPMIPKAlternative `yaml:"alternatives,omitempty" json:"alternatives,omitempty"`
	// Whether to mark the package as auto-installed.
	AutoInstalled bool `yaml:"auto_installed,omitempty" json:"auto_installed,omitempty"`
	// Whether to mark the package as essential.
	Essential bool `yaml:"essential,omitempty" json:"essential,omitempty"`
	// Packages installed completely, scripts included, before the package.
	Predepends []string `yaml:"predepends,omitempty" json:"predepends,omitempty"`
	// Tags of the package.
	Tags []string `yaml:"tags,omitempty" json:"tags,omitempty"`
	// Additional fields of the control file; empty fields are left out.
	Fields map[string]string `yaml:"fields,omitempty" json:"fields,omitempty"`
}

// NFPMMSIXIdentity is the identity of an MSIX package.
type NFPMMSIXIdentity struct {
	// Resource identifier of the package.
	ResourceID string `yaml:"resource_id,omitempty" json:"resource_id,omitempty"`
}

// NFPMMSIXProperties is the display properties of an MSIX package.
type NFPMMSIXProperties struct {
	// Name of the package shown to users. Default: the package name.
	// Templated.
	DisplayName string `yaml:"display_name,omitempty" json:"display_name,omitempty"`
	// Publisher name shown to users. Default: the package name. Templated.
	PublisherDisplayName string `yaml:"publisher_display_name,omitempty" json:"publisher_display_name,omitempty"`
	// Path of the package logo. Required. Templated.
	Logo string `yaml:"logo,omitempty" json:"logo,omitempty"`
}

// NFPMMSIXVisualElements is the visual presentation of an application in an
// MSIX package.
type NFPMMSIXVisualElements struct {
	// Name of the application shown to users. Default: the package name.
	DisplayName string `yaml:"display_name,omitempty" json:"display_name,omitempty"`
	// Description of the application. Default: the package description.
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	// Background color of the application's tile. Default: `transparent`.
	BackgroundColor string `yaml:"background_color,omitempty" json:"background_color,omitempty"`
	// Path of the 150x150 logo. Default: the package logo.
	Square150x150Logo string `yaml:"square150x150_logo,omitempty" json:"square150x150_logo,omitempty"`
	// Path of the 44x44 logo. Default: the package logo.
	Square44x44Logo string `yaml:"square44x44_logo,omitempty" json:"square44x44_logo,omitempty"`
}

// NFPMMSIXApplication is an application declared in an MSIX package.
type NFPMMSIXApplication struct {
	// ID of the application. Required.
	ID string `yaml:"id" json:"id"`
	// Path of the application's executable in the package. Required.
	Executable string `yaml:"executable" json:"executable"`
	// Entry point of the application. Default:
	// `Windows.FullTrustApplication`.
	EntryPoint string `yaml:"entry_point,omitempty" json:"entry_point,omitempty"`
	// Visual presentation of the application.
	VisualElements NFPMMSIXVisualElements `yaml:"visual_elements,omitempty" json:"visual_elements,omitempty"`
}

// NFPMMSIXTargetDeviceFamily is a device family an MSIX package targets.
type NFPMMSIXTargetDeviceFamily struct {
	// Name of the device family. Required.
	Name string `yaml:"name" json:"name"`
	// Minimum version of the device family. Required.
	MinVersion string `yaml:"min_version" json:"min_version"`
	// Highest version of the device family the package was tested on.
	// Required.
	MaxVersionTested string `yaml:"max_version_tested" json:"max_version_tested"`
}

// NFPMMSIXDependencies is the dependencies of an MSIX package.
type NFPMMSIXDependencies struct {
	// Device families the package targets. Default: `Windows.Desktop`, from
	// `10.0.17763.0` to `10.0.22621.0`.
	TargetDeviceFamilies []NFPMMSIXTargetDeviceFamily `yaml:"target_device_families,omitempty" json:"target_device_families,omitempty"`
}

// NFPMMSIXCapabilities is the capabilities an MSIX package declares.
type NFPMMSIXCapabilities struct {
	// General capabilities of the package.
	Capabilities []string `yaml:"capabilities,omitempty" json:"capabilities,omitempty"`
	// Device capabilities of the package.
	DeviceCapabilities []string `yaml:"device_capabilities,omitempty" json:"device_capabilities,omitempty"`
	// Restricted capabilities of the package; `runFullTrust` is added for
	// full trust applications.
	Restricted []string `yaml:"restricted,omitempty" json:"restricted,omitempty"`
}

// NFPMMSIXSignature is the signing of an MSIX package.
type NFPMMSIXSignature struct {
	// Path of the PFX certificate; the package is signed when it is set.
	// Templated.
	PFXFile       string `yaml:"pfx_file,omitempty" json:"pfx_file,omitempty"`
	KeyPassphrase string `yaml:"-" json:"-"` // populated from environment variable
}

// NFPMMSIX is the settings of an MSIX package.
type NFPMMSIX struct {
	// Architecture of the package, in MSIX naming. Default: derived from the
	// binary's `GOARCH`.
	Arch string `yaml:"arch,omitempty" json:"arch,omitempty"`
	// Publisher identity of the package, matching the subject of the signing
	// certificate. Required. Templated.
	Publisher string `yaml:"publisher,omitempty" json:"publisher,omitempty"`
	// Identity of the package.
	Identity NFPMMSIXIdentity `yaml:"identity,omitempty" json:"identity,omitempty"`
	// Display properties of the package.
	Properties NFPMMSIXProperties `yaml:"properties,omitempty" json:"properties,omitempty"`
	// Applications declared in the package; at least one is required.
	Applications []NFPMMSIXApplication `yaml:"applications,omitempty" json:"applications,omitempty"`
	// Dependencies of the package.
	Dependencies NFPMMSIXDependencies `yaml:"dependencies,omitempty" json:"dependencies,omitempty"`
	// Capabilities the package declares.
	Capabilities NFPMMSIXCapabilities `yaml:"capabilities,omitempty" json:"capabilities,omitempty"`
	// Signing of the package.
	Signature NFPMMSIXSignature `yaml:"signature,omitempty" json:"signature,omitempty"`
}

// NFPMOverridables is the package settings a format can override.
type NFPMOverridables struct {
	// Name of the package file. Default:
	// `{{ .PackageName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}` and the
	// target's variant. Templated.
	FileNameTemplate string `yaml:"file_name_template,omitempty" json:"file_name_template,omitempty"`
	// Name of the package. Default: the project name. Templated.
	PackageName string `yaml:"package_name,omitempty" json:"package_name,omitempty"`
	// Epoch of the package version. Default: taken from the version when it
	// is semantic.
	Epoch string `yaml:"epoch,omitempty" json:"epoch,omitempty"`
	// Release of the package version.
	Release string `yaml:"release,omitempty" json:"release,omitempty"`
	// Prerelease part of the package version. Default: taken from the version
	// when it is semantic.
	Prerelease string `yaml:"prerelease,omitempty" json:"prerelease,omitempty"`
	// Metadata part of the package version. Default: taken from the version
	// when it is semantic.
	VersionMetadata string `yaml:"version_metadata,omitempty" json:"version_metadata,omitempty"`
	// Packages the package depends on.
	Dependencies []string `yaml:"dependencies,omitempty" json:"dependencies,omitempty"`
	// Packages the package recommends installing.
	Recommends []string `yaml:"recommends,omitempty" json:"recommends,omitempty"`
	// Packages the package suggests installing.
	Suggests []string `yaml:"suggests,omitempty" json:"suggests,omitempty"`
	// Packages that conflict with the package.
	Conflicts []string `yaml:"conflicts,omitempty" json:"conflicts,omitempty"`
	// Umask applied to files with no mode of their own. Default: `0o002`.
	Umask fs.FileMode `yaml:"umask,omitempty" json:"umask,omitempty" jsonschema:"oneof_type=string;integer"`
	// Packages the package replaces.
	Replaces []string `yaml:"replaces,omitempty" json:"replaces,omitempty"`
	// Packages the package provides.
	Provides []string `yaml:"provides,omitempty" json:"provides,omitempty"`
	// Files added to the package besides its binaries.
	Contents []NFPMContent `yaml:"contents,omitempty" json:"contents,omitempty"`
	// Maintainer scripts of the package.
	Scripts NFPMScripts `yaml:"scripts,omitempty" json:"scripts,omitempty"`
	// Settings of the `rpm` format.
	RPM NFPMRPM `yaml:"rpm,omitempty" json:"rpm,omitempty"`
	// Settings of the `deb` format.
	Deb NFPMDeb `yaml:"deb,omitempty" json:"deb,omitempty"`
	// Settings of the `apk` format.
	APK NFPMAPK `yaml:"apk,omitempty" json:"apk,omitempty"`
	// Settings of the `archlinux` format.
	ArchLinux NFPMArchLinux `yaml:"archlinux,omitempty" json:"archlinux,omitempty"`
	// Settings of the `ipk` format.
	IPK NFPMIPK `yaml:"ipk,omitempty" json:"ipk,omitempty"`
	// Settings of the `msix` format.
	MSIX NFPMMSIX `yaml:"msix,omitempty" json:"msix,omitempty"`
}

// NFPMContent is a file, directory or link added to a package.
type NFPMContent struct {
	// Path or glob of the source. Templated.
	Source string `yaml:"src,omitempty" json:"src,omitempty"`
	// Path of the entry in the package; a trailing `/` copies `src` into it.
	// Required. Templated.
	Destination string `yaml:"dst" json:"dst"`
	// Type of the entry; empty means a regular file.
	Type string `yaml:"type,omitempty" json:"type,omitempty" jsonschema:"enum=symlink,enum=ghost,enum=config,enum=config|noreplace,enum=dir,enum=tree,enum=,default="`
	// Format the entry is limited to. Empty means all.
	Packager string `yaml:"packager,omitempty" json:"packager,omitempty"`
	// File info set on the entry.
	FileInfo FileInfo `yaml:"file_info,omitempty" json:"file_info,omitempty"`
	// Has no effect.
	Expand bool `yaml:"expand,omitempty" json:"expand,omitempty"`
}

// SRPM is a source RPM package of the project's source archive.
type SRPM struct {
	// One-line summary of the package.
	Summary string `yaml:"summary,omitempty" json:"summary,omitempty"`
	// Group of the package.
	Group string `yaml:"group,omitempty" json:"group,omitempty"`
	// Compression algorithm of the package.
	Compression string `yaml:"compression,omitempty" json:"compression,omitempty"`
	// Signing of the package.
	Signature NFPMRPMSignature `yaml:"signature,omitempty" json:"signature,omitempty"`
	// Whether to build the package.
	Enabled bool `yaml:"enabled,omitempty" json:"enabled,omitempty"`
	// Name of the package. Default: the project name.
	PackageName string `yaml:"package_name,omitempty" json:"package_name,omitempty"`
	// Epoch of the package version.
	Epoch string `yaml:"epoch,omitempty" json:"epoch,omitempty"`
	// Go import path of the project.
	ImportPath string `yaml:"import_path,omitempty" json:"import_path,omitempty"`
	// Section of the package.
	Section string `yaml:"section,omitempty" json:"section,omitempty"`
	// Maintainer of the package.
	Maintainer string `yaml:"maintainer,omitempty" json:"maintainer,omitempty"`
	// Name of the package file; `.src.rpm` is appended when missing.
	// Default: `{{ .PackageName }}-{{ .Version }}.src.rpm`. Templated.
	FileNameTemplate string `yaml:"file_name_template,omitempty" json:"file_name_template,omitempty"`
	// Path of the `.spec` file, a template. Required.
	SpecFile string `yaml:"spec_file,omitempty" json:"spec_file,omitempty"`
	// License of the package.
	License string `yaml:"license,omitempty" json:"license,omitempty"`
	// Name of the license file.
	LicenseFileName string `yaml:"license_file_name,omitempty" json:"license_file_name,omitempty"`
	// Vendor of the package.
	Vendor string `yaml:"vendor,omitempty" json:"vendor,omitempty"`
	// Homepage of the project.
	URL string `yaml:"url,omitempty" json:"url,omitempty"`
	// Person or organization that built the package.
	Packager string `yaml:"packager,omitempty" json:"packager,omitempty"`
	// Description of the package.
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	// Go import paths of the binaries, keyed by binary name. Default: the
	// project name mapped to `%{goipath}`.
	Bins map[string]string `yaml:"bins,omitempty" json:"bins,omitempty"`
	// Documentation files of the package.
	Docs []string `yaml:"docs,omitempty" json:"docs,omitempty"`
	// Files added to the package besides the source archive and the spec
	// file.
	Contents []NFPMContent `yaml:"contents,omitempty" json:"contents,omitempty"`
}

// SBOM is a software bill of materials of a set of artifacts, written by a
// command.
type SBOM struct {
	// ID of the SBOM, unique among the SBOMs. Default: `default`.
	ID string `yaml:"id,omitempty" json:"id,omitempty"`
	// Command that writes the SBOM, run in `dist`. Default: `syft`.
	Cmd string `yaml:"cmd,omitempty" json:"cmd,omitempty"`
	// Environment variables of the command, in the format `KEY=value`.
	// Default: `SYFT_FILE_METADATA_CATALOGER_ENABLED=true` when `cmd` is
	// `syft` and `artifacts` is `source` or `archive`. Templated.
	Env []string `yaml:"env,omitempty" json:"env,omitempty"`
	// Arguments of the command. Default when `cmd` is `syft`:
	// `["$artifact", "--output", "spdx-json=$document", "--enrich", "all"]`.
	// Templated.
	Args []string `yaml:"args,omitempty" json:"args,omitempty"`
	// Paths of the SBOM documents, relative to `dist`; more than one only
	// with `artifacts: any`. Default: `{{ .ArtifactName }}.sbom.json`; for
	// `binary`, `{{ .Binary }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}` and
	// the target's variant, then `.sbom.json`; none for `any`. Templated.
	Documents []string `yaml:"documents,omitempty" json:"documents,omitempty"`
	// Kind of artifacts cataloged; `any` leaves the choice to the command.
	// Default: `archive`.
	Artifacts string `yaml:"artifacts,omitempty" json:"artifacts,omitempty" jsonschema:"enum=source,enum=package,enum=archive,enum=binary,enum=any,enum=none,default=archive"`
	// IDs of the artifacts to catalog. Empty means all. Has no effect for
	// `source` and `any`.
	IDs []string `yaml:"ids,omitempty" json:"ids,omitempty"`

	// Whether to skip this SBOM. Templated.
	Disable string `yaml:"disable,omitempty" json:"disable,omitempty" jsonschema:"oneof_type=string;boolean"`
}

// Sign is a signing of a set of artifacts by a command.
type Sign struct {
	// ID of the signing, unique among the signings. Default: `default`.
	ID string `yaml:"id,omitempty" json:"id,omitempty"`
	// Signing command. Default: Git's `gpg.program`, else `gpg`, for
	// `signs`; `cosign` for `docker_signs`.
	Cmd string `yaml:"cmd,omitempty" json:"cmd,omitempty"`
	// Arguments of the command. Default for `signs`:
	// `["--output", "$signature", "--detach-sig", "$artifact"]`. Default for
	// `docker_signs`:
	// `["sign", "--key=cosign.key", "${artifact}@${digest}", "--yes"]`.
	// Templated.
	Args []string `yaml:"args,omitempty" json:"args,omitempty"`
	// Name of the signature file. Default for `signs`: `${artifact}.sig`.
	// Templated.
	Signature string `yaml:"signature,omitempty" json:"signature,omitempty"`
	// Kind of artifacts signed. Default: `none` for `signs`; for
	// `docker_signs`, the images of `dockers_v2`.
	Artifacts string `yaml:"artifacts,omitempty" json:"artifacts,omitempty" jsonschema:"enum=all,enum=manifests,enum=images,enum=checksum,enum=source,enum=package,enum=archive,enum=binary,enum=sbom,enum=none"`
	// IDs of the artifacts to sign. Empty means all. Has no effect for
	// `checksum` and `source`.
	IDs []string `yaml:"ids,omitempty" json:"ids,omitempty"`
	// Data given to the command on its standard input. Templated.
	Stdin *string `yaml:"stdin,omitempty" json:"stdin,omitempty"`
	// Path of a file given to the command on its standard input.
	StdinFile string `yaml:"stdin_file,omitempty" json:"stdin_file,omitempty"`
	// Environment variables of the command, in the format `KEY=value`.
	// Templated.
	Env []string `yaml:"env,omitempty" json:"env,omitempty"`
	// Name of the certificate file the command writes, available to `args`
	// as `${certificate}`. Templated.
	Certificate string `yaml:"certificate,omitempty" json:"certificate,omitempty"`
	// Whether to always print the command's output, otherwise shown only in
	// verbose mode. Templated.
	Output string `yaml:"output,omitempty" json:"output,omitempty" jsonschema:"oneof_type=string;boolean"`
}

// BinarySign is a signing of a set of binaries by a command.
type BinarySign struct {
	// ID of the signing, unique among the signings. Default: `default`.
	ID string `yaml:"id,omitempty" json:"id,omitempty"`
	// Signing command. Default: Git's `gpg.program`, else `gpg`.
	Cmd string `yaml:"cmd,omitempty" json:"cmd,omitempty"`
	// Arguments of the command. Default:
	// `["--output", "$signature", "--detach-sig", "$artifact"]`. Templated.
	Args []string `yaml:"args,omitempty" json:"args,omitempty"`
	// Name of the signature file. Default:
	// `${artifact}_{{ .Os }}_{{ .Arch }}{{ targetVariant . }}`. Templated.
	Signature string `yaml:"signature,omitempty" json:"signature,omitempty"`
	// Kind of artifacts signed. Default: `binary`.
	Artifacts string `yaml:"artifacts,omitempty" json:"artifacts,omitempty" jsonschema:"enum=binary,enum=none"`
	// IDs of the builds whose binaries are signed, a universal binary's `id`
	// for a universal binary. Empty means all.
	IDs []string `yaml:"ids,omitempty" json:"ids,omitempty"`
	// Data given to the command on its standard input. Templated.
	Stdin *string `yaml:"stdin,omitempty" json:"stdin,omitempty"`
	// Path of a file given to the command on its standard input.
	StdinFile string `yaml:"stdin_file,omitempty" json:"stdin_file,omitempty"`
	// Environment variables of the command, in the format `KEY=value`.
	// Templated.
	Env []string `yaml:"env,omitempty" json:"env,omitempty"`
	// Name of the certificate file the command writes, available to `args`
	// as `${certificate}`. Templated.
	Certificate string `yaml:"certificate,omitempty" json:"certificate,omitempty"`
	// Whether to always print the command's output, otherwise shown only in
	// verbose mode. Templated.
	Output string `yaml:"output,omitempty" json:"output,omitempty" jsonschema:"oneof_type=string;boolean"`
}

// Notarize is the signing and notarization of macOS binaries.
type Notarize struct {
	// Signings and notarizations of darwin binaries, done with a P12
	// certificate and an App Store Connect API key.
	MacOS []MacOSSignNotarize `yaml:"macos" json:"macos"`
}

// MacOSSignNotarize is the signing and notarization of a set of darwin
// binaries.
type MacOSSignNotarize struct {
	// IDs of the builds whose binaries are signed. Default: the project name.
	IDs []string `yaml:"ids,omitempty" json:"ids,omitempty"`
	// Whether to sign and notarize the binaries. Templated.
	Enabled string `yaml:"enabled,omitempty" json:"enabled,omitempty" jsonschema:"oneof_type=string;boolean"`
	// Signing of the binaries.
	Sign MacOSSign `yaml:"sign" json:"sign"`
	// Notarization of the signed binaries; they are only signed when its
	// `issuer_id`, `key_id` or `key` is empty.
	Notarize MacOSNotarize `yaml:"notarize" json:"notarize"`
}

// MacOSNotarize is the notarization of signed darwin binaries.
type MacOSNotarize struct {
	// Issuer ID of the App Store Connect API key, a UUID. Templated.
	IssuerID string `yaml:"issuer_id" json:"issuer_id"`
	// Path of the App Store Connect API key, a `.p8` file, or its
	// base64-encoded contents. Templated.
	Key string `yaml:"key" json:"key"`
	// ID of the App Store Connect API key. Templated.
	KeyID string `yaml:"key_id" json:"key_id"`
	// Time to wait for the notarization when `wait` is set, at most `20m`.
	// Default: `10m`.
	Timeout time.Duration `yaml:"timeout,omitempty" json:"timeout,omitempty" jsonschema:"type=string"`
	// Whether to wait for the notarization to finish.
	Wait bool `yaml:"wait,omitempty" json:"wait,omitempty"`
}

// MacOSSign is the signing of darwin binaries.
type MacOSSign struct {
	// Path of the `.p12` certificate, or its base64-encoded contents.
	// Templated.
	Certificate string `yaml:"certificate" json:"certificate"`
	// Password of the certificate. Templated.
	Password string `yaml:"password" json:"password"`

	// Path of the entitlements XML file of the signature. Templated.
	Entitlements string `yaml:"entitlements,omitempty" json:"entitlements,omitempty"`
}

// SnapcraftAppMetadata is an app of a snap.
type SnapcraftAppMetadata struct {
	// Command of the app. Default: the app's name.
	Command string `yaml:"command" json:"command"`
	// Arguments appended to `command`.
	Args string `yaml:"args,omitempty" json:"args,omitempty"`

	// Kind of wrapper generated for the command.
	Adapter string `yaml:"adapter,omitempty" json:"adapter,omitempty"`
	// Apps ordered to start after this one.
	After []string `yaml:"after,omitempty" json:"after,omitempty"`
	// Aliases of the app's command.
	Aliases []string `yaml:"aliases,omitempty" json:"aliases,omitempty"`
	// Name of the `.desktop` file that starts the app with the desktop
	// session.
	Autostart string `yaml:"autostart,omitempty" json:"autostart,omitempty"`
	// Apps ordered to start before this one.
	Before []string `yaml:"before,omitempty" json:"before,omitempty"`
	// D-Bus name the service is reachable as; required when `daemon` is
	// `dbus`.
	BusName string `yaml:"bus_name,omitempty" json:"bus_name,omitempty"`
	// Commands run in order before the app's command.
	CommandChain []string `yaml:"command_chain,omitempty" json:"command_chain,omitempty"`
	// ID of the app's desktop entry in an external AppStream file.
	CommonID string `yaml:"common_id,omitempty" json:"common_id,omitempty"`
	// Path of the app's Bash completion script, copied into the snap.
	Completer string `yaml:"completer,omitempty" json:"completer,omitempty"`
	// Kind of daemon the app runs as.
	Daemon string `yaml:"daemon,omitempty" json:"daemon,omitempty"`
	// Path of the app's `.desktop` file in the snap.
	Desktop string `yaml:"desktop,omitempty" json:"desktop,omitempty"`
	// Environment variables of the app, keyed by name.
	Environment map[string]any `yaml:"environment,omitempty" json:"environment,omitempty"`
	// Snapcraft extensions the app uses.
	Extensions []string `yaml:"extensions,omitempty" json:"extensions,omitempty"`
	// Whether a freshly installed daemon starts automatically. Requires
	// `daemon`.
	InstallMode string `yaml:"install_mode,omitempty" json:"install_mode,omitempty"`
	// Attributes passed through to the app in `snap.yaml` unvalidated.
	Passthrough map[string]any `yaml:"passthrough,omitempty" json:"passthrough,omitempty"`
	// Plugs of the app, granting it access outside its confinement.
	Plugs []string `yaml:"plugs,omitempty" json:"plugs,omitempty"`
	// Command run from inside the snap after the service stops.
	PostStopCommand string `yaml:"post_stop_command,omitempty" json:"post_stop_command,omitempty"`
	// Whether the daemon is restarted during a snap refresh.
	RefreshMode string `yaml:"refresh_mode,omitempty" json:"refresh_mode,omitempty"`
	// Command that asks the service to reload its configuration. Requires
	// `daemon`.
	ReloadCommand string `yaml:"reload_command,omitempty" json:"reload_command,omitempty"`
	// Condition on which the service restarts.
	RestartCondition string `yaml:"restart_condition,omitempty" json:"restart_condition,omitempty"`
	// Delay before the service restarts, a duration.
	RestartDelay string `yaml:"restart_delay,omitempty" json:"restart_delay,omitempty"`
	// Slots of the app.
	Slots []string `yaml:"slots,omitempty" json:"slots,omitempty"`
	// Sockets of the daemon, keyed by name; requires `network-bind` in
	// `plugs`.
	Sockets map[string]any `yaml:"sockets,omitempty" json:"sockets,omitempty"`
	// Time to wait for the daemon to start, a duration.
	StartTimeout string `yaml:"start_timeout,omitempty" json:"start_timeout,omitempty"`
	// Command that stops the service. Requires `daemon`.
	StopCommand string `yaml:"stop_command,omitempty" json:"stop_command,omitempty"`
	// How the daemon is stopped. Requires `daemon`.
	StopMode string `yaml:"stop_mode,omitempty" json:"stop_mode,omitempty"`
	// Time to wait for the daemon to stop, a duration.
	StopTimeout string `yaml:"stop_timeout,omitempty" json:"stop_timeout,omitempty"`
	// Schedule of the service. Requires `daemon`.
	Timer string `yaml:"timer,omitempty" json:"timer,omitempty"`
	// Watchdog timeout of the service, a duration; requires `daemon-notify`
	// in `plugs`.
	WatchdogTimeout string `yaml:"watchdog_timeout,omitempty" json:"watchdog_timeout,omitempty"`
}

// SnapcraftLayoutMetadata is what a path in a snap's sandbox maps to.
type SnapcraftLayoutMetadata struct {
	// Path the sandbox path is a symbolic link to.
	Symlink string `yaml:"symlink,omitempty" json:"symlink,omitempty"`
	// Directory bind-mounted on the sandbox path.
	Bind string `yaml:"bind,omitempty" json:"bind,omitempty"`
	// File bind-mounted on the sandbox path.
	BindFile string `yaml:"bind_file,omitempty" json:"bind_file,omitempty"`
	// Type of the filesystem mounted on the sandbox path; `tmpfs` mounts a
	// private in-memory filesystem.
	Type string `yaml:"type,omitempty" json:"type,omitempty"`
}

// Snapcraft is a snap of the Linux binaries of builds, one per architecture.
type Snapcraft struct {
	// Name of the snap file, without its extension. Default:
	// `{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}` and the
	// target's variant. Templated.
	NameTemplate string `yaml:"name_template,omitempty" json:"name_template,omitempty"`
	// Whether to publish the snap to the Snapcraft store.
	Publish bool `yaml:"publish,omitempty" json:"publish,omitempty"`
	// ID of the snap, unique among the snaps.
	ID string `yaml:"id,omitempty" json:"id,omitempty"`
	// IDs of the builds whose binaries are packaged. Empty means all.
	IDs []string `yaml:"ids,omitempty" json:"ids,omitempty"`
	// Name of the snap. Default: the project name.
	Name string `yaml:"name,omitempty" json:"name,omitempty"`
	// Title of the snap, shown in graphical store fronts.
	Title string `yaml:"title,omitempty" json:"title,omitempty"`
	// One-line summary of the snap, at most 79 characters. Required.
	// Templated.
	Summary string `yaml:"summary" json:"summary"`
	// Description of the snap. Required. Templated.
	Description string `yaml:"description" json:"description"`
	// Path of the icon of the snap in store fronts.
	Icon string `yaml:"icon,omitempty" json:"icon,omitempty"`
	// Base snap that is the snap's execution environment.
	Base string `yaml:"base,omitempty" json:"base,omitempty"`
	// License of the snap, an SPDX license expression.
	License string `yaml:"license,omitempty" json:"license,omitempty"`
	// Grade of the snap; `devel` limits its release to the `edge` and `beta`
	// channels. Default: `stable`. Templated.
	Grade string `yaml:"grade,omitempty" json:"grade,omitempty" jsonschema:"enum=stable,enum=devel,default=stable"`
	// Store channels the snap is released to. Default: `[edge, beta]` for
	// the `devel` grade, `[edge, beta, candidate, stable]` otherwise.
	// Templated.
	ChannelTemplates []string `yaml:"channel_templates,omitempty" json:"channel_templates,omitempty"`
	// Confinement policy of the snap. Default: `strict`.
	Confinement string `yaml:"confinement,omitempty" json:"confinement,omitempty" jsonschema:"enum=strict,enum=classic,enum=devmode,default=strict"`
	// Features the core must support for the snap to install.
	Assumes []string `yaml:"assumes,omitempty" json:"assumes,omitempty"`
	// Paths of the sandbox mapped to paths of the snap, keyed by sandbox path.
	Layout map[string]SnapcraftLayoutMetadata `yaml:"layout,omitempty" json:"layout,omitempty"`
	// Apps of the snap, keyed by name. Default: one app running the binary,
	// named `name` or after the binary.
	Apps map[string]SnapcraftAppMetadata `yaml:"apps,omitempty" json:"apps,omitempty"`
	// Hooks of the snap and the plugs they use, keyed by hook name.
	Hooks map[string]any `yaml:"hooks,omitempty" json:"hooks,omitempty"`
	// Settings of the snap's plugs, keyed by plug name.
	Plugs map[string]any `yaml:"plugs,omitempty" json:"plugs,omitempty"`
	// Whether to skip this snap. Templated.
	Disable string `yaml:"disable,omitempty" json:"disable,omitempty" jsonschema:"oneof_type=string;boolean"`

	// Extra files added to the snap.
	Files []SnapcraftExtraFiles `yaml:"extra_files,omitempty" json:"extra_files,omitempty"`

	// IDs of the builds whose binaries are packaged.
	//
	// Deprecated: use `ids`.
	Builds []string `yaml:"builds,omitempty" json:"builds,omitempty" jsonschema:"deprecated=true"`
}

// SnapcraftExtraFiles is an extra file added to a snap.
type SnapcraftExtraFiles struct {
	// Path of the file, relative to the project directory. Required.
	Source string `yaml:"source" json:"source"`
	// Path of the file in the snap, relative to its prime directory.
	// Default: `source`.
	Destination string `yaml:"destination,omitempty" json:"destination,omitempty"`
	// Mode of the file. Default: `0644`.
	Mode uint32 `yaml:"mode,omitempty" json:"mode,omitempty"`
}

// Flatpak is a Flatpak bundle of the Linux binaries of builds, one per
// architecture.
type Flatpak struct {
	// ID of the bundle, unique among the bundles.
	ID string `yaml:"id,omitempty" json:"id,omitempty"`
	// IDs of the builds whose binaries are packaged. Empty means all.
	IDs []string `yaml:"ids,omitempty" json:"ids,omitempty"`
	// Name of the bundle file, without its extension. Default:
	// `{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}` and the
	// target's variant. Templated.
	NameTemplate string `yaml:"name_template,omitempty" json:"name_template,omitempty"`

	// Application ID, in reverse-DNS notation. Required.
	AppID string `yaml:"app_id" json:"app_id"`

	// Runtime of the application. Required.
	Runtime string `yaml:"runtime" json:"runtime"`

	// Version of the runtime. Required.
	RuntimeVersion string `yaml:"runtime_version" json:"runtime_version"`

	// SDK of the application, compatible with `runtime`. Required.
	SDK string `yaml:"sdk" json:"sdk"`

	// Command run inside the sandbox. Default: the binary's name.
	Command string `yaml:"command,omitempty" json:"command,omitempty"`

	// Sandbox permissions of the application.
	FinishArgs []string `yaml:"finish_args,omitempty" json:"finish_args,omitempty"`

	// Whether to skip this bundle. Templated.
	Disable string `yaml:"disable,omitempty" json:"disable,omitempty" jsonschema:"oneof_type=string;boolean"`
}

// Snapshot is the settings of a snapshot release.
type Snapshot struct {
	// Version of a snapshot release. Templated.
	//
	// Deprecated: use `version_template`.
	NameTemplate string `yaml:"name_template,omitempty" json:"name_template,omitempty" jsonschema:"deprecated=true"`
	// Version of a snapshot release. Default:
	// `{{ .Version }}-SNAPSHOT-{{ .ShortCommit }}`. Templated.
	VersionTemplate string `yaml:"version_template,omitempty" json:"version_template,omitempty"`
}

// Checksum is the checksums of the release's artifacts.
type Checksum struct {
	// Name of the checksums file. Default:
	// `{{ .ProjectName }}_{{ .Version }}_checksums.txt`, or
	// `{{ .ArtifactName }}.{{ .Algorithm }}` with `split`. Templated.
	NameTemplate string `yaml:"name_template,omitempty" json:"name_template,omitempty"`
	// Hash algorithm of the checksums. Default: `sha256`.
	Algorithm string `yaml:"algorithm,omitempty" json:"algorithm,omitempty"`
	// Whether to write one checksum file per artifact.
	Split bool `yaml:"split,omitempty" json:"split,omitempty"`
	// IDs of the artifacts whose checksums are written. Empty means all.
	IDs []string `yaml:"ids,omitempty" json:"ids,omitempty"`
	// Whether to skip the checksums.
	Disable bool `yaml:"disable,omitempty" json:"disable,omitempty"`
	// Existing files whose checksums are added.
	ExtraFiles []ExtraFile `yaml:"extra_files,omitempty" json:"extra_files,omitempty"`
}

// Retry is the retrying of an operation that fails.
type Retry struct {
	// Maximum number of attempts; `1` disables retries. Default: `10`.
	Attempts uint `yaml:"attempts,omitempty" json:"attempts,omitempty"`
	// Delay between attempts, a duration. Default: `10s`.
	Delay time.Duration `yaml:"delay,omitempty" json:"delay,omitempty" jsonschema:"type=string"`
	// Maximum delay between attempts, a duration. Default: `5m`.
	MaxDelay time.Duration `yaml:"max_delay,omitempty" json:"max_delay,omitempty" jsonschema:"type=string"`
}

// Docker is a Docker image of the binaries and packages of one target.
//
// Deprecated: use `dockers_v2`.
type Docker struct {
	// ID of the image, unique among the images.
	ID string `yaml:"id,omitempty" json:"id,omitempty"`
	// IDs of the binaries and packages copied into the build context; the
	// image is skipped when one is missing. Empty means all.
	IDs []string `yaml:"ids,omitempty" json:"ids,omitempty"`
	// `GOOS` of the binaries and packages used. Default: `linux`.
	Goos string `yaml:"goos,omitempty" json:"goos,omitempty"`
	// `GOARCH` of the binaries and packages used. Default: `amd64`.
	Goarch string `yaml:"goarch,omitempty" json:"goarch,omitempty"`
	// `GOARM` of the binaries and packages used. Default: `6`.
	Goarm string `yaml:"goarm,omitempty" json:"goarm,omitempty" jsonschema:"oneof_type=string;integer"`
	// `GOAMD64` of the binaries and packages used. Default: `v1`.
	Goamd64 string `yaml:"goamd64,omitempty" json:"goamd64,omitempty"`
	// Path of the Dockerfile. Default: `Dockerfile`. Templated.
	Dockerfile string `yaml:"dockerfile,omitempty" json:"dockerfile,omitempty"`
	// Names of the image. Templated.
	ImageTemplates []string `yaml:"image_templates,omitempty" json:"image_templates,omitempty"`
	// Whether to skip pushing the image; `auto` skips it for a prerelease.
	// Templated.
	SkipPush string `yaml:"skip_push,omitempty" json:"skip_push,omitempty" jsonschema:"oneof_type=string;boolean"`
	// Paths of other files copied into the build context, relative to the
	// working directory.
	Files []string `yaml:"extra_files,omitempty" json:"extra_files,omitempty"`
	// Flags of the build command. Templated.
	BuildFlagTemplates []string `yaml:"build_flag_templates,omitempty" json:"build_flag_templates,omitempty"`
	// Extra flags of the push command.
	PushFlags []string `yaml:"push_flags,omitempty" json:"push_flags,omitempty"`
	// Tool that builds the image. Default: `docker`.
	Use string `yaml:"use,omitempty" json:"use,omitempty" jsonschema:"enum=docker,enum=buildx,default=docker"`

	// Retrying of the build and push. Default: the top-level `retry`.
	//
	// Deprecated: use the top-level `retry`.
	Retry Retry `yaml:"retry,omitempty" json:"retry,omitempty"`
}

// DockerManifest is a Docker manifest list of images.
//
// Deprecated: use `dockers_v2`.
type DockerManifest struct {
	// ID of the manifest.
	ID string `yaml:"id,omitempty" json:"id,omitempty"`
	// Name of the manifest. Templated.
	NameTemplate string `yaml:"name_template,omitempty" json:"name_template,omitempty"`
	// Whether to skip the manifest; `auto` skips it for a prerelease.
	// Templated.
	SkipPush string `yaml:"skip_push,omitempty" json:"skip_push,omitempty" jsonschema:"oneof_type=string;boolean"`
	// Names of the images in the manifest. Templated.
	ImageTemplates []string `yaml:"image_templates,omitempty" json:"image_templates,omitempty"`
	// Extra flags of the manifest create command.
	CreateFlags []string `yaml:"create_flags,omitempty" json:"create_flags,omitempty"`
	// Extra flags of the manifest push command.
	PushFlags []string `yaml:"push_flags,omitempty" json:"push_flags,omitempty"`
	// Tool that creates the manifest. Default: `docker`.
	Use string `yaml:"use,omitempty" json:"use,omitempty"`

	// Retrying of the manifest commands. Default: the top-level `retry`.
	//
	// Deprecated: use the top-level `retry`.
	Retry Retry `yaml:"retry,omitempty" json:"retry,omitempty"`
}

// DockerV2 is a multi-platform Docker image of the binaries and packages of
// builds, built and pushed in one step.
type DockerV2 struct {
	// ID of the image, unique among the images. Default: the project name.
	ID string `yaml:"id,omitempty" json:"id,omitempty"`
	// IDs of the binaries and packages copied into the build context. Empty
	// means all.
	IDs []string `yaml:"ids,omitempty" json:"ids,omitempty"`
	// Path of the Dockerfile. Default: `Dockerfile`. Templated.
	Dockerfile string `yaml:"dockerfile,omitempty" json:"dockerfile,omitempty"`
	// Names of the image, without tags; empty names are ignored. Templated.
	Images []string `yaml:"images,omitempty" json:"images,omitempty"`
	// Tags of the image; empty tags are ignored. Default: `{{ .Tag }}`.
	// Templated.
	Tags []string `yaml:"tags,omitempty" json:"tags,omitempty"`
	// Labels of the image; entries with an empty key or value are ignored.
	// Templated.
	Labels map[string]string `yaml:"labels,omitempty" json:"labels,omitempty"`
	// Annotations of the image, a key optionally prefixed with its scopes, as
	// in `index,manifest:<key>`; entries with an empty key or value are
	// ignored. Templated.
	Annotations map[string]string `yaml:"annotations,omitempty" json:"annotations,omitempty"`
	// Paths of other files copied into the build context, relative to the
	// working directory.
	ExtraFiles []string `yaml:"extra_files,omitempty" json:"extra_files,omitempty"`
	// Platforms the image is built for. Default:
	// `[linux/amd64, linux/arm64]`. Templated.
	Platforms []string `yaml:"platforms,omitempty" json:"platforms,omitempty"`
	// Build arguments of the image. Templated.
	BuildArgs map[string]string `yaml:"build_args,omitempty" json:"build_args,omitempty"`
	// Extra flags of the build command, in the format `--flag=value`.
	// Templated.
	Flags []string `yaml:"flags,omitempty" json:"flags,omitempty"`
	// Whether to skip this image. Templated.
	Disable string `yaml:"disable,omitempty" json:"disable,omitempty" jsonschema:"oneof_type=string;boolean"`
	// Whether to create and attach an SBOM to the image. Default: `true`.
	// Templated.
	SBOM string `yaml:"sbom,omitempty" json:"sbom,omitempty" jsonschema:"oneof_type=string;boolean"`
	// Commands run before and after the build.
	Hooks BuildHookConfig `yaml:"hooks,omitempty" json:"hooks,omitempty"`

	// Retrying of the build. Default: the top-level `retry`.
	//
	// Deprecated: use the top-level `retry`.
	Retry Retry `yaml:"retry,omitempty" json:"retry,omitempty"`
}

// DockerDigest is the file of the digests of the pushed Docker images.
type DockerDigest struct {
	// Whether to skip the digest file. Templated.
	Disable string `yaml:"disable,omitempty" json:"disable,omitempty" jsonschema:"oneof_type=string;boolean"`
	// Name of the digest file. Default: `digests.txt`. Templated.
	NameTemplate string `yaml:"name_template,omitempty" json:"name_template,omitempty"`
}

// Filters is the selection of commits by their message.
type Filters struct {
	// Regular expressions a commit's first line must match to be kept;
	// `exclude` is ignored when it is set.
	Include []string `yaml:"include,omitempty" json:"include,omitempty"`
	// Regular expressions whose matching commits are dropped, matched against
	// a commit's first line.
	Exclude []string `yaml:"exclude,omitempty" json:"exclude,omitempty"`
}

// Changelog is the changelog of the release.
type Changelog struct {
	// Selection of the commits listed.
	Filters Filters `yaml:"filters,omitempty" json:"filters,omitempty"`
	// Order of the entries by commit message; empty keeps the order the
	// commits come in.
	Sort string `yaml:"sort,omitempty" json:"sort,omitempty" jsonschema:"enum=asc,enum=desc,enum=,default="`
	// Whether to skip the changelog, which leaves the release notes empty.
	// Templated.
	Disable string `yaml:"disable,omitempty" json:"disable,omitempty" jsonschema:"oneof_type=string;boolean"`
	// Source of the changelog. Default: `git`.
	Use string `yaml:"use,omitempty" json:"use,omitempty" jsonschema:"enum=git,enum=github,enum=github-native,enum=gitlab,enum=gitea,default=git"`
	// Format of an entry. Default: `{{ .SHA }} {{ .Message }}` for `git`;
	// otherwise the SHA, the message and the author. Templated.
	Format string `yaml:"format,omitempty" json:"format,omitempty"`
	// Groups of the entries, matched by commit message.
	Groups []ChangelogGroup `yaml:"groups,omitempty" json:"groups,omitempty"`
	// Maximum length of the commit hashes; `0` keeps them whole and `-1`
	// drops them.
	Abbrev int `yaml:"abbrev,omitempty" json:"abbrev,omitempty"`
}

// ChangelogGroup is a group of changelog entries.
type ChangelogGroup struct {
	// Title of the group. Required.
	Title string `yaml:"title" json:"title"`
	// Regular expression matched against a commit's first line; empty
	// matches all commits.
	Regexp string `yaml:"regexp,omitempty" json:"regexp,omitempty"`
	// Position of the group among the groups.
	Order int `yaml:"order,omitempty" json:"order,omitempty"`
}

// EnvFiles is the files that hold the SCM tokens, read when the matching
// environment variable is unset.
type EnvFiles struct {
	// Path of the file holding the GitHub token. Default:
	// `~/.config/go-publisher/github_token`.
	GitHubToken string `yaml:"github_token,omitempty" json:"github_token,omitempty"`
	// Path of the file holding the GitLab token. Default:
	// `~/.config/go-publisher/gitlab_token`.
	GitLabToken string `yaml:"gitlab_token,omitempty" json:"gitlab_token,omitempty"`
	// Path of the file holding the Gitea token. Default:
	// `~/.config/go-publisher/gitea_token`.
	GiteaToken string `yaml:"gitea_token,omitempty" json:"gitea_token,omitempty"`
}

// Before is the commands run before the build.
type Before struct {
	// Commands run before the build, in order. Templated.
	Hooks []string `yaml:"hooks,omitempty" json:"hooks,omitempty"`
}

// Blob is an upload of the release's artifacts to a storage bucket.
type Blob struct {
	// Name of the bucket. Required. Templated.
	Bucket string `yaml:"bucket,omitempty" json:"bucket,omitempty"`
	// Cloud provider of the bucket. Required. Templated.
	Provider string `yaml:"provider,omitempty" json:"provider,omitempty"`
	// Region of the bucket. S3 only. Templated.
	Region string `yaml:"region,omitempty" json:"region,omitempty"`
	// Whether to disable SSL. S3 only.
	DisableSSL bool `yaml:"disable_ssl,omitempty" json:"disable_ssl,omitempty"`
	// Directory in the bucket. Default: `{{ .ProjectName }}/{{ .Tag }}`.
	// Templated.
	Directory string `yaml:"directory,omitempty" json:"directory,omitempty"`
	// Secrets keeper URL of the key that encrypts each file before upload,
	// as in `awskms://alias/<name>`.
	KMSKey string `yaml:"kms_key,omitempty" json:"kms_key,omitempty"`
	// IDs of the artifacts to upload. Empty means all.
	IDs []string `yaml:"ids,omitempty" json:"ids,omitempty"`
	// Endpoint of an S3-compatible backend; sets path-style addressing.
	// S3 only. Templated.
	Endpoint string `yaml:"endpoint,omitempty" json:"endpoint,omitempty"`
	// Existing files uploaded besides the artifacts.
	ExtraFiles []ExtraFile `yaml:"extra_files,omitempty" json:"extra_files,omitempty"`
	// Whether to skip this upload. Templated.
	Disable string `yaml:"disable,omitempty" json:"disable,omitempty" jsonschema:"oneof_type=string;boolean"`
	// Whether to use path-style addressing. S3 only. Default: `true` when
	// `endpoint` is set.
	S3ForcePathStyle *bool `yaml:"s3_force_path_style,omitempty" json:"s3_force_path_style,omitempty"`
	// Canned ACL applied to the uploaded files. S3 only.
	ACL string `yaml:"acl,omitempty" json:"acl,omitempty"`
	// Cache control directives of the uploaded files.
	CacheControl []string `yaml:"cache_control,omitempty" json:"cache_control,omitempty"`
	// Content disposition of the uploaded files; `-` sets none. Default:
	// `attachment;filename={{.Filename}}`. Templated.
	ContentDisposition string `yaml:"content_disposition,omitempty" json:"content_disposition,omitempty"`
	// Whether to also upload `metadata.json` and `artifacts.json`.
	IncludeMeta bool `yaml:"include_meta,omitempty" json:"include_meta,omitempty"`
	// Whether to upload only `extra_files`.
	ExtraFilesOnly bool `yaml:"extra_files_only,omitempty" json:"extra_files_only,omitempty"`
}

// Upload is an upload of the release's artifacts to an HTTP server.
type Upload struct {
	// Name of the upload, unique among the uploads. Required.
	Name string `yaml:"name,omitempty" json:"name,omitempty"`
	// IDs of the artifacts to upload. Empty means all.
	IDs []string `yaml:"ids,omitempty" json:"ids,omitempty"`
	// File extensions of the artifacts to upload. Empty means all.
	Exts []string `yaml:"exts,omitempty" json:"exts,omitempty"`
	// URL the artifacts are uploaded to, without the artifact's name unless
	// `custom_artifact_name` is set. Required. Templated.
	Target string `yaml:"target,omitempty" json:"target,omitempty"`
	// Username of the basic authentication. Default: the
	// `UPLOAD_<NAME>_USERNAME` environment variable, or
	// `ARTIFACTORY_<NAME>_USERNAME` under `artifactories`. Templated.
	Username string `yaml:"username,omitempty" json:"username,omitempty"`
	// Kind of artifacts uploaded. Default: `archive`.
	Mode string `yaml:"mode,omitempty" json:"mode,omitempty" jsonschema:"enum=binary,enum=archive,default=archive"`
	// HTTP method of the upload request. Default: `PUT`.
	Method string `yaml:"method,omitempty" json:"method,omitempty"`
	// Name of the request header that carries the artifact's SHA256 checksum.
	ChecksumHeader string `yaml:"checksum_header,omitempty" json:"checksum_header,omitempty"`
	// Path of the PEM client certificate presented to the server; requires
	// `client_x509_key`.
	ClientX509Cert string `yaml:"client_x509_cert,omitempty" json:"client_x509_cert,omitempty"`
	// Path of the PEM key of `client_x509_cert`.
	ClientX509Key string `yaml:"client_x509_key,omitempty" json:"client_x509_key,omitempty"`
	// PEM certificate chain the server's certificate is validated against.
	TrustedCerts string `yaml:"trusted_certificates,omitempty" json:"trusted_certificates,omitempty"`
	// Whether to also upload the checksums.
	Checksum bool `yaml:"checksum,omitempty" json:"checksum,omitempty"`
	// Whether to also upload the signatures and certificates.
	Signature bool `yaml:"signature,omitempty" json:"signature,omitempty"`
	// Whether to also upload `metadata.json`.
	Meta bool `yaml:"meta,omitempty" json:"meta,omitempty"`
	// Whether `target` holds the artifact's name, which is otherwise appended
	// to it.
	CustomArtifactName bool `yaml:"custom_artifact_name,omitempty" json:"custom_artifact_name,omitempty"`
	// Extra headers of the upload request. Templated.
	CustomHeaders map[string]string `yaml:"custom_headers,omitempty" json:"custom_headers,omitempty"`
	// Existing files uploaded besides the artifacts.
	ExtraFiles []ExtraFile `yaml:"extra_files,omitempty" json:"extra_files,omitempty"`
	// Whether to upload only `extra_files`.
	ExtraFilesOnly bool `yaml:"extra_files_only,omitempty" json:"extra_files_only,omitempty"`
	// Whether to skip this upload. Templated.
	Skip string `yaml:"skip,omitempty" json:"skip,omitempty" jsonschema:"oneof_type=string;boolean"`
	// Password of the basic authentication. Default: the
	// `UPLOAD_<NAME>_SECRET` environment variable, or
	// `ARTIFACTORY_<NAME>_SECRET` under `artifactories`. Templated.
	Password string `yaml:"password,omitempty" json:"password,omitempty"`
}

// Publisher is a command run on each of a set of the release's artifacts.
type Publisher struct {
	// Name of the publisher.
	Name string `yaml:"name,omitempty" json:"name,omitempty"`
	// IDs of the artifacts to publish. Empty means all.
	IDs []string `yaml:"ids,omitempty" json:"ids,omitempty"`
	// Whether to also publish the checksums.
	Checksum bool `yaml:"checksum,omitempty" json:"checksum,omitempty"`
	// Whether to also publish the signatures and certificates.
	Signature bool `yaml:"signature,omitempty" json:"signature,omitempty"`
	// Whether to also publish `metadata.json`.
	Meta bool `yaml:"meta,omitempty" json:"meta,omitempty"`
	// Working directory of the command. Templated.
	Dir string `yaml:"dir,omitempty" json:"dir,omitempty"`
	// Command run for each artifact. Required. Templated.
	Cmd string `yaml:"cmd,omitempty" json:"cmd,omitempty"`
	// Environment variables of the command, as `KEY=value`. Templated.
	Env []string `yaml:"env,omitempty" json:"env,omitempty"`
	// Existing files published besides the artifacts.
	ExtraFiles []ExtraFile `yaml:"extra_files,omitempty" json:"extra_files,omitempty"`
	// Whether to skip this publisher. Templated.
	Disable string `yaml:"disable,omitempty" json:"disable,omitempty" jsonschema:"oneof_type=string;boolean"`
}

// Source is the archive of the project's source code.
type Source struct {
	// Name of the archive, without its extension. Default:
	// `{{ .ProjectName }}-{{ .Version }}`. Templated.
	NameTemplate string `yaml:"name_template,omitempty" json:"name_template,omitempty"`
	// Format of the archive. Default: `tar.gz`.
	Format string `yaml:"format,omitempty" json:"format,omitempty" jsonschema:"enum=tar,enum=tgz,enum=tar.gz,enum=zip,default=tar.gz"`
	// Whether to create the archive.
	Enabled bool `yaml:"enabled,omitempty" json:"enabled,omitempty"`
	// Prefix of each file's path in the archive. Templated.
	PrefixTemplate string `yaml:"prefix_template,omitempty" json:"prefix_template,omitempty"`
	// Extra files added to the archive.
	Files []File `yaml:"files,omitempty" json:"files,omitempty"`
}

// Project is the whole configuration file.
type Project struct {
	// Version of the configuration file's format.
	Version int `yaml:"version,omitempty" json:"version,omitempty" jsonschema:"enum=2,default=2"`
	// Whether the file is written for the commercial edition; keys this tool
	// does not know then warn instead of failing the load.
	Pro bool `yaml:"pro,omitempty" json:"pro,omitempty"`
	// Name of the project. Default: guessed from the project's files and its
	// Git remote.
	ProjectName string `yaml:"project_name,omitempty" json:"project_name,omitempty"`
	// Environment variables set for the whole run, as `KEY=value`.
	// Templated.
	Env []string `yaml:"env,omitempty" json:"env,omitempty"`
	// Release on the repository's SCM.
	Release Release `yaml:"release,omitempty" json:"release,omitempty"`
	// Milestones closed on release.
	Milestones []Milestone `yaml:"milestones,omitempty" json:"milestones,omitempty"`
	// Homebrew casks.
	Casks []HomebrewCask `yaml:"homebrew_casks,omitempty" json:"homebrew_casks,omitempty"`
	// Nix packages.
	Nix []Nix `yaml:"nix,omitempty" json:"nix,omitempty"`
	// WinGet manifests.
	Winget []Winget `yaml:"winget,omitempty" json:"winget,omitempty"`
	// Arch User Repository packages of prebuilt binaries.
	AURs []AUR `yaml:"aurs,omitempty" json:"aurs,omitempty"`
	// Arch User Repository packages built from source.
	AURSources []AURSource `yaml:"aur_sources,omitempty" json:"aur_sources,omitempty"`
	// Krew plugin manifests.
	Krews []Krew `yaml:"krews,omitempty" json:"krews,omitempty"`
	// Container images built with ko.
	Kos []Ko `yaml:"kos,omitempty" json:"kos,omitempty"`
	// Scoop manifests.
	Scoops []Scoop `yaml:"scoops,omitempty" json:"scoops,omitempty"`
	// Builds of the project's binaries.
	Builds []Build `yaml:"builds,omitempty" json:"builds,omitempty"`
	// Archives of the built binaries.
	Archives []Archive `yaml:"archives,omitempty" json:"archives,omitempty"`
	// System packages of the built binaries.
	NFPMs []NFPM `yaml:"nfpms,omitempty" json:"nfpms,omitempty"`
	// Source RPM package.
	SRPM SRPM `yaml:"srpm,omitempty" json:"srpm,omitempty"`
	// Snaps of the Linux binaries.
	Snapcrafts []Snapcraft `yaml:"snapcrafts,omitempty" json:"snapcrafts,omitempty"`
	// Flatpak bundles of the Linux binaries.
	Flatpaks []Flatpak `yaml:"flatpak,omitempty" json:"flatpak,omitempty"`
	// Settings of a snapshot release.
	Snapshot Snapshot `yaml:"snapshot,omitempty" json:"snapshot,omitempty"`
	// Checksums of the release's artifacts.
	Checksum Checksum `yaml:"checksum,omitempty" json:"checksum,omitempty"`
	// Multi-platform Docker images.
	DockersV2 []DockerV2 `yaml:"dockers_v2,omitempty" json:"dockers_v2,omitempty"`
	// File of the digests of the pushed Docker images.
	DockerDigest DockerDigest `yaml:"docker_digest,omitempty" json:"docker_digest,omitempty"`
	// Uploads of the release's artifacts to Artifactory instances.
	Artifactories []Upload `yaml:"artifactories,omitempty" json:"artifactories,omitempty"`
	// Uploads of the release's artifacts to HTTP servers.
	Uploads []Upload `yaml:"uploads,omitempty" json:"uploads,omitempty"`
	// Uploads of the release's artifacts to storage buckets.
	Blobs []Blob `yaml:"blobs,omitempty" json:"blobs,omitempty"`
	// Commands run on the release's artifacts to publish them.
	Publishers []Publisher `yaml:"publishers,omitempty" json:"publishers,omitempty"`
	// Changelog of the release.
	Changelog Changelog `yaml:"changelog,omitempty" json:"changelog,omitempty"`
	// Directory the artifacts are written to. Default: `dist`.
	Dist string `yaml:"dist,omitempty" json:"dist,omitempty"`
	// Signings of the release's artifacts.
	Signs []Sign `yaml:"signs,omitempty" json:"signs,omitempty"`
	// Signing and notarization of macOS binaries.
	Notarize Notarize `yaml:"notarize,omitempty" json:"notarize,omitempty"`
	// Signings of the Docker images and manifests.
	DockerSigns []Sign `yaml:"docker_signs,omitempty" json:"docker_signs,omitempty"`
	// Signings of the built binaries.
	BinarySigns []BinarySign `yaml:"binary_signs,omitempty" json:"binary_signs,omitempty"`
	// Files that hold the SCM tokens.
	EnvFiles EnvFiles `yaml:"env_files,omitempty" json:"env_files,omitempty"`
	// Commands run before the build.
	Before Before `yaml:"before,omitempty" json:"before,omitempty"`
	// Archive of the project's source code.
	Source Source `yaml:"source,omitempty" json:"source,omitempty"`
	// Settings of the project's Go module.
	GoMod GoMod `yaml:"gomod,omitempty" json:"gomod,omitempty"`
	// Announcements of the release.
	Announce Announce `yaml:"announce,omitempty" json:"announce,omitempty"`
	// Software bills of materials.
	SBOMs []SBOM `yaml:"sboms,omitempty" json:"sboms,omitempty"`
	// Chocolatey packages.
	Chocolateys []Chocolatey `yaml:"chocolateys,omitempty" json:"chocolateys,omitempty"`
	// Reading of the current and previous tags from Git.
	Git Git `yaml:"git,omitempty" json:"git,omitempty"`
	// Whether to report the size of each artifact.
	ReportSizes bool `yaml:"report_sizes,omitempty" json:"report_sizes,omitempty"`
	// Settings of the metadata files.
	Metadata ProjectMetadata `yaml:"metadata,omitempty" json:"metadata,omitempty"`
	// Makeself self-extracting packages.
	Makeselfs []Makeself `yaml:"makeselfs,omitempty" json:"makeselfs,omitempty"`
	// macOS universal binaries.
	UniversalBinaries []UniversalBinary `yaml:"universal_binaries,omitempty" json:"universal_binaries,omitempty"`
	// Compressions of the built binaries with `upx`.
	UPXs []UPX `yaml:"upx,omitempty" json:"upx,omitempty"`
	// Publication of an MCP server to the MCP registry.
	MCP MCP `yaml:"mcp,omitempty" json:"mcp,omitempty"`
	// Publication of the release's artifacts as Iru Custom Apps.
	Iru Iru `yaml:"iru,omitempty" json:"iru,omitempty"`
	// Retrying of the operations that fail.
	Retry Retry `yaml:"retry,omitempty" json:"retry,omitempty"`
	// SCM whose token is used when several are set. Default: the
	// `GO_PUBLISHER_FORCE_TOKEN` environment variable.
	ForceToken string `yaml:"force_token,omitempty" json:"force_token,omitempty" jsonschema:"enum=github,enum=gitlab,enum=gitea,enum=,default="`
	// URLs of a GitHub Enterprise instance.
	GitHubURLs GitHubURLs `yaml:"github_urls,omitempty" json:"github_urls,omitempty"`
	// URLs of a self-hosted GitLab instance.
	GitLabURLs GitLabURLs `yaml:"gitlab_urls,omitempty" json:"gitlab_urls,omitempty"`
	// URLs of a Gitea instance.
	GiteaURLs GiteaURLs `yaml:"gitea_urls,omitempty" json:"gitea_urls,omitempty"`

	// Homebrew formulas.
	//
	// Deprecated: use `homebrew_casks`.
	Brews []Homebrew `yaml:"brews,omitempty" json:"brews,omitempty" jsonschema:"deprecated=true"`

	// Docker images of one target each.
	//
	// Deprecated: use `dockers_v2`.
	Dockers []Docker `yaml:"dockers,omitempty" json:"dockers,omitempty" jsonschema:"deprecated=true"`

	// Docker manifest lists.
	//
	// Deprecated: use `dockers_v2`.
	DockerManifests []DockerManifest `yaml:"docker_manifests,omitempty" json:"docker_manifests,omitempty" jsonschema:"deprecated=true"`
}

// ProjectMetadata is the settings of the metadata files.
type ProjectMetadata struct {
	// Modification time set on the metadata files, a Unix timestamp.
	// Templated.
	ModTimestamp string `yaml:"mod_timestamp,omitempty" json:"mod_timestamp,omitempty"`
}

// GoMod is the settings of the project's Go module.
type GoMod struct {
	// Whether to build from the Go module proxy's copy of the module at the
	// current tag. Ignored for snapshots.
	Proxy bool `yaml:"proxy,omitempty" json:"proxy,omitempty"`
	// Environment variables of the `go` commands run on the module, as
	// `KEY=value`.
	Env []string `yaml:"env,omitempty" json:"env,omitempty"`
	// Go binary run on the module. Default: `go`.
	GoBinary string `yaml:"gobinary,omitempty" json:"gobinary,omitempty"`
	// Value of the `-mod` flag of `go list -m`.
	Mod string `yaml:"mod,omitempty" json:"mod,omitempty"`
	// Directory holding the module's `go.mod`.
	Dir string `yaml:"dir,omitempty" json:"dir,omitempty"`
}

// Announce is the announcements of the release.
type Announce struct {
	// Whether to skip all announcements. Templated.
	Skip string `yaml:"skip,omitempty" json:"skip,omitempty" jsonschema:"oneof_type=string;boolean"`
	// Announcement on Twitter.
	Twitter Twitter `yaml:"twitter,omitempty" json:"twitter,omitempty"`
	// Announcement on Mastodon.
	Mastodon Mastodon `yaml:"mastodon,omitempty" json:"mastodon,omitempty"`
	// Announcement on Reddit.
	Reddit Reddit `yaml:"reddit,omitempty" json:"reddit,omitempty"`
	// Announcement on Slack.
	Slack Slack `yaml:"slack,omitempty" json:"slack,omitempty"`
	// Announcement on Discord.
	Discord Discord `yaml:"discord,omitempty" json:"discord,omitempty"`
	// Announcement on Microsoft Teams.
	Teams Teams `yaml:"teams,omitempty" json:"teams,omitempty"`
	// Announcement by email.
	SMTP SMTP `yaml:"smtp,omitempty" json:"smtp,omitempty"`
	// Announcement on Mattermost.
	Mattermost Mattermost `yaml:"mattermost,omitempty" json:"mattermost,omitempty"`
	// Announcement on LinkedIn.
	LinkedIn LinkedIn `yaml:"linkedin,omitempty" json:"linkedin,omitempty"`
	// Announcement on Telegram.
	Telegram Telegram `yaml:"telegram,omitempty" json:"telegram,omitempty"`
	// Announcement to an HTTP endpoint.
	Webhook Webhook `yaml:"webhook,omitempty" json:"webhook,omitempty"`
	// Announcement on Open Collective.
	OpenCollective OpenCollective `yaml:"opencollective,omitempty" json:"opencollective,omitempty"`
	// Announcement on Bluesky.
	Bluesky Bluesky `yaml:"bluesky,omitempty" json:"bluesky,omitempty"`
	// Announcement on a Discourse forum.
	Discourse Discourse `yaml:"discourse,omitempty" json:"discourse,omitempty"`
}

// Webhook is an announcement of the release to an HTTP endpoint.
type Webhook struct {
	// Whether to send the announcement. Templated.
	Enabled string `yaml:"enabled,omitempty" json:"enabled,omitempty" jsonschema:"oneof_type=string;boolean"`
	// Whether to skip verifying the endpoint's TLS certificate.
	SkipTLSVerify bool `yaml:"skip_tls_verify,omitempty" json:"skip_tls_verify,omitempty"`
	// Body of the request. Default: a JSON object whose `message` names the
	// project, the tag and the release URL. Templated.
	MessageTemplate string `yaml:"message_template,omitempty" json:"message_template,omitempty"`
	// URL the request is posted to. Required. Templated.
	EndpointURL string `yaml:"endpoint_url,omitempty" json:"endpoint_url,omitempty"`
	// Headers of the request; each replaces the default header of its name.
	Headers map[string]string `yaml:"headers,omitempty" json:"headers,omitempty"`
	// Content type of the request. Default:
	// `application/json; charset=utf-8`.
	ContentType string `yaml:"content_type,omitempty" json:"content_type,omitempty"`
	// HTTP status codes of a successful response. Default: `200`, `201`,
	// `202`, `204`.
	ExpectedStatusCodes []int `yaml:"expected_status_codes,omitempty" json:"expected_status_codes,omitempty"`
}

// Twitter is an announcement of the release on Twitter.
type Twitter struct {
	// Whether to post the announcement. Templated.
	Enabled string `yaml:"enabled,omitempty" json:"enabled,omitempty" jsonschema:"oneof_type=string;boolean"`
	// Message posted. Default:
	// `{{ .ProjectName }} {{ .Tag }} is out! Check it out at {{ .ReleaseURL }}`.
	// Templated.
	MessageTemplate string `yaml:"message_template,omitempty" json:"message_template,omitempty"`
}

// Mastodon is an announcement of the release on Mastodon.
type Mastodon struct {
	// Whether to post the announcement. Templated.
	Enabled string `yaml:"enabled,omitempty" json:"enabled,omitempty" jsonschema:"oneof_type=string;boolean"`
	// Message posted. Default:
	// `{{ .ProjectName }} {{ .Tag }} is out! Check it out at {{ .ReleaseURL }}`.
	// Templated.
	MessageTemplate string `yaml:"message_template,omitempty" json:"message_template,omitempty"`
	// URL of the Mastodon server; the announcement is skipped without it.
	Server string `yaml:"server" json:"server"`
}

// Reddit is an announcement of the release on Reddit.
type Reddit struct {
	// Whether to post the announcement. Templated.
	Enabled string `yaml:"enabled,omitempty" json:"enabled,omitempty" jsonschema:"oneof_type=string;boolean"`
	// ID of the Reddit application.
	ApplicationID string `yaml:"application_id,omitempty" json:"application_id,omitempty"`
	// Username of the account that posts.
	Username string `yaml:"username,omitempty" json:"username,omitempty"`
	// Title of the post. Default: `{{ .ProjectName }} {{ .Tag }} is out!`.
	// Templated.
	TitleTemplate string `yaml:"title_template,omitempty" json:"title_template,omitempty"`
	// URL the post links to. Default: `{{ .ReleaseURL }}`. Templated.
	URLTemplate string `yaml:"url_template,omitempty" json:"url_template,omitempty"`
	// Subreddit the post goes to.
	Sub string `yaml:"sub,omitempty" json:"sub,omitempty"`
}

// Slack is an announcement of the release on Slack.
type Slack struct {
	// Whether to post the announcement. Templated.
	Enabled string `yaml:"enabled,omitempty" json:"enabled,omitempty" jsonschema:"oneof_type=string;boolean"`
	// Message posted. Default:
	// `{{ .ProjectName }} {{ .Tag }} is out! Check it out at {{ .ReleaseURL }}`.
	// Templated.
	MessageTemplate string `yaml:"message_template,omitempty" json:"message_template,omitempty"`
	// Channel the message is posted to.
	Channel string `yaml:"channel,omitempty" json:"channel,omitempty"`
	// Username the message is posted as. Default: `go-publisher`.
	Username string `yaml:"username,omitempty" json:"username,omitempty"`
	// Emoji shown as the message's icon; overrides `icon_url`.
	IconEmoji string `yaml:"icon_emoji,omitempty" json:"icon_emoji,omitempty"`
	// URL of the image shown as the message's icon.
	IconURL string `yaml:"icon_url,omitempty" json:"icon_url,omitempty"`
	// Layout blocks of the message. Templated.
	Blocks []SlackBlock `yaml:"blocks,omitempty" json:"blocks,omitempty"`
	// Attachments of the message. Templated.
	Attachments []SlackAttachment `yaml:"attachments,omitempty" json:"attachments,omitempty"`
}

// Discord is an announcement of the release on Discord.
type Discord struct {
	// Whether to post the announcement. Templated.
	Enabled string `yaml:"enabled,omitempty" json:"enabled,omitempty" jsonschema:"oneof_type=string;boolean"`
	// Message posted. Default:
	// `{{ .ProjectName }} {{ .Tag }} is out! Check it out at {{ .ReleaseURL }}`.
	// Templated.
	MessageTemplate string `yaml:"message_template,omitempty" json:"message_template,omitempty"`
	// Author of the embed. Default: `go-publisher`.
	Author string `yaml:"author,omitempty" json:"author,omitempty"`
	// Color of the embed, as a decimal number. Default: `3888754`.
	Color string `yaml:"color,omitempty" json:"color,omitempty"`
	// URL of the image shown as the embed author's icon.
	IconURL string `yaml:"icon_url,omitempty" json:"icon_url,omitempty"`
}

// Teams is an announcement of the release on Microsoft Teams.
type Teams struct {
	// Whether to post the announcement. Templated.
	Enabled string `yaml:"enabled,omitempty" json:"enabled,omitempty" jsonschema:"oneof_type=string;boolean"`
	// Title of the message. Default: `{{ .ProjectName }} {{ .Tag }} is out!`.
	// Templated.
	TitleTemplate string `yaml:"title_template,omitempty" json:"title_template,omitempty"`
	// Message posted. Default:
	// `{{ .ProjectName }} {{ .Tag }} is out! Check it out at {{ .ReleaseURL }}`.
	// Templated.
	MessageTemplate string `yaml:"message_template,omitempty" json:"message_template,omitempty"`
	// Color of the message, in hexadecimal. Default: `#2D313E`.
	Color string `yaml:"color,omitempty" json:"color,omitempty"`
	// URL of the image shown beside the message.
	IconURL string `yaml:"icon_url,omitempty" json:"icon_url,omitempty"`
}

// Mattermost is an announcement of the release on Mattermost.
type Mattermost struct {
	// Whether to post the announcement. Templated.
	Enabled string `yaml:"enabled,omitempty" json:"enabled,omitempty" jsonschema:"oneof_type=string;boolean"`
	// Message posted. Default:
	// `{{ .ProjectName }} {{ .Tag }} is out! Check it out at {{ .ReleaseURL }}`.
	// Templated.
	MessageTemplate string `yaml:"message_template,omitempty" json:"message_template,omitempty"`
	// Title of the message. Default: `{{ .ProjectName }} {{ .Tag }} is out!`.
	// Templated.
	TitleTemplate string `yaml:"title_template,omitempty" json:"title_template,omitempty"`
	// Color of the message, in hexadecimal. Default: `#2D313E`.
	Color string `yaml:"color,omitempty" json:"color,omitempty"`
	// Channel the message is posted to.
	Channel string `yaml:"channel,omitempty" json:"channel,omitempty"`
	// Username the message is posted as. Default: `go-publisher`.
	Username string `yaml:"username,omitempty" json:"username,omitempty"`
	// Emoji shown as the message's icon; overrides `icon_url`.
	IconEmoji string `yaml:"icon_emoji,omitempty" json:"icon_emoji,omitempty"`
	// URL of the image shown as the message's icon.
	IconURL string `yaml:"icon_url,omitempty" json:"icon_url,omitempty"`
}

// SMTP is an announcement of the release by email.
type SMTP struct {
	// Whether to send the announcement. Templated.
	Enabled string `yaml:"enabled,omitempty" json:"enabled,omitempty" jsonschema:"oneof_type=string;boolean"`
	// Host of the SMTP server. Default: the `SMTP_HOST` environment
	// variable.
	Host string `yaml:"host,omitempty" json:"host,omitempty"`
	// Port of the SMTP server. Default: the `SMTP_PORT` environment
	// variable.
	Port int `yaml:"port,omitempty" json:"port,omitempty"`
	// Username on the SMTP server. Default: the `SMTP_USERNAME` environment
	// variable.
	Username string `yaml:"username,omitempty" json:"username,omitempty"`
	// Address the email is sent from.
	From string `yaml:"from,omitempty" json:"from,omitempty"`
	// Addresses the email is sent to.
	To []string `yaml:"to,omitempty" json:"to,omitempty"`
	// Subject of the email. Default: `{{ .ProjectName }} {{ .Tag }} is out!`.
	// Templated.
	SubjectTemplate string `yaml:"subject_template,omitempty" json:"subject_template,omitempty"`
	// Body of the email. Default:
	// `You can view details from: {{ .ReleaseURL }}`. Templated.
	BodyTemplate string `yaml:"body_template,omitempty" json:"body_template,omitempty"`
	// Whether to skip verifying the server's certificate chain and host name.
	InsecureSkipVerify bool `yaml:"insecure_skip_verify,omitempty" json:"insecure_skip_verify,omitempty"`
}

// LinkedIn is an announcement of the release on LinkedIn.
type LinkedIn struct {
	// Whether to post the announcement. Templated.
	Enabled string `yaml:"enabled,omitempty" json:"enabled,omitempty" jsonschema:"oneof_type=string;boolean"`
	// Message posted. Default:
	// `{{ .ProjectName }} {{ .Tag }} is out! Check it out at {{ .ReleaseURL }}`.
	// Templated.
	MessageTemplate string `yaml:"message_template,omitempty" json:"message_template,omitempty"`
}

// Telegram is an announcement of the release on Telegram.
type Telegram struct {
	// Whether to post the announcement. Templated.
	Enabled string `yaml:"enabled,omitempty" json:"enabled,omitempty" jsonschema:"oneof_type=string;boolean"`
	// Message posted. Default: the project name, the tag and the release URL,
	// escaped for `parse_mode`. Templated.
	MessageTemplate string `yaml:"message_template,omitempty" json:"message_template,omitempty"`
	// ID of the chat, an integer or an `@` channel name. Templated.
	ChatID string `yaml:"chat_id,omitempty" json:"chat_id,omitempty" jsonschema:"oneof_type=string;integer"`
	// Formatting of the message. Default: `MarkdownV2`.
	ParseMode string `yaml:"parse_mode,omitempty" json:"parse_mode,omitempty" jsonschema:"enum=MarkdownV2,enum=HTML,default=MarkdownV2"`
	// ID of the thread the message is posted in. Templated.
	MessageThreadID string `yaml:"message_thread_id,omitempty" json:"message_thread_id,omitempty" jsonschema:"oneof_type=string;integer"`
}

// OpenCollective is an announcement of the release as an Open Collective
// update.
type OpenCollective struct {
	// Whether to post the announcement. Templated.
	Enabled string `yaml:"enabled,omitempty" json:"enabled,omitempty" jsonschema:"oneof_type=string;boolean"`
	// Slug of the collective; the announcement is skipped without it.
	Slug string `yaml:"slug,omitempty" json:"slug,omitempty"`
	// Title of the update. Default: `{{ .Tag }}`. Templated.
	TitleTemplate string `yaml:"title_template,omitempty" json:"title_template,omitempty"`
	// HTML body of the update. Default: the project name, the tag and a link
	// to the release. Templated.
	MessageTemplate string `yaml:"message_template,omitempty" json:"message_template,omitempty"`
}

// Bluesky is an announcement of the release on Bluesky.
type Bluesky struct {
	// Whether to post the announcement. Templated.
	Enabled string `yaml:"enabled,omitempty" json:"enabled,omitempty" jsonschema:"oneof_type=string;boolean"`
	// Handle of the account that posts.
	Username string `yaml:"username,omitempty" json:"username,omitempty"`
	// Message posted. Default:
	// `{{ .ProjectName }} {{ .Tag }} is out! Check it out at {{ .ReleaseURL }}`.
	// Templated.
	MessageTemplate string `yaml:"message_template,omitempty" json:"message_template,omitempty"`
}

// Discourse is an announcement of the release as a Discourse topic.
type Discourse struct {
	// Whether to post the announcement. Templated.
	Enabled string `yaml:"enabled,omitempty" json:"enabled,omitempty" jsonschema:"oneof_type=string;boolean"`
	// Title of the topic. Default: `{{ .ProjectName }} {{ .Tag }} is out!`.
	// Templated.
	TitleTemplate string `yaml:"title_template,omitempty" json:"title_template,omitempty"`
	// Body of the post. Default:
	// `{{ .ProjectName }} {{ .Tag }} is out! Check it out at {{ .ReleaseURL }}`.
	// Templated.
	MessageTemplate string `yaml:"message_template,omitempty" json:"message_template,omitempty"`
	// Base URL of the forum, without a trailing slash. Required.
	Server string `yaml:"server,omitempty" json:"server,omitempty"`
	// ID of the category the topic is posted in. Required.
	CategoryID int `yaml:"category_id,omitempty" json:"category_id,omitempty"`
	// Username of the topic's author, an existing user. Default: `system`.
	Username string `yaml:"username,omitempty" json:"username,omitempty"`
}

// SlackBlock is a Slack layout block, taken as written.
type SlackBlock struct {
	Internal any
}

// SlackAttachment is a Slack message attachment, taken as written.
type SlackAttachment struct {
	Internal any
}

// Chocolatey is a Chocolatey package of the release's Windows archives.
type Chocolatey struct {
	// ID of the package, without spaces. Default: the project name.
	Name string `yaml:"name,omitempty" json:"name,omitempty"`
	// IDs of the archives packaged. Empty means all.
	IDs []string `yaml:"ids,omitempty" json:"ids,omitempty"`
	// URL of the package's packaging files.
	PackageSourceURL string `yaml:"package_source_url,omitempty" json:"package_source_url,omitempty"`
	// Owners of the package.
	Owners string `yaml:"owners,omitempty" json:"owners,omitempty"`
	// Human-readable title of the package. Default: the project name.
	Title string `yaml:"title,omitempty" json:"title,omitempty"`
	// Authors of the software.
	Authors string `yaml:"authors,omitempty" json:"authors,omitempty"`
	// URL of the software's project. Required.
	ProjectURL string `yaml:"project_url,omitempty" json:"project_url,omitempty"`
	// Download URL of an archive. Default: the release's download URL.
	// Templated.
	URLTemplate string `yaml:"url_template,omitempty" json:"url_template,omitempty"`
	// URL of the package's icon.
	IconURL string `yaml:"icon_url,omitempty" json:"icon_url,omitempty"`
	// Copyright notice of the software. Templated.
	Copyright string `yaml:"copyright,omitempty" json:"copyright,omitempty"`
	// URL of the software's license.
	LicenseURL string `yaml:"license_url,omitempty" json:"license_url,omitempty"`
	// Whether the user must accept the license before installing.
	RequireLicenseAcceptance bool `yaml:"require_license_acceptance,omitempty" json:"require_license_acceptance,omitempty"`
	// URL of the software's source code.
	ProjectSourceURL string `yaml:"project_source_url,omitempty" json:"project_source_url,omitempty"`
	// URL of the software's documentation.
	DocsURL string `yaml:"docs_url,omitempty" json:"docs_url,omitempty"`
	// URL of the software's issue tracker.
	BugTrackerURL string `yaml:"bug_tracker_url,omitempty" json:"bug_tracker_url,omitempty"`
	// Tags of the package, separated by spaces.
	Tags string `yaml:"tags,omitempty" json:"tags,omitempty"`
	// Short summary of the software. Templated.
	Summary string `yaml:"summary,omitempty" json:"summary,omitempty"`
	// Description of the package, in Markdown. Templated.
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	// Release notes of the package, in Markdown. Templated.
	ReleaseNotes string `yaml:"release_notes,omitempty" json:"release_notes,omitempty"`
	// Packages the package depends on.
	Dependencies []ChocolateyDependency `yaml:"dependencies,omitempty" json:"dependencies,omitempty"`
	// Whether to skip pushing the package.
	SkipPublish bool `yaml:"skip_publish,omitempty" json:"skip_publish,omitempty"`
	// API key of the repository the package is pushed to; the push is skipped
	// without it. Templated.
	APIKey string `yaml:"api_key,omitempty" json:"api_key,omitempty"`
	// URL of the repository the package is pushed to. Default:
	// `https://push.chocolatey.org/`.
	SourceRepo string `yaml:"source_repo,omitempty" json:"source_repo,omitempty"`
	// `GOAMD64` of the binaries packaged. Default: `v1`.
	Goamd64 string `yaml:"goamd64,omitempty" json:"goamd64,omitempty"`
}

// ChocolateyDependency is a package a Chocolatey package depends on.
type ChocolateyDependency struct {
	// ID of the package depended on.
	ID string `yaml:"id,omitempty" json:"id,omitempty"`
	// Version of the package depended on.
	Version string `yaml:"version,omitempty" json:"version,omitempty"`
}

// Makeself is a self-extracting Makeself package of the binaries of builds,
// one per target.
type Makeself struct {
	// ID of the package, unique among the packages. Default: `default`.
	ID string `yaml:"id,omitempty" json:"id,omitempty"`
	// Name of the package file, with its extension. Default: the project
	// name, the version and the target, ending in `.run`. Templated.
	Filename string `yaml:"filename,omitempty" json:"filename,omitempty"`
	// Path of the setup script run when the package is executed. Required.
	// Templated.
	Script string `yaml:"script" json:"script"`
	// Compression of the package. Default: the choice of `makeself`.
	// Templated.
	Compression string `yaml:"compression,omitempty" json:"compression,omitempty" jsonschema:"enum=gzip,enum=bzip2,enum=xz,enum=lzo,enum=compress,enum=none"`
	// Extra arguments of `makeself`. Templated.
	ExtraArgs []string `yaml:"extra_args,omitempty" json:"extra_args,omitempty"`
	// Extra files added to the package.
	Files []MakeselfFile `yaml:"files,omitempty" json:"files,omitempty"`
	// Whether to skip this package. Templated.
	Disable string `yaml:"disable,omitempty" json:"disable,omitempty" jsonschema:"oneof_type=string;boolean"`
	// IDs of the builds packaged. Empty means all.
	IDs []string `yaml:"ids,omitempty" json:"ids,omitempty"`
	// Operating systems packaged. Default: `linux` and `darwin`.
	Goos []string `yaml:"goos,omitempty" json:"goos,omitempty"`
	// Architectures packaged. Empty means all.
	Goarch []string `yaml:"goarch,omitempty" json:"goarch,omitempty"`
	// Name of the package. Default: `{{ .ProjectName }}`. Templated.
	Name string `yaml:"name,omitempty" json:"name,omitempty"`
	// Description of the package. Templated.
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	// Maintainer of the package. Templated.
	Maintainer string `yaml:"maintainer,omitempty" json:"maintainer,omitempty"`
	// Keywords of the package. Templated.
	Keywords []string `yaml:"keywords,omitempty" json:"keywords,omitempty"`
	// Homepage of the software. Templated.
	Homepage string `yaml:"homepage,omitempty" json:"homepage,omitempty"`
	// License of the software. Templated.
	License string `yaml:"license,omitempty" json:"license,omitempty"`
}

// MakeselfFile is a set of files added to a Makeself package.
type MakeselfFile struct {
	// Glob of the files to add. Templated.
	Source string `yaml:"src,omitempty" json:"src,omitempty"`
	// Directory the files go in, their paths kept relative to the glob.
	Destination string `yaml:"dst,omitempty" json:"dst,omitempty"`
	// Whether to drop the files' parent directories, putting each file
	// directly in `dst`.
	StripParent bool `yaml:"strip_parent,omitempty" json:"strip_parent,omitempty"`
}

// MCP is the publication of an MCP server to the MCP registry.
type MCP struct {
	// Description of the server, nested.
	//
	// Deprecated: use the same keys directly under `mcp`.
	GitHub MCPDetails `yaml:"github,omitempty" json:"github,omitempty" jsonschema:"deprecated=Use top-level MCP fields instead"`

	MCPDetails `yaml:",inline" json:",inline"`
}

// MCPDetails is the description of an MCP server in the MCP registry.
type MCPDetails struct {
	// Name of the server in reverse-DNS form, with one `/` between the
	// namespace and the name. Required. Templated.
	Name string `yaml:"name" json:"name"`
	// Human-readable title of the server. Required. Templated.
	Title string `yaml:"title" json:"title"`
	// Description of the server. Templated.
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	// URL of the server's homepage. Templated.
	Homepage string `yaml:"homepage,omitempty" json:"homepage,omitempty"`
	// Packages the server is distributed in.
	Packages []MCPPackage `yaml:"packages,omitempty" json:"packages,omitempty"`
	// Transports of the server. Has no effect.
	Transports []MCPTransport `yaml:"transports,omitempty" json:"transports,omitempty"`
	// Whether to skip the publication, or `auto` to skip it for a
	// prerelease. Templated.
	Disable string `yaml:"disable,omitempty" json:"disable,omitempty" jsonschema:"oneof_type=string;boolean"`
	// Source repository of the server.
	Repository MCPRepository `yaml:"repository,omitempty" json:"repository,omitempty"`
	// Authentication to the registry.
	Auth MCPAuth `yaml:"auth" json:"auth"`
}

// MCPRepository is the source repository of an MCP server.
type MCPRepository struct {
	// URL of the repository; the repository is left out without it.
	// Templated.
	URL string `yaml:"url,omitempty" json:"url,omitempty" `
	// Hosting service of the repository. Templated.
	Source string `yaml:"source,omitempty" json:"source,omitempty" jsonschema:"enum=github,enum=gitlab,enum=gitea"`
	// ID of the repository at its hosting service. Templated.
	ID string `yaml:"id,omitempty" json:"id,omitempty"`
	// Path of the server within the repository. Templated.
	Subfolder string `yaml:"subfolder,omitempty" json:"subfolder,omitempty"`
}

// MCPAuth is the authentication to the MCP registry.
type MCPAuth struct {
	// Method of authentication. Default: `none`. Templated.
	Type string `json:"type" yaml:"type,omitempty" jsonschema:"enum=none,enum=github,enum=github-oidc,default=none"`
	// Token of the `github` method. Default: the `MCP_GITHUB_TOKEN`
	// environment variable. Templated.
	Token string `yaml:"token,omitempty" json:"token,omitempty"`
}

// MCPPackage is a package an MCP server is distributed in.
type MCPPackage struct {
	// Registry the package is published in.
	RegistryType string `yaml:"registry_type" json:"registry_type" jsonschema:"enum=oci,enum=npm,enum=pypi,enum=nuget"`
	// Name or URL of the package. Required. Templated.
	Identifier string `yaml:"identifier" json:"identifier"`
	// Transport of the package.
	Transport MCPTransport `yaml:"transport" json:"transport"`
}

// MCPTransport is the transport an MCP server is reached by.
type MCPTransport struct {
	// Transport protocol. Required.
	Type string `yaml:"type,omitempty" json:"type,omitempty" jsonschema:"enum=stdio,enum=streamable-http,enum=sse"`
	// URL of the server; required by the HTTP transports. Templated.
	URL string `yaml:"url,omitempty" json:"url,omitempty"`
}

// Iru is the publication of the release's macOS artifacts as Iru Custom
// Apps.
type Iru struct {
	// Base URL of the Iru API. Required. Templated.
	URL string `yaml:"url,omitempty" json:"url,omitempty"`
	// Name of the Custom App, one per artifact. Default: the project name.
	// Templated.
	Name string `yaml:"name,omitempty" json:"name,omitempty"`
	// IDs of the archives and binaries to publish. Empty means all.
	IDs []string `yaml:"ids,omitempty" json:"ids,omitempty"`
	// API token. Default: the `IRU_API_TOKEN` environment variable.
	// Templated.
	APIToken string `yaml:"api_token,omitempty" json:"api_token,omitempty"`
	// ID of an existing Custom App to update instead of creating one; exactly
	// one artifact must then match. Templated.
	LibraryItemID string `yaml:"library_item_id,omitempty" json:"library_item_id,omitempty"`
	// Kind of file installed, matching the artifacts' extension. Default:
	// `zip` when creating a Custom App.
	InstallType string `yaml:"install_type,omitempty" json:"install_type,omitempty" jsonschema:"enum=package,enum=zip,enum=image,default=zip"`
	// Enforcement of the installation. Default: `install_once` when creating
	// a Custom App.
	InstallEnforcement string `yaml:"install_enforcement,omitempty" json:"install_enforcement,omitempty" jsonschema:"enum=install_once,enum=continuously_enforce,enum=no_enforcement,default=install_once"`
	// Path the zip file is extracted to; required with `install_type` `zip`
	// and allowed only with it.
	UnzipLocation *string `yaml:"unzip_location,omitempty" json:"unzip_location,omitempty"`
	// Audit script; required with `install_enforcement`
	// `continuously_enforce` and allowed only with it.
	AuditScript *string `yaml:"audit_script,omitempty" json:"audit_script,omitempty"`
	// Script run before the installation.
	PreinstallScript *string `yaml:"preinstall_script,omitempty" json:"preinstall_script,omitempty"`
	// Script run after the installation.
	PostinstallScript *string `yaml:"postinstall_script,omitempty" json:"postinstall_script,omitempty"`
	// Whether to show the app in Self Service; required with
	// `install_enforcement` `no_enforcement`.
	ShowInSelfService *bool `yaml:"show_in_self_service,omitempty" json:"show_in_self_service,omitempty"`
	// ID of the app's Self Service category; required when
	// `show_in_self_service` is `true` and allowed only then. Templated.
	SelfServiceCategoryID *string `yaml:"self_service_category_id,omitempty" json:"self_service_category_id,omitempty"`
	// Whether to mark the app as recommended in Self Service; requires
	// `show_in_self_service`.
	SelfServiceRecommended *bool `yaml:"self_service_recommended,omitempty" json:"self_service_recommended,omitempty"`
	// Whether the Custom App is active and installable. Default: `true` on
	// creation.
	Active *bool `yaml:"active,omitempty" json:"active,omitempty"`
	// Whether to restart the device after a successful installation.
	Restart *bool `yaml:"restart,omitempty" json:"restart,omitempty"`
	// Whether to skip this publication. Templated.
	Disable string `yaml:"disable,omitempty" json:"disable,omitempty" jsonschema:"oneof_type=string;boolean"`
}
