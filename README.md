# llm-guard

Lightweight open-source **LLM Guard for Go** — локальный **precision-oriented
prefilter**: обнаружение PII и секретов в документированных поддерживаемых
формах, обратимая маскировка и восстановление текста в LLM-пайплайнах.
Снижает риск утечки для supported scope, но **не заменяет** high-recall DLP,
generic NER или domain-specific security review.

**Статус:** MVP **готов к `v0.1.0`** — полный built-in detector pack, immutable allow/mask/block policy, deterministic resolver, reversible masking/restore, framework-neutral safe observability (Noop by default), offline evaluation runner, benchmarks, OSS policies и reproducible release dry-run. Тег `v0.1.0` и GitHub release **ещё не опубликованы**; см. [docs/release-checklist.md](docs/release-checklist.md).

## Зачем

Встроить в Go-приложение как библиотеку, чтобы:

- маскировать чувствительные данные **до** отправки в LLM;
- восстанавливать placeholders **после** ответа модели;
- не зависеть от внешних API и отдельного gateway (в MVP).

```text
App → Guard (mask) → LLM → Guard (restore) → App
```

## Принципы MVP

- **PII-first**, **RU-first** (английский — вторичный)
- **library-first**, pure Go, CPU-only
- расширяемые detectors; proxy/adapters — позже

## Документация

| Путь | Описание |
|------|----------|
| [docs/light_llm_guard_go_mvp_plan.md](docs/light_llm_guard_go_mvp_plan.md) | Черновик MVP-плана |
| [docs/milestones/](docs/milestones/) | Milestone scope, status dashboard и orchestration runbook |
| [openspec/](openspec/) | Spec-driven workflow (OpenSpec) |
| [docs/evaluation-baseline.md](docs/evaluation-baseline.md) | Schema v1 smoke corpus and reproduction command |
| [docs/evaluation/suite-v2.md](docs/evaluation/suite-v2.md) | Suite schema v2, profiles, and metric formulas |
| [docs/evaluation/sources.md](docs/evaluation/sources.md) | External source manifests, fetch/normalize, licenses |
| [docs/evaluation/external-baseline.md](docs/evaluation/external-baseline.md) | RedMadRobot diagnostic prefilter baseline (safe aggregates) |
| [docs/benchmark-baseline.md](docs/benchmark-baseline.md) | Representative Detect/Mask/Restore benchmarks (no SLO) |
| [docs/m8-quality-benchmark-comparison.md](docs/m8-quality-benchmark-comparison.md) | M8 regression comparison vs M7 baselines |
| [docs/compatibility-versioning.md](docs/compatibility-versioning.md) | Pre-1.0 SemVer, Go 1.26.2+ support |
| [docs/known-limitations.md](docs/known-limitations.md) | Published MVP limitations |
| [docs/mvp-readiness-matrix.md](docs/mvp-readiness-matrix.md) | Definition of Done evidence map |
| [docs/release-checklist.md](docs/release-checklist.md) | Manual `v0.1.0` tag/release steps |
| [docs/dependency-license-inventory.md](docs/dependency-license-inventory.md) | Full dependency/data license inventory |
| [docs/secret-patterns.md](docs/secret-patterns.md) | Versioned secret pattern snapshot and update procedure |
| [SECURITY.md](SECURITY.md) | Vulnerability reporting (private, synthetic fixtures only) |
| [CONTRIBUTING.md](CONTRIBUTING.md) | Contribution, OpenSpec, and provenance requirements |
| [CHANGELOG.md](CHANGELOG.md) | Unreleased / planned `v0.1.0` notes |
| [THIRD_PARTY_NOTICES](THIRD_PARTY_NOTICES) | Shipped dependency notices |
| [AGENTS.md](AGENTS.md) | Инструкции для coding agents |

## Модуль

```bash
go get github.com/muonsoft/llm-guard
```

```go
module github.com/muonsoft/llm-guard

go 1.26.2
```

External consumer verification (synthetic fixture, local `replace`):

```bash
./scripts/release-check.sh consumer
```

Source: [testdata/external-consumer/main.go](testdata/external-consumer/main.go)

## Использование

