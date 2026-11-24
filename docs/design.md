# go-publisher

go-publisher builds and publishes releases of Go projects. It is a fork of [GoReleaser](https://github.com/goreleaser/goreleaser) v2.20.3, the open-source edition, and keeps its configuration format, so an existing GoReleaser configuration releases as it did.

It adds what a monorepo needs: a module is released from its own prefixed tags, beside the repository's other tags. The configuration is checked against its version's JSON Schema before anything runs.

## Concepts

| Term | Meaning |
|---|---|
| tag | the git tag a release is made from |
| tag prefix | the text a module's tags start with, such as `mock-module/` |
| version | the tag without its prefix and its `v`, such as `1.2.3` |
| format version | the configuration format a configuration is written in, and its schema's version |

## Configuration

The configuration keeps GoReleaser's:

- File names: `.goreleaser.yaml`, `.goreleaser.yml`, `goreleaser.yaml`, `goreleaser.yml` and their `.config/` forms, or the file `--config` names.
- `version: 2`, the format a configuration states.
- The `GORELEASER_*` environment variables, such as `GORELEASER_CURRENT_TAG`.

A configuration is YAML 1.2. Anchors and aliases are read; the merge key `<<` belongs to YAML 1.1 and is an ordinary key, so a configuration that writes one is refused.

A GoReleaser Pro configuration, one that states `pro: true`, is read with its Pro-only fields ignored.

## Validation

A configuration is checked against the schema of the format version it states before it is decoded:

- `version: 2` is checked against the newest 2.x schema.
- Every problem is reported at once, each at its file, line and column, with the lines around it.
- A field the schema does not define is a problem, and a near name is suggested.
- A configuration that does not state `version: 2` is not checked: it is warned about, naming the version go-publisher reads, and refused when it does not decode.
- A Pro configuration is not checked: its Pro-only fields are not in the schema.

## Versions

The configuration format and go-publisher carry separate versions. The format's version is the schema's, `schemas/go-publisher-v<major>.<minor>.schema.json`, and go-publisher's version is the release.

| Change | Format version | go-publisher version |
|---|---|---|
| a breaking format change | a new major | a new major |
| an additive format change, such as a new field | a new minor | a new minor |
| go-publisher's code only | none | a new patch |

| Format version | Holds |
|---|---|
| 2.0 | GoReleaser v2.20.3's format |
| 2.1 | 2.0 and `git.tag_prefix` |

- A minor's schema is a file of its own, frozen once released.
- A configuration states a major; it is read as that major's newest minor.
- Each release attaches every minor's schema, and each major's as its newest minor, so a modeline naming `go-publisher-v2.schema.json` follows the newest 2.x.
- `schemas/CHANGELOG.md` records what each minor adds and the release that first shipped it.

## Tag prefix

`git.tag_prefix` names the prefix of the tags a module is released from:

- The current tag is a tag at HEAD that starts with the prefix, or else the nearest such tag before HEAD.
- The previous tag is the nearest tag before the current one that starts with the prefix.
- The version is the tag without the prefix and its `v`: `mock-module/v1.2.3` is `1.2.3`.
- `.Tag` and `.PreviousTag` hold the tags whole, so the release is created on `mock-module/v1.2.3` and its download addresses name that tag.
- With no tag starting with the prefix, the run is refused, naming the prefix.
- Without `tag_prefix`, every tag is the project's.

## Distribution

- Binaries for Linux and macOS, `amd64` and `arm64`, attached to each release.
- `go install github.com/litebitus/go-publisher@<version>`.

## TODO

- A `merge` macro, as the YAML DSL standard defines it: a DSL definition whose function table declares the macro, and a loader that reads the expanded document, so a configuration merges a map into another.
- Remove Windows support: the winget, Chocolatey and Scoop publishers and the `windows` build target.
- `changelog.paths`, so a module's changelog lists only the commits that touch its paths.
- A prefixed `.Tag` where a tag cannot carry a `/`: Docker's default image version `{{.Tag}}`, and the Go module proxy's `module@<tag>`.
- Write the schemas' descriptions.
- A coverage floor for `make test`.
- Run `make test` in the `Release` workflow before it tags, which needs the tools the `Test` workflow installs.
- The `&tmpl.Error{}` expected errors in the brew, cask, krew, sbom, sign, exec and gitea tests.
- The code comments that name GoReleaser.
- A cache for the Node.js archives `nodedist` downloads, in the user cache directory, served only when its SHA-256 matches the embedded index, and bounded in size.
