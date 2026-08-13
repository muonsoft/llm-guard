# Release checklist (`v0.1.0`)

Use this checklist before creating the first public `v0.1.0` tag. M8 makes the
repository **ready** for release; it does **not** create the tag, push, or
GitHub release.

## Preflight

- [ ] On `main` (or the intended release branch) with a clean working tree
- [ ] All OpenSpec M8 artifacts reviewed and archived by the orchestrator
- [ ] [CHANGELOG.md](../CHANGELOG.md) Unreleased section finalized for `v0.1.0`
- [ ] [docs/dependency-license-inventory.md](dependency-license-inventory.md) and
      [THIRD_PARTY_NOTICES](../THIRD_PARTY_NOTICES) current for `go.mod`
- [ ] No unresolved security advisories for supported scope

## Local verification (side-effect free)

Run from repository root:

```bash
sh -n scripts/release-check.sh
./scripts/release-check.sh
go test ./...
go vet ./...
go test -race ./...
go run ./cmd/llmguard-eval -corpus ./testdata/evaluation/cases.jsonl -format json -fail-on-regression
git diff --check
```

Focused modes (also exercised in CI):

```bash
./scripts/release-check.sh consumer
./scripts/release-check.sh license
./scripts/release-check.sh fuzz
```

Confirm:

- [ ] Full dry-run exits zero
- [ ] External consumer passes via local `replace` (API boundary only; does not
      prove module proxy availability pre-tag)
- [ ] License manifest matches `go list -m all`
- [ ] Bounded fuzz targets pass: `FuzzMaskRestoreRoundTrip`,
      `FuzzResolveInvariants`, `FuzzCustomRegexpDetectorInvariants`
- [ ] [mvp-readiness-matrix.md](mvp-readiness-matrix.md) evidence links are current
- [ ] [m8-quality-benchmark-comparison.md](m8-quality-benchmark-comparison.md)
      recorded for this checkout

## CI verification

- [ ] Pull-request CI green on Go `1.26.2` and `stable` (test/vet)
- [ ] Race, consumer, evaluation, fuzz, and license jobs green
- [ ] Manual **Release check (dry-run)** workflow green (read-only permissions)

## Manual maintainer approval

- [ ] Review diff since last release-candidate checkpoint
- [ ] Confirm no unintended API or default-behavior changes
- [ ] Confirm README and GoDoc state pre-release → release transition accurately
- [ ] Approve version `v0.1.0` and release notes text

## Tag, push, and publish (separate explicit actions)

> **These steps are intentionally manual and OUT OF SCOPE for automated dry-run.**

Only after all items above are complete:

1. Create signed/annotated tag `v0.1.0` on the approved commit.
2. `git push origin v0.1.0` (and merge commit if applicable).
3. Create GitHub release from tag with notes from CHANGELOG.
4. Verify `go get github.com/muonsoft/llm-guard@v0.1.0` from a clean module.
5. Post-release: update CHANGELOG links and milestone dashboard.

Do **not** run tag push or release publication from `scripts/release-check.sh` or
the release-check workflow.

## Rollback

If a tag was created in error:

- Remove or replace the GitHub release per org policy
- Document retraction in CHANGELOG
- Do not rewrite public history without maintainer governance approval

## References

- [compatibility-versioning.md](compatibility-versioning.md)
- [dependency-license-inventory.md](dependency-license-inventory.md)
- [known-limitations.md](known-limitations.md)
- [SECURITY.md](../SECURITY.md)
