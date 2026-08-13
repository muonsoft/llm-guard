# Compatibility and versioning policy

`github.com/muonsoft/llm-guard` is pre-1.0. This policy applies until `v1.0.0`.

## Semantic versioning (pre-1.0)

- `v0.MINOR.PATCH` follows SemVer intent, but **minor** releases may include
  breaking API changes while the public surface stabilizes.
- Any user-visible exported API, default security behavior, or documented contract
  change MUST be recorded in [CHANGELOG.md](../CHANGELOG.md).
- After `v0.1.0`, treat exported symbols in the root `llmguard` package and
  stable detector constructors as the compatibility surface unless marked
  experimental in GoDoc.

### Allowed without major bump (still changelog-worthy)

- Bug fixes that restore documented behavior
- New optional detectors or configuration hooks that do not change defaults
- Documentation and test-only changes

### Requires explicit changelog and review

- Removing or renaming exported symbols
- Changing default secret/policy behavior
- Altering byte-span or mask/restore contracts
- Adding runtime dependencies

## Go toolchain support

| Check | Version | Purpose |
| --- | --- | --- |
| Minimum supported | **Go 1.26.2** (`go` directive in `go.mod`) | documents the lowest supported toolchain |
| CI matrix | `1.26.2` and `stable` | proves minimum boundary and forward compatibility signal |
| Release dry-run | Go 1.26.2 | canonical local/CI reproduction per [release-checklist.md](release-checklist.md) |

Consumers MUST use Go **1.26.2 or newer**. Older toolchains are unsupported.

Forward failures on `stable` are compatibility signals; fixing them may require a
policy/changelog update but does not by itself change the declared minimum until
maintainers explicitly revise this document and `go.mod`.

## Module import path

Canonical import:

```text
github.com/muonsoft/llm-guard
```

External consumers MUST depend only on the public module path. Internal packages
under `internal/` are not supported for external use.

Verification:

```bash
./scripts/release-check.sh consumer
```

## Security and observability compatibility

- Default observer remains `NoopObserver`.
- Secrets remain `ActionBlock` by default.
- `TokenSet` stays caller-owned; serialization formats are not a compatibility
  guarantee in MVP.
- `WithUnsafeDevelopmentObserver` is not a stable production contract.

## Post-MVP exclusions

The following are intentionally absent from MVP compatibility promises:

- OpenAI/provider adapters and HTTP proxy
- Persistent audit storage or exporter services
- ML/NER detectors beyond rule-based MVP scope
- Zero false-negative guarantees

See [known-limitations.md](known-limitations.md).

## Release boundary

Repository readiness for `v0.1.0` is established by green `./scripts/release-check.sh`
and the [MVP readiness matrix](mvp-readiness-matrix.md). Creating the git tag and
GitHub release remains a **separate manual step** documented in the release
checklist.
