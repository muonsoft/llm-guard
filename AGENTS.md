# Agents guide — llm-guard

Light LLM Guard for Go: локальное обнаружение PII/secrets, обратимая маскировка и
восстановление для LLM-пайплайнов. Приоритет MVP: PII-first, RU-first, pure Go,
embeddable library.

## Где искать контекст

| Путь | Назначение |
|------|------------|
| `docs/light_llm_guard_go_mvp_plan.md` | Черновик MVP-плана (исходная спецификация) |
| `openspec/specs/<slug>/spec.md` | Канонические требования после переноса в OpenSpec |
| `go.mod` | Модуль `github.com/muonsoft/llm-guard` |

## Managed skills

Скиллы из hub vendored в `.agents/skills/`; lock — `skills.lock.yaml`. Не редактировать
managed-директории для переиспользуемого поведения — через hub PR и `agentmem skills pull`.

Ключевые скиллы: `golang-*`, `backend-structure`, `api-conventions`, `task-delegation`,
`herdr`, `openspec-*`, `closeout`, `agent-memory-usage`.

`llm-guard-orchestration-experiment` — project-owned overlay в
`project-skills/llm-guard-orchestration-experiment/SKILL.md`; он не управляется hub
и не должен попадать в `skills.lock.yaml`. Overlay задаёт экспериментальные правила
крупных milestone-срезов поверх managed `task-delegation`.

## Эксперимент оркестрации

Перед нетривиальной реализацией полностью прочитать
`project-skills/llm-guard-orchestration-experiment/SKILL.md` и использовать его как
project-specific overlay:

- GPT-5.6 Sol планирует, пишет task packet, проводит полный review и принимает работу;
- Composer 2.5 — единственный агент, который пишет product code;
- по умолчанию один green semantic milestone передаётся одним primary job;
- делить milestone можно только по независимым acceptance, verification и shipping boundaries;
- один пишущий Composer на worktree; read-only проверки можно выполнять параллельно;
- все замечания полного review отправляются одним correction packet.

Варианты и журнал метрик описаны в `docs/orchestration_experiment.md`. По умолчанию
использовать вариант B (крупные срезы через `cursor-executor`). Вариант C использует
долгоживущий Composer через Herdr и не включается при неуспешном preflight.

**Preflight для A/B**

```bash
node .agents/skills/task-delegation/scripts/cursor-executor.mjs doctor
```

**Дополнительный preflight для C**

```bash
test "${HERDR_ENV:-}" = 1
herdr --version
herdr agent list
```

**Проверка milestone**

```bash
go test ./...
go vet ./...
```

## OpenSpec workflow

Изменения планируются через **OpenSpec** (CLI `openspec` 1.8+).

**Ключевые пути**

| Путь | Назначение |
|------|------------|
| `openspec/config.yaml` | Контекст проекта и правила артефактов |
| `openspec/specs/<slug>/spec.md` | Канонические спеки |
| `openspec/changes/<date>-<name>/` | Текущий change |
| `openspec/changes/archive/` | Завершённые changes |

**Последовательность артефактов**

1. `proposal.md` — зачем и что меняется
2. `specs/<capability>/spec.md` — требования (WHEN/THEN/AND)
3. `design.md` — решения и tradeoffs
4. `tasks.md` — чеклист реализации

**CLI**

```bash
openspec new change "<name>"
openspec status --change "<name>" --json
openspec instructions <artifact> --change "<name>" --json
openspec archive "<name>"
```

**Skills** — `openspec-propose`, `openspec-new-change`, `openspec-apply-change`,
`openspec-explore`, `openspec-verify-change`, `openspec-archive-change`, и др. в
`.agents/skills/openspec-*/`.

**Cursor** — slash-команды в `.cursor/commands/`: `/opsx-propose`, `/opsx-new`,
`/opsx-apply`, `/opsx-archive`, `/opsx-explore`, `/opsx-verify`, `/opsx-ff`,
`/opsx-continue`, `/opsx-sync`, `/opsx-onboard`.

Первый change: перенести MVP из `docs/light_llm_guard_go_mvp_plan.md` в канонические
спеки (`/opsx-propose` или `openspec-new-change`).

## Agent memory

<!-- agentmem:closeout:start -->
This repository is registered in agentmem as `muonsoft/llm-guard`.
Run `@closeout for muonsoft/llm-guard` after non-trivial work (skill: `.agents/skills/closeout/SKILL.md`).
Consult `.agents/skills/agent-memory-usage/SKILL.md` for MCP usage.
<!-- agentmem:closeout:end -->
