# Changelog

## Unreleased

- Configuration
  - The configuration is `.gopublisher.yml`, or the file `--config` names; GoReleaser's file names are no longer read, and `init` writes `.gopublisher.yml`.
- Schemas
  - The 2.0 and 2.1 schemas write every definition in place, with no embedded `$id`, so Ajv and the editors built on it compile them.

## v0.0.1 — 2026-11-24

- Git
  - `git.tag_prefix` releases a module from its own prefixed tags, such as `mock-module/v1.2.3`: the current and previous tags carry the prefix, and the version is the tag without it.
