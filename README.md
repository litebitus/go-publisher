# go-publisher

Builds and publishes releases of Go projects, monorepo modules on prefixed tags included. A fork of [GoReleaser](https://github.com/goreleaser/goreleaser) v2.20.3 by Carlos Alexandro Becker and contributors.

## Features

- Everything GoReleaser's open-source edition does: builds, archives, checksums, releases on GitHub, GitLab and Gitea, packages and package managers, container images, signing and SBOMs.
- A GoReleaser configuration works as it is: the same file names, the same `version: 2` schema and the same `GORELEASER_*` environment variables.
- `git.tag_prefix` releases a module from its own prefixed tags, such as `mock-module/v1.2.3`, beside the repository's other tags: the current and previous tags carry the prefix, and the version is the tag without it.
- A GoReleaser Pro configuration runs with its Pro-only features ignored.

The design, the configuration and its versions are in [docs/design.md](docs/design.md).

## Usage

```sh
go install github.com/litebitus/go-publisher@<version>
```

```sh
go-publisher check
go-publisher release --snapshot --clean
go-publisher release --clean
```

A module in a monorepo, tagged `mock-module/vX.Y.Z`:

```yaml
version: 2
git:
  tag_prefix: mock-module/
builds:
  - dir: mock-module
```

## Published schemas

Every release attaches the configuration's JSON Schemas as assets, for a modeline to name:

| Schema | Asset | In the latest release |
|---|---|---|
| a minor | `go-publisher-v<major>.<minor>.schema.json` | [`go-publisher-v2.1.schema.json`](https://github.com/litebitus/go-publisher/releases/latest/download/go-publisher-v2.1.schema.json) |
| a major, as its newest minor | `go-publisher-v<major>.schema.json` | [`go-publisher-v2.schema.json`](https://github.com/litebitus/go-publisher/releases/latest/download/go-publisher-v2.schema.json) |

A configuration naming the newest 2.x never needs its modeline edited:

```yaml
# yaml-language-server: $schema=https://github.com/litebitus/go-publisher/releases/latest/download/go-publisher-v2.schema.json
version: 2
```

The schemas are in [schemas/](schemas), one file per minor, and what each minor adds is in [schemas/CHANGELOG.md](schemas/CHANGELOG.md).

## Development

Go builds and tests the repo.

| Target | Does |
|---|---|
| `make build` | runs the tests, then builds `bin/go-publisher` |
| `make test` | `go vet`, a `gofmt` check, and the Go tests with the race detector and coverage |

The packagers' tests run their tools, such as snapcraft, flatpak, Zig, cosign and syft; the `Test` workflow installs them.

### Releases

The `Release` workflow, dispatched by hand, computes the next version, stamps `CHANGELOG.md` and `schemas/CHANGELOG.md`, tags it, and publishes the archives and the schemas with go-publisher itself.
