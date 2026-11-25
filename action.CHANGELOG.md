# Action changelog

Each release is tagged `action/vX.Y.Z`, and its major tag, such as `action/v1`, moves to it.

## Unreleased

## v1.0.1 — 2026-11-25

- GitHub Action
  - Installs the go-publisher release it carries, the newest when the action is released; the `version` input is gone.

## v1.0.0 — 2026-11-25

- GitHub Action
  - Installs the go-publisher release `version` names, checked against the release's checksums, and runs it with `args` in `workdir`.
