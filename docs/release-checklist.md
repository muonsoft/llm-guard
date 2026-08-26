# Release checklist (`v0.1.0`)

Use this checklist before dispatching the first public `v0.1.0` release. M8 makes
the repository **ready** for release; publication is a separate maintainer action
via the GitHub **Release** workflow.

## Preflight

- [ ] On `main` with a clean working tree
- [ ] All OpenSpec M8 artifacts reviewed and archived by the orchestrator
- [ ] [CHANGELOG.md](../CHANGELOG.md) planned `v0.1.0` section ready (or
      `[Unreleased]` populated for future releases)
- [ ] [docs/dependency-license-inventory.md](dependency-license-inventory.md) and
      [THIRD_PARTY_NOTICES](../THIRD_PARTY_NOTICES) current for `go.mod`
- [ ] No unresolved security advisories for supported scope

## Local verification (side-effect free)

Run from repository root:

```bash
sh -n scripts/release-check.sh
bash -n scripts/prepare-release.sh scripts/prepare-release-test.sh
bash scripts/prepare-release-test.sh
bash scripts/prepare-release.sh 0.1.0 --check-only
./scripts/release-check.sh
GOTOOLCHAIN=go1.26.6 ./scripts/release-check.sh vuln
go test ./...
go vet ./...
go test -race ./...
go run ./cmd/llmguard-eval -corpus ./testdata/evaluation/cases.jsonl -format json -fail-on-regression
go run ./cmd/llmguard-eval -suite ./testdata/evaluation/generated/smoke.jsonl -profile contract -format json -fail-on-regression
go run ./cmd/llmguard-eval -suite ./testdata/evaluation/generated/smoke.jsonl -profile lifecycle -format json -fail-on-regression
git diff --check
```

Focused modes (also exercised in CI):

```bash
./scripts/release-check.sh consumer
./scripts/release-check.sh license
./scripts/release-check.sh fuzz
GOTOOLCHAIN=go1.26.6 ./scripts/release-check.sh vuln
```

Confirm (required for publication):

- [ ] Full dry-run exits zero (network-free; does not require external baseline
      `measured_commit` to equal HEAD or the future tag commit; does not run `vuln`)
- [ ] Pinned vulnerability scan exits zero on exact Go 1.26.6 with no reachable
      findings (`GOTOOLCHAIN=go1.26.6 ./scripts/release-check.sh vuln`; requires
      network or warmed module/vuln cache)
- [ ] `bash scripts/prepare-release.sh 0.1.0 --check-only` succeeds on current
      CHANGELOG
- [ ] External consumer passes via local `replace` (API boundary only; does not
      prove module proxy availability pre-tag)
- [ ] License manifest matches `go list -m all`
- [ ] Bounded fuzz targets pass: `FuzzMaskRestoreRoundTrip`,
      `FuzzResolveInvariants`, `FuzzCustomRegexpDetectorInvariants`
- [ ] [mvp-readiness-matrix.md](mvp-readiness-matrix.md) evidence links are current
- [ ] [m8-quality-benchmark-comparison.md](m8-quality-benchmark-comparison.md)
      recorded for this checkout

## Optional diagnostic evidence (does not block publication)

External RedMadRobot baseline is **optional diagnostic evidence**. Stale or
missing exact-tag match blocks only external benchmark claims, not `v0.1.0`
publication. A pre-dispatch report does not measure the workflow-created
changelog-only child commit or final tag.

```bash
# Inspect committed diagnostic metadata (not a publication gate)
test -f docs/evaluation/external-baseline.json

# Refresh baseline metadata after measurement at a chosen commit
./scripts/release-check.sh evidence
```

No checklist completion is required here before dispatching **Release**.

## CI verification

- [ ] Pull-request CI green on Go `1.26.6` and `stable` (test/vet)
- [ ] Race, prepare-release, consumer, evaluation, **evaluation-smoke**, fuzz,
      license, and **vuln** jobs green
- [ ] External RedMadRobot lane (`evaluation-external.yml`) optional; scheduled
      diagnostic only — failure on Hugging Face outage is acceptable

## Manual maintainer approval

- [ ] Review diff since last release-candidate checkpoint
- [ ] Confirm no unintended API or default-behavior changes
- [ ] Confirm README and GoDoc describe the release lifecycle accurately
- [ ] Approve version `v0.1.0` and release notes text

## Dispatch Release workflow (primary path)

> **OUT OF SCOPE for local scripts and normal CI.** Only after all items above are
> complete.

1. Confirm normal CI is green on current `main`.
2. In GitHub Actions, open the **Release** workflow, select branch `main`, set
   version `v0.1.0`, and run workflow.
3. The workflow revalidates offline and vulnerability gates, verifies `main` has
   not moved, finalizes the changelog commit on `main`, creates tag `v0.1.0` at
   that release commit, and publishes a source-only GitHub Release.
4. Verify `go get github.com/muonsoft/llm-guard@v0.1.0` from a clean module.
5. Post-release: update CHANGELOG compare links and milestone dashboard.

## Prepared-tag fallback

If changelog is already finalized on `main` (dated section, not `— planned`):

1. Create and push tag `v0.1.0` on the approved commit.
2. Tag push triggers **Release** workflow validation (`--require-final`) and
   GitHub Release creation without rewriting `main`.

Do **not** run tag push or release publication from `scripts/release-check.sh` or
normal CI jobs.

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