```go
guard, err := llmguard.New(
	llmguard.WithDetector(llmguard.NewPersonDetector()),
	llmguard.WithDetector(llmguard.NewPhoneDetector()),
    llmguard.WithDetector(llmguard.NewIPDetector()),
    llmguard.WithDetector(llmguard.NewURLDetector()),
    llmguard.WithDetector(llmguard.NewINNDetector()),
    llmguard.WithDetector(llmguard.NewSNILSDetector()),
    llmguard.WithDetector(llmguard.NewBankCardDetector()),
    llmguard.WithDetector(llmguard.NewEmailDetector()),
)
if err != nil {
    return err
}

result, err := guard.Mask(ctx, prompt)
if err != nil {
    return err
}

llmResponse := callLLM(result.Text)

restored, err := guard.Restore(ctx, llmResponse, result.Tokens)
```

### Release-candidate embedded flow

```go
guard, err := llmguard.New(
    llmguard.WithDetector(llmguard.NewPersonDetector()),
    llmguard.WithDetector(llmguard.NewEmailDetector()),
    llmguard.WithDetector(llmguard.NewAPIKeyDetector()),
    llmguard.WithObserver(llmguard.ObserverFunc(func(event llmguard.Event) {
        // Safe counters only: operation, outcome, lengths, durations, entity/action counts.
    })),
)
if err != nil {
    return err
}

masked, err := guard.Mask(ctx, userPrompt)
if err != nil {
    return err // secrets block by default (ErrBlocked)
}

llmResponse := callYourLLM(masked.Text) // caller-owned boundary; keep masked.Tokens secret

restored, err := guard.Restore(ctx, llmResponse, masked.Tokens)
```

`WithObserver` is optional; default is `NoopObserver` with no callbacks or side effects.

**UNSAFE FOR PRODUCTION:** `WithUnsafeDevelopmentObserver` exposes raw text and findings for local debugging only. It is independent from the safe observer and must never be enabled in production paths.

`TokenSet` принадлежит caller и не раскрывает чувствительные значения через `String`, `GoString` или JSON.

`Restore` подставляет исходный substring byte-for-byte и **не** согласует словоформу с грамматическим контекстом, изменённым LLM (например, после переноса PERSON token в другой падеж).

### Russian PERSON (M4)

Консервативный rule-based detector для согласованных русских ФИО:

| Supported forms | Notes |
|-----------------|-------|
| `имя фамилия`, `фамилия имя` | Требуются capitalized кириллические компоненты из project-authored role tables |
| `имя отчество фамилия`, `фамилия имя отчество` | Bounded declined forms (nominative/dative/instrumental) |
| `фамилия И. О.`, `И. О. фамилия` | Initials and dots входят в единый span |

Одиночные имена/фамилии, street-like contexts, lowercase pairs и произвольные пары capitalized слов **не** принимаются. Tokenization выполняет `github.com/muonsoft/go-razdel` v0.1.0; quality boundary — `testdata/person/cases.jsonl` и `docs/person-quality-report.md`.

### Russian ADDRESS (M5)

Консервативный compositional detector для согласованных русских адресов:

| Supported composition | Notes |
|-----------------------|-------|
| `street + house` (minimum) | Explicit street labels (`ул.`, `улица`, `проспект`, `пр-т`, `переулок`, `пер.`, `шоссе`) and bounded house identifiers |
| `settlement + street + house` | Settlement (`г.`, `город`, or comma-separated capitalized name) extends an already accepted street+house span only |
| Extended building parts | `корпус`/`корп.`, `строение`/`стр.`, `квартира`/`кв.` after house in one maximal span |

Одиночные города/регионы/улицы, `settlement + street` без дома, postal index, geocoding и normalization **не** поддерживаются. Resolver policy сохраняет полный ADDRESS над вложенным PERSON (например, `ул. Академика Сахарова, 10`). Quality boundary — `testdata/address/cases.jsonl` и `docs/address-quality-report.md`.

```go
guard, err := llmguard.New(
	llmguard.WithDetector(llmguard.NewAddressDetector()),
	llmguard.WithDetector(llmguard.NewPersonDetector()),
)
```

### Structured forms (M3)

| Entity | Supported forms | Validation limits |
|--------|-----------------|-------------------|
| PASSPORT | `NNNN NNNNNN`, `NN NN NNNNNN` after RU marker (`паспорт`, `паспорт РФ`, `серия`, `паспортные данные`) | No checksum; separated `серия …, номер …` unsupported |
| BANK_ACCOUNT | 20 ASCII digits or five `####` groups after RU marker (`р/с`, `расчётный счёт`, …) | Context-first without BIK; same-line BIK triggers checksum when present |
| DATE_OF_BIRTH | `DD.MM.YYYY`, `DD/MM/YYYY`, or `D <month> YYYY` after birth marker (`дата рождения`, `д.р.`, `родился`, `родилась`) | Calendar validity only; ordinary/contract dates ignored |

