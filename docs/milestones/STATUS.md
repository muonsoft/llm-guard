# Состояние milestone-задач

Этот файл — единственный durable dashboard реализации. Его меняет только
оркестратор. Данные Composer result file являются evidence, но не обновляют статус
автоматически.

## Правила обновления

- `Status` меняется только по определениям из [README.md](README.md).
- `Tasks` имеет вид `done/total`; до OpenSpec FF — `—`.
- `Primary` и `Corrections` содержат число завершённых Composer jobs.
- `Review`, `Verify`, `Specs`, `Archive`: `—`, `active`, `blocked` или `done`.
- В `Evidence` указываются repo-relative change/archive/result/report paths, а не
  вставляется чувствительный output.
- После blocker строка сохраняет change name и последнюю достигнутую gate.

## Dashboard

| ID | Задача / shipping boundary | Зависит от | Status | Change | Variant | Tasks | Primary | Corrections | Review | Verify | Specs | Archive | Evidence |
|---|---|---|---|---|---|---:|---:|---:|---|---|---|---|---|
| [M0](M0-core-detection-baseline.md) | Detection-only core и repository baseline | — | `archived` | `m0-core-detection-baseline` | C | 11/11 | 1 | 2 | `done` | `done` | `done` | `done` | `openspec/changes/archive/2026-08-12-m0-core-detection-baseline/`, `openspec/specs/core-detection/spec.md` |
| [R0](R0-nlp-decision-gate.md) | NLP decision gate и reference harness | M0 | `archived` | `r0-nlp-decision-gate` | C | 29/29 | 1 | 2 | `done` | `done` | `done` | `done` | `openspec/changes/archive/2026-08-12-r0-nlp-decision-gate/`, `openspec/specs/nlp-reference/spec.md`, `docs/adr/0003-nlp-runtime-boundary.md` |
| [M1](M1-reversible-email-slice.md) | EMAIL `Detect → Mask → Restore` | M0 | `archived` | `m1-reversible-email-slice` | C | 21/21 | 1 | 2 | `done` | `done` | `done` | `done` | `openspec/changes/archive/2026-08-12-m1-reversible-email-slice/`, `openspec/specs/finding-resolution/spec.md`, `openspec/specs/reversible-masking/spec.md`, `openspec/specs/structured-pii/spec.md` |
| [M2](M2-structured-pii-pack.md) | Основной structured PII pack | M1 | `archived` | `m2-structured-pii-pack` | C | 29/29 | 1 | 2 | `done` | `done` | `done` | `done` | `openspec/changes/archive/2026-08-12-m2-structured-pii-pack/`, `openspec/specs/structured-pii/spec.md`, `openspec/specs/finding-resolution/spec.md` |
| [M3](M3-structured-completeness.md) | Полный structured scope и custom regexp | M2 | `archived` | `m3-structured-completeness` | C | 30/30 | 1 | 2 | `done` | `done` | `done` | `done` | `openspec/changes/archive/2026-08-12-m3-structured-completeness/`, `openspec/specs/structured-pii/spec.md`, `openspec/specs/custom-detection/spec.md`, `docs/adr/0004-mvp-public-api-boundary.md` |
| [M4](M4-russian-person.md) | Консервативный RU PERSON | M3, R0 | `archived` | `m4-russian-person` | C | 31/31 | 1 | 2 | `done` | `done` | `done` | `done` | `openspec/changes/archive/2026-08-12-m4-russian-person/`, `openspec/specs/russian-person/spec.md`, `docs/person-quality-report.md` |
| [M5](M5-russian-address.md) | Композиционный RU ADDRESS | M4 | `archived` | `m5-russian-address` | C | 33/33 | 1 | 2 | `done` | `done` | `done` | `done` | `openspec/changes/archive/2026-08-12-m5-russian-address/`, `openspec/specs/russian-address/spec.md`, `openspec/specs/finding-resolution/spec.md`, `docs/address-quality-report.md` |
| [M6](M6-secrets-and-policy.md) | Basic secrets и minimal action policy | M3 | `archived` | `m6-secrets-and-policy` | C | 33/33 | 1 | 2 | `done` | `done` | `done` | `done` | `openspec/changes/archive/2026-08-12-m6-secrets-and-policy/`, `openspec/specs/secret-detection/spec.md`, `openspec/specs/minimal-policy/spec.md`, `docs/secret-patterns.md` |
| [M7](M7-safe-observability-rc.md) | Safe observability и MVP release candidate | M5, M6 | `archived` | `m7-safe-observability-rc` | C | 35/35 | 1 | 2 | `done` | `done` | `done` | `done` | `openspec/changes/archive/2026-08-12-m7-safe-observability-rc/`, `openspec/specs/safe-observability/spec.md`, `docs/evaluation-baseline.md`, `docs/benchmark-baseline.md`, `docs/safe-surface-audit.md` |
| [M8](M8-oss-stabilization.md) | OSS-ready `v0.1.0` | M7 | `archived` | `m8-oss-stabilization` | C | 36/36 | 1 | 2 | `done` | `done` | `done` | `done` | `openspec/changes/archive/2026-08-12-m8-oss-stabilization/`, `openspec/specs/oss-distribution/spec.md`, `docs/mvp-readiness-matrix.md`, `docs/dependency-license-inventory.md`, `docs/m8-quality-benchmark-comparison.md` |
| [PEB](PEB-prefilter-evaluation-benchmark.md) | Precision-oriented prefilter evaluation benchmark | M8 | `active` | `prefilter-evaluation-benchmark` | Cursor-native | 0/52 | 0 | 0 | `—` | `—` | `—` | `—` | `openspec/changes/prefilter-evaluation-benchmark/` |

## Текущая сессия

| Поле | Значение |
|---|---|
| Milestone | PEB |
| Orchestrator session | `/opsx-apply prefilter-evaluation-benchmark` (Grok 4.6 high) |
| Active change | `prefilter-evaluation-benchmark` |
| Baseline Git status | clean `main` `8e9ee02` |
| Herdr agent / pane / Cursor session | C preflight failed (`HERDR_ENV` unset); Cursor-native `composer-worker` + Grok 4.6 high review |
| Primary result | `.agent-orchestration/results/peb-core-evaluation.md` |
| Correction results | — |
| Current gate | primary Composer job (core evaluation) |
| Blocker / resume condition | — |

После `archived` секция очищается перед завершением сессии; evidence остаётся в
строке milestone и `docs/orchestration_experiment.md`.
