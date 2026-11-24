# Configuration schema changelog

Each version is a schema file of that name; a configuration states its major in `version`, which reads as the newest minor. A version not yet released is headed `## <version> — unreleased`; the release workflow stamps it with the go-publisher release that ships it and the date.

## 2.1 — go-publisher v0.0.1, 2026-11-24

- `git.tag_prefix`, the prefix of the tags a module is released from.

## 2.0 — go-publisher v0.0.1, 2026-11-24

- GoReleaser v2.20.3's configuration format.
