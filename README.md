# llm-guard

**Read in Russian:** [README.ru.md](README.ru.md)

Lightweight open-source **LLM Guard for Go** — a local **precision-oriented
prefilter** that detects documented PII and secret forms, masks them reversibly,
and restores originals in LLM pipelines. It reduces leakage risk for supported
scope but **does not replace** high-recall DLP, generic NER, or domain-specific
security review.

**Status:** MVP **ready for `v0.1.0`** — full built-in detector pack, immutable
allow/mask/block policy, deterministic resolver, reversible masking/restore,
framework-neutral safe observability (Noop by default), and reproducible release
gates. Publication uses the GitHub **Release** workflow after green CI; see
[docs/release-checklist.md](docs/release-checklist.md).

**Library-first:** pure Go, CPU-only, embeddable in your application. Processing
runs locally with no mandatory external API or separate gateway. Built-in
detectors are **RU-first** for PERSON and ADDRESS; other families follow
documented structural rules.

## Install

Development snapshot from `main`:

```bash
go get github.com/muonsoft/llm-guard@main
```

Released version (when a semantic tag exists):

```bash
go get github.com/muonsoft/llm-guard@v0.1.0
```

Requires **Go 1.26.6+** — see
[docs/compatibility-versioning.md](docs/compatibility-versioning.md).

## Pipeline

```text
App → Guard (mask) → LLM → Guard (restore) → App
```

Keep `TokenSet` in your process memory on your side of the LLM boundary — do not
send it to the model or untrusted parties.

## Quick start

The example below mirrors the public `Mask` / `Restore` flow in
[example_test.go](example_test.go): one guard, explicit error handling, and a
caller-provided LLM callback.

```go
package myapp

import (
	"context"

	"github.com/muonsoft/llm-guard"
)

func ProcessWithLLM(
	ctx context.Context,
	prompt string,
	callLLM func(string) (string, error),
) (string, error) {
	guard, err := llmguard.New(
		llmguard.WithDetector(llmguard.NewEmailDetector()),
		llmguard.WithDetector(llmguard.NewPersonDetector()),
	)
	if err != nil {
		return "", err
	}

	masked, err := guard.Mask(ctx, prompt)
	if err != nil {
		return "", err
	}

	llmResponse, err := callLLM(masked.Text)
	if err != nil {
		return "", err
	}

	return guard.Restore(ctx, llmResponse, masked.Tokens)
}
```

When secret detectors are registered, secrets block `Mask` by default. Opt in to
reversible secret masking with `WithSecretAction(llmguard.ActionMask)` or
per-entity overrides with `WithEntityAction`. Pattern shapes are documented in
[docs/secret-patterns.md](docs/secret-patterns.md).

## Built-in coverage and defaults

| Family | Constructor(s) | Default action |
|--------|----------------|----------------|
| EMAIL | `NewEmailDetector` | mask |
| PHONE | `NewPhoneDetector` | mask |
| URL | `NewURLDetector` | mask |
| IP address | `NewIPDetector` | mask |
| INN | `NewINNDetector` | mask |
| SNILS | `NewSNILSDetector` | mask |
| PASSPORT | `NewPassportDetector` | mask |
| BANK_CARD | `NewBankCardDetector` | mask |
| BANK_ACCOUNT | `NewBankAccountDetector` | mask |
| DATE_OF_BIRTH | `NewDateOfBirthDetector` | mask |
| PERSON (RU FIO) | `NewPersonDetector` | mask |
| ADDRESS (RU compositional) | `NewAddressDetector` | mask |
| SECRET_JWT | `NewJWTDetector` | **block** |
| SECRET_PRIVATE_KEY | `NewPEMPrivateKeyDetector` | **block** |
| SECRET_API_KEY | `NewAPIKeyDetector` | **block** |
| CONNECTION_STRING | `NewDSNDetector` | **block** |
| Custom regexp | `NewCustomRegexpDetector` | mask |

PII entities mask by default. Secret entities block `Mask` unless you explicitly
configure masking. `Guard` is immutable after construction and safe for concurrent
`Detect`, `Mask`, and `Restore` calls.

## Security and operations

- Default observer is `NoopObserver` — no callbacks or side effects. Safe
  observers expose bounded metadata only and must not log original text or
  `TokenSet` contents.
- Public errors omit original input and matched sensitive values.
- `TokenSet` lives in your process memory; it does not expose sensitive values
  through `String`, `GoString`, or JSON.
- `Restore` substitutes original substrings **byte-for-byte** and does not perform
  morphological agreement when the LLM changes surrounding grammar.
- If the model mutates placeholder tokens, restore may miss or partially fail.
- **UNSAFE FOR PRODUCTION:** `WithUnsafeDevelopmentObserver` leaks raw text and
  findings for local debugging only.

## Known limitations

Full list: [docs/known-limitations.md](docs/known-limitations.md).

- **Prefilter, not DLP** — reduces risk for documented supported forms; does not
  guarantee detection of all PII (single names, city-only addresses,
  checksum-invalid INN/SNILS, unknown secret shapes).
- Conservative **RU-first** detectors; English and edge cases may be missed or
  only partially supported.
- No zero false-negative guarantee, prompt-injection protection, content
  moderation, or persistent token storage.
- Evaluation corpora are representative smoke checks, not production SLOs;
  benchmarks vary by hardware.

## Quality and development

```bash
go test ./...
go vet ./...
./scripts/release-check.sh
GOTOOLCHAIN=go1.26.6 ./scripts/release-check.sh vuln
./scripts/release-check.sh consumer
```

`./scripts/release-check.sh` runs the offline full release dry-run (network-free).
The vulnerability scan is a separate online gate and must use the exact minimum
toolchain (`GOTOOLCHAIN=go1.26.6`); see [docs/release-checklist.md](docs/release-checklist.md).

Family quality boundaries: [docs/person-quality-report.md](docs/person-quality-report.md),
[docs/address-quality-report.md](docs/address-quality-report.md). Evaluation and
benchmark methodology: [docs/evaluation-baseline.md](docs/evaluation-baseline.md),
[docs/benchmark-baseline.md](docs/benchmark-baseline.md).

## Further reading

| Document | Purpose |
|----------|---------|
| [docs/known-limitations.md](docs/known-limitations.md) | Published MVP boundaries |
| [docs/compatibility-versioning.md](docs/compatibility-versioning.md) | Pre-1.0 SemVer and Go support |
| [docs/secret-patterns.md](docs/secret-patterns.md) | Secret pattern snapshot |
| [docs/release-checklist.md](docs/release-checklist.md) | Release checklist and GitHub Release workflow |
| [CHANGELOG.md](CHANGELOG.md) | Release notes and planned `v0.1.0` scope |

## Contributing, security, license

- [CONTRIBUTING.md](CONTRIBUTING.md) — contribution and provenance requirements
- [SECURITY.md](SECURITY.md) — vulnerability reporting (private, synthetic
  fixtures only)
- [MIT](LICENSE) — Copyright (c) 2026 MuonSoft
