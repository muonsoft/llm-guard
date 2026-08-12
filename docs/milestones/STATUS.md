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
| [M4](M4-russian-person.md) | Консервативный RU PERSON | M3, R0 | `planned` | — | C | — | 0 | 0 | — | — | — | — | — |
| [M5](M5-russian-address.md) | Композиционный RU ADDRESS | M4 | `planned` | — | C | — | 0 | 0 | — | — | — | — | — |
| [M6](M6-secrets-and-policy.md) | Basic secrets и minimal action policy | M3 | `planned` | — | C | — | 0 | 0 | — | — | — | — | — |
| [M7](M7-safe-observability-rc.md) | Safe observability и MVP release candidate | M5, M6 | `planned` | — | C | — | 0 | 0 | — | — | — | — | — |
| [M8](M8-oss-stabilization.md) | OSS-ready `v0.1.0` | M7 | `planned` | — | C | — | 0 | 0 | — | — | — | — | — |

## Текущая сессия

| Поле | Значение |
|---|---|
| Milestone | — |
| Orchestrator session | — |
| Active change | — |
| Baseline Git status | — |
| Herdr agent / pane / Cursor session | — |
| Primary result | — |
| Correction result | — |
| Current gate | — |
| Blocker / resume condition | — |

После `archived` секция очищается перед завершением сессии; evidence остаётся в
строке milestone и `docs/orchestration_experiment.md`.
