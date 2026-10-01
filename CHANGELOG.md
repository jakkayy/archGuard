# Changelog

All notable changes to this project are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and the project uses [Semantic Versioning](https://semver.org/).

## [1.0.0] - 2026-10-01

### Breaking changes
- Go module path is now `github.com/jakkayy/archGuard` (previously `github.com/archguard/archguard`, which made `go install` fail).
- Public rule types (`Rule`, `Issue`, `Severity`, `ScanContext`, `ScanResult`) moved from `internal/core` to `pkg/policy`.
- `archguard scan` exit codes: `0` passed, `1` ERROR-level violations, `2` usage/config/runtime error (previously every failure was `1`).
- Unknown rule IDs, unknown parameters, wrong parameter types, invalid severities and unsupported config versions are now errors instead of being silently ignored.
- Default `openapi-exists` path is `docs/openapi.json` everywhere (the rule previously defaulted to `docs/openapi.yaml`).

### Added
- Reusable GitHub Action (`uses: jakkayy/archGuard@v1`) with SARIF upload and an `exit-code` output.
- `archguard uninstall-hook` command and `install-hook --force`.
- `no-secrets`: GitHub token and AWS temporary key (`ASIA…`) patterns, per-line `archguard:ignore` directive.
- SARIF: `region.startLine`, `tool.driver.rules` catalog, `ruleIndex`, tool version.
- Glob and root-anchored patterns in `ignore`.
- Rule registry with parameter specs shared by `scan`, `rules` and config validation.
- GoReleaser releases with checksums and changelog; golangci-lint, race detector and an 80% coverage gate in CI.
- End-to-end CLI tests, `LICENSE`, `CONTRIBUTING.md`.

### Fixed
- `--version` printed a hard-coded `0.4.0`; it now reports the release tag (or the module version for `go install`).
- `no-secrets` never matched real `-----BEGIN RSA/EC/OPENSSH PRIVATE KEY-----` headers.
- `no-secrets` reported only one issue per file per pattern and no line numbers.
- Lower-case severities such as `error` were treated as non-errors, so scans passed when they should fail.
- Issue order changed between runs (map iteration); output is now sorted by file, line and rule.
- Directories such as `build`, `env`, `out` or `target` were skipped at any depth, hiding real source code; they are now only skipped at the project root, and `.gitignore` is respected.
- `install-hook` overwrote existing hooks and failed from subdirectories, worktrees and with `core.hooksPath`.
- `init` produced broken YAML when a path contained quotes, and left `no-secrets`/`required-files` disabled.
- Unsupported `--format` values silently fell back to console output.
- SARIF: empty results are serialized as `[]` instead of `null`.
- CI did not fail on unformatted code and used a Go version older than `go.mod` requires.

## [0.4.1] - 2026-08-02

### Added
- Automated cross-platform binary release workflow.

## [0.4.0] - 2026-08-02

- Initial public release: `init`, `scan`, `rules`, `install-hook`; rules `file-naming`, `no-secrets`, `required-files`, `openapi-exists`; console, JSON and SARIF reporters.