Custom regexp boundaries are caller-owned; the adapter returns full matches without implicit word boundaries.

```go
detector, err := llmguard.NewCustomRegexpDetector(llmguard.RegexDetectorConfig{
    Name:       "employee_id",
    Entity:     llmguard.EntityType("EMPLOYEE_ID"),
    Pattern:    `EMP-[0-9]{6}`,
    Confidence: 0.9,
})
```

### Secrets and policy (M6)

Conservative offline detectors for structurally valid credentials plus a minimal immutable policy layer:

| Detector | Entity | Default action |
|----------|--------|----------------|
| `NewJWTDetector` | `SECRET_JWT` | `block` |
| `NewPEMPrivateKeyDetector` | `SECRET_PRIVATE_KEY` | `block` |
| `NewAPIKeyDetector` | `SECRET_API_KEY` | `block` |
| `NewDSNDetector` | `CONNECTION_STRING` | `block` |

Secrets block `Mask` by default (zero result, checkable `ErrBlocked`). Use `WithSecretAction(ActionMask)` for reversible masking or `WithEntityAction` for per-entity overrides. Pattern shapes and snapshot date are documented in [docs/secret-patterns.md](docs/secret-patterns.md).

```go
guard, err := llmguard.New(
    llmguard.WithDetector(llmguard.NewJWTDetector()),
    llmguard.WithDetector(llmguard.NewAPIKeyDetector()),
    llmguard.WithSecretAction(llmguard.ActionMask), // explicit opt-in to mask secrets
)
```

## Observability and quality (M7)

Safe terminal events cover Detect, Mask, and Restore (`success`, `error`, `blocked`, `restore_miss`) with bounded lengths, durations, and low-cardinality entity/action counts only. No Prometheus/OpenTelemetry/logging dependency in core.

```bash
# Full-MVP evaluation (representative corpus; exits non-zero on FP/FN)
go run ./cmd/llmguard-eval -corpus ./testdata/evaluation/cases.jsonl -format markdown -fail-on-regression

# Benchmarks (development baseline, not an SLO)
go test ./... -run '^$' -bench . -benchmem -count=5
```

See [docs/evaluation-baseline.md](docs/evaluation-baseline.md) and [docs/benchmark-baseline.md](docs/benchmark-baseline.md).

### Security considerations

- Default observer and public errors never include original text, finding values, detector causes, or TokenSet mappings.
- Secrets use `ActionBlock` by default; masking requires explicit `WithSecretAction(ActionMask)`.
- `TokenSet` is caller-owned and must not be logged or sent to untrusted parties.
- Unsafe development diagnostics (`WithUnsafeDevelopmentObserver`) leak sensitive data by design — keep them out of production.

### Known limitations

Full list: [docs/known-limitations.md](docs/known-limitations.md). Highlights:

- **Prefilter, not DLP** — reduces risk for documented supported forms; does not
  guarantee detection of all PII (single names, city-only addresses,
  checksum-invalid INN/SNILS, unknown secret shapes).
- Conservative RU-first detectors; English and edge cases may be missed or partially supported.
- Restore does not perform morphological agreement; LLM may mutate placeholders.
- No zero-FN guarantee, prompt-injection protection, moderation, or persistent token storage.
- Unified evaluation corpus is representative; deep regression lives in family corpora under `testdata/`.
- Benchmark numbers vary by hardware and are not production SLOs.
- No exporter server, persistent audit store, or distributed tracing in core.

## Разработка

Планирование изменений — через [OpenSpec](https://github.com/Fission-AI/OpenSpec) (`openspec` CLI 1.8+). В Cursor: `/opsx-propose`.

```bash
go test ./...
go vet ./...
./scripts/release-check.sh          # full side-effect-free release dry-run
./scripts/release-check.sh consumer # external module compile/run only
agentmem skills verify
openspec doctor
```

Supported Go: **1.26.2** (minimum) with CI forward check on `stable` — see
[docs/compatibility-versioning.md](docs/compatibility-versioning.md).

## Лицензия

[MIT](LICENSE) — Copyright (c) 2026 MuonSoft
