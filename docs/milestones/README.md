# План milestone-сессий llm-guard

Эта директория — durable execution map для MVP. Каждый milestone выполняется в
отдельной оркестраторской сессии и в отдельном OpenSpec change. Детальный цикл
описан в [RUNBOOK.md](RUNBOOK.md).

Исходный продуктовый scope берётся из
[`docs/light_llm_guard_go_mvp_plan.md`](../light_llm_guard_go_mvp_plan.md), а после
архивации каждого change каноническими становятся требования в
`openspec/specs/<capability>/spec.md`.

## Статусы

| Статус | Значение |
|---|---|
| `planned` | Scope зафиксирован только в milestone-файле |
| `active` | Открыт change, текущая сессия владеет milestone |
| `blocked` | Цикл остановлен с зафиксированным blocker evidence |
| `implemented` | Primary implementation и correction cycle завершены |
| `verified` | Broad checks и OpenSpec verification прошли |
| `synced` | Delta specs слиты в main specs и провалидированы |
| `archived` | Change архивирован; создан reviewed green checkpoint |

Только `archived` удовлетворяет зависимости следующего milestone. `verified` или
зелёные тесты сами по себе не считаются завершением.

## Сводное состояние

Единственный dashboard находится в [STATUS.md](STATUS.md). Оркестратор обновляет
его в начале сессии и после каждого terminal gate. Milestone-файлы не дублируют
runtime status: их checklist описывает плановое task coverage, а фактические
checkbox живут в активном OpenSpec `tasks.md`.

`Variant C` означает Composer 2.5 через Herdr. Сменить вариант можно только до
primary job и с отражением в таблице; внутри milestone варианты не смешиваются.

## Порядок выполнения

```text
M0 ──┬──> M1 ──> M2 ──> M3 ──┬──> M4 ──> M5 ──┐
     └──> R0 ─────────────────┘                ├──> M7 ──> M8
                              └──> M6 ─────────┘
```

Рекомендуемый линейный порядок для минимизации переключений контекста:

```text
M0 → R0 → M1 → M2 → M3 → M4 → M5 → M6 → M7 → M8
```

После M3 milestone M6 технически можно выполнить раньше M4/M5, если NLP workstream
заблокирован. Одновременные writing sessions в одном worktree запрещены.

## Общие границы MVP

В план входят embedded pure-Go library, structured PII, PERSON, ADDRESS, secrets,
reversible masking, safe audit/metrics abstraction и OSS stabilization. OpenAI
adapters/proxy, persistent token storage, Redis/DB, ML/ONNX, policy DSL, prompt
injection и moderation оформляются отдельными post-MVP changes.
