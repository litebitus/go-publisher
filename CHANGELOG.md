# Changelog

## Unreleased

- Job summary
  - The GitHub Actions job summary opens with `Published <project> <version>` under a Go Publisher Summary heading, and a release lists each of its assets.
- Node builds
  - The Node.js archives a build downloads are cached in the project's `node_modules/.cache/go-publisher/nodedist/`, served only when their SHA-256 matches the release index, and the least recently used are removed past 1 GiB.

## v0.0.2 — 2026-11-24

- Configuration
  - The configuration is `.gopublisher.yml`, or the file `--config` names; GoReleaser's file names are no longer read, and `init` writes `.gopublisher.yml`.
  - `env_files` reads tokens from `~/.config/go-publisher/` by default, and a Go build's default `ldflags` set `main.builtBy` to `go-publisher`.
  - The environment variables are prefixed `GO_PUBLISHER_`, such as `GO_PUBLISHER_CURRENT_TAG`, in place of `GORELEASER_`.
- Schemas
  - The 2.0 and 2.1 schemas write every definition in place, with no embedded `$id`, so Ajv and the editors built on it compile them.
  - Every definition and field in the schemas carries a description, generated from the configuration types' doc comments.

## v0.0.1 — 2026-11-24

- Git
  - `git.tag_prefix` releases a module from its own prefixed tags, such as `mock-module/v1.2.3`: the current and previous tags carry the prefix, and the version is the tag without it.
