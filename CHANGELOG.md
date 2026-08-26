# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html)
with pre-1.0 exceptions documented in
[docs/compatibility-versioning.md](docs/compatibility-versioning.md).

## [Unreleased]

## [0.1.0] — planned

> **Note:** `v0.1.0` is the target first public release. The tag and GitHub
> release are created only after a separate maintainer action following
> [docs/release-checklist.md](docs/release-checklist.md). This changelog entry
> describes the intended MVP scope; it does not assert that the release already
> exists.

### Changed

- Built-in structured PII and secret scanners now live in `internal/detect`.
  Public constructors, `Detector`, and `Finding` contracts are unchanged.
- Minimum supported Go toolchain raised to **1.26.6** to include standard-library
  security fixes required for `v0.1.0` readiness.

### Added

- Pure Go library for local PII and secret detection with reversible masking and
  restore for LLM pipelines.
- Built-in detectors: structured PII (phone, email, IP, URL, INN, SNILS, bank card,
  passport, bank account, date of birth), Russian PERSON and ADDRESS, custom regexp
  entities, and conservative secret detectors (JWT, PEM private key, API key, DSN).
- Deterministic resolver, immutable allow/mask/block policy, caller-owned
  `TokenSet`, framework-neutral safe observability (noop by default), unsafe
  development diagnostics, evaluation CLI, and representative benchmarks.
- OSS distribution boundary: external consumer fixture, `scripts/release-check.sh`
  dry-run gates, CI matrix (Go 1.26.6 + stable), bounded fuzz smoke, license
  inventory consistency checks, pinned `govulncheck` vulnerability gate, and
  maintainer-dispatched `Release` workflow (revalidation, changelog commit, tag,
  source-only GitHub Release).
- Public policies: `SECURITY.md`, `CONTRIBUTING.md`, `THIRD_PARTY_NOTICES`, release
  checklist, compatibility policy, MVP readiness matrix, and known limitations.
- M8 quality/benchmark comparison report referencing M7 baselines.

### Documentation

- README documents install paths, release gates, and the GitHub Release workflow.
- Package GoDoc clarifies concurrency, byte spans, caller-owned `TokenSet`, and
  security boundary.

### Security

- Secrets block by default; masking requires explicit configuration.
- Safe observers and public errors avoid leaking original sensitive values.
- Release readiness requires a pinned `govulncheck` scan on Go 1.26.6 with no
  reachable standard-library or module findings.

### Known limitations

See [docs/known-limitations.md](docs/known-limitations.md).

<!-- Compare/release link definitions for v0.1.0 are added only after the tag is published. -->
