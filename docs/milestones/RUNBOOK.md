# Runbook отдельной milestone-сессии

## Контракт сессии

Одна сессия оркестратора выполняет ровно один milestone и ровно один OpenSpec
change от fast-forward до archive. Она не начинает следующий milestone и не
оставляет принятый product diff поверх неархивированного change.

Стартовый запрос для новой сессии:

```text
Выполни milestone <ID> по docs/milestones/<file>.md полностью.
Следуй docs/milestones/RUNBOOK.md и варианту оркестрации из dashboard.
Не переходи к следующему milestone. Заверши OpenSpec verify, sync specs и archive.
```

Оркестратор — GPT-5.6 Sol. Единственный writer нетривиального product code —
Composer 2.5. OpenSpec-артефакты, task checkboxes, status dashboard, review,
verification, spec sync и acceptance принадлежат оркестратору.

## Terminal gates

```text
dependency gate
  → OpenSpec FF
  → executor preflight
  → primary Composer job
  → whole-diff review
  → consolidated correction
  → broad verification
  → OpenSpec verification
  → sync delta specs
  → validate main specs
  → archive change
  → reviewed green checkpoint
```

Если gate не пройден, цикл останавливается на нём. Нельзя архивировать с
неисправленным CRITICAL/WARNING, неполными tasks или красной проверкой только
потому, что OpenSpec archive допускает подтверждение warning.

## 1. Начало и claim milestone

Оркестратор обязан:

1. Прочитать `AGENTS.md`, этот runbook, milestone-файл и применимые skills.
2. Получить agentmem context pack в режиме `implement`.
3. Проверить `git status --short`, `openspec list --json` и зависимости в dashboard.
4. Убедиться, что нет другого writer и незавершённого milestone в worktree.
5. Зафиксировать baseline Git status и выбранный вариант в dashboard.
6. Перевести milestone в `active` только после успешного dependency gate.

Неизвестные dirty changes считаются пользовательскими: их не удаляют и не
перезаписывают. Если они пересекаются со scope milestone, сессия останавливается.

## 2. OpenSpec fast-forward

`OpenSpec FF` означает action/skill `/openspec-ff-change`, а не несуществующую
команду `openspec ff`.

Для milestone создаётся kebab-case change вида `m1-reversible-email-slice`:

1. `openspec new change "<change>"`.
2. `openspec status --change "<change>" --json`.
3. Для каждого требуемого artifact в dependency order —
   `openspec instructions <artifact> --change "<change>" --json`.
4. Создать полный transitive set: proposal, delta specs, design и tasks, если они
   требуются схемой.
5. Перечитать dependency artifacts перед созданием следующего.
6. Выполнить `openspec validate "<change>" --strict --no-interactive`.

Milestone-файл — входной scope, а созданные change artifacts — контракт текущей
реализации. Расширение scope сверх milestone требует явного решения пользователя.
Каждая задача в `tasks.md` должна укладываться примерно в два часа, но task не
становится отдельным executor job автоматически.

Перед реализацией оркестратор получает:

```bash
openspec instructions apply --change "<change>" --json
```

и читает все перечисленные `contextFiles`.

## 3. Execution map и выбор транспорта

Default — вариант C: один Composer 2.5 через Herdr на primary implementation и
один consolidated correction cycle.

Preflight C:

```bash
test "${HERDR_ENV:-}" = 1
herdr --version
herdr agent list
```

Перед управлением Herdr оркестратор читает skill `herdr` и сверяет установленный
CLI через `herdr --help`, `herdr agent` и `herdr pane`. Создаётся sibling pane без
перехвата focus, с текущим repository cwd. Agent получает уникальное имя и
нативно pin-ится к `composer-2.5` согласно текущему CLI.

Если preflight C не прошёл, milestone получает `blocked`. Нельзя молча перейти на
B или писать product code оркестратором. Повтор под B разрешён только после явного
abandon C, фиксации отсутствия неоднозначного diff и объявления B до нового
primary job.

Для B выполняется:

```bash
node .agents/skills/task-delegation/scripts/cursor-executor.mjs doctor
```

Варианты нельзя смешивать внутри milestone.

## 4. Task packet и primary implementation

Оркестратор создаёт `.agent-orchestration/tasks/<milestone>-primary.md` с секциями
`Goal`, `Repo context`, `Acceptance criteria`, `Files / areas to touch`, `How to
verify`, `Guardrails`, `Return format`.

Обычно весь milestone — один primary job. Разделение допустимо только если каждая
часть имеет независимые acceptance, focused verification и shippable green
boundary. Число файлов и число OpenSpec tasks сами по себе не являются причиной
делить работу.

Для C packet обязательно называет:

```text
.agent-orchestration/results/<milestone>-primary.md
```

и требует outcome, changed paths, выполненные команды с результатами, blockers и
уникальный completion marker. Composer получает только pointer:

```text
Read .agent-orchestration/tasks/<packet>.md fully and execute it exactly
```

Composer меняет product code, tests, fixtures и механический fallout и запускает
focused checks. Он не меняет OpenSpec tasks, dashboard, verification ledger или
архитектурные решения. После evidence соответствующие checkbox обновляет Sol.

## 5. Review и correction cycle

После primary job Sol обязан:

1. Прочитать result file и весь diff от сохранённого baseline.
2. Проверить каждую acceptance criterion и каждый изменённый файл.
3. Самостоятельно повторить focused checks.
4. Проверить security defaults, sensitive-data leakage, byte spans, concurrency и
   public API fallout в объёме milestone.
5. Завершить полный review до отправки первой правки.

Все actionable findings объединяются в один
`.agent-orchestration/tasks/<milestone>-correction.md` и отправляются в ту же
Composer session. Второй correction допустим только для новых фактов, выявленных
последующей проверкой. Review и acceptance нельзя делегировать Composer.

Transport C классифицируется как `healthy`, `degraded` или `blocked` по project
overlay. Result file или Herdr state `done` не заменяют diff review и независимые
проверки. При `degraded` выполняется одна bounded liveness retry; при `blocked`
milestone останавливается без fallback внутри него.

## 6. Broad и OpenSpec verification

После исправлений Sol запускает минимум:

```bash
go test ./...
go vet ./...
go test -race ./...
```

Milestone-файл может добавлять corpus, fuzz, benchmark, example или license
checks. Нельзя ослаблять команды из milestone без фиксации blocker.

Затем вызывается `/openspec-verify-change <change>`. Оркестратор проверяет:

- completeness всех tasks и requirements;
- correctness реализации каждого requirement и scenario;
- coherence с `design.md` и проектными паттернами.

Любой CRITICAL исправляется. WARNING либо исправляется, либо превращается в явно
согласованное изменение artifacts; по умолчанию milestone с WARNING не
архивируется. После исправления повторяются relevant focused checks, broad checks
и OpenSpec verification. Только после чистого отчёта статус становится
`verified`.

## 7. Sync specs и archive

Порядок строгий: сначала sync, затем archive.

1. Вызвать `/openspec-sync-specs <change>`.
2. Брать delta paths только из
   `artifactPaths.specs.existingOutputPaths` результата `status --json`.
3. Перед первым main-spec write получить один snapshot:
   `openspec instructions specs --change "<change>" --json`.
4. Интеллектуально слить ADDED/MODIFIED/REMOVED/RENAMED requirements, не
   перезаписывая неупомянутые требования main spec.
5. Выполнить:

   ```bash
   openspec validate --specs --strict --no-interactive
   ```

6. Повторно сравнить каждую delta spec с main spec. Только отсутствие оставшихся
   изменений переводит milestone в `synced`.
7. Вызвать `/openspec-archive-change <change>` и выбрать archive уже
   синхронизированного change. Не использовать `--skip-specs`, если delta specs
   существуют.
8. Проверить, что change исчез из active list, появился в archive, а main specs
   остались валидны.

Raw CLI `openspec archive -y "<change>"` допустим только как реализация action
archive после тех же проверок; он не отменяет отдельный agent-driven sync и
post-sync comparison.

## 8. Закрытие сессии

Оркестратор:

1. Обновляет dashboard: `archived`, change name, variant, task count, review,
   verify, specs и archive evidence.
2. Заполняет строку milestone в `docs/orchestration_experiment.md`.
3. Проверяет полный final diff и повторяет `git status --short`.
4. Создаёт reviewed green checkpoint до следующей milestone-сессии. Локальный
   commit допустим в рамках явно запущенного milestone; push не выполняется без
   отдельного запроса.
5. Выполняет `@closeout for muonsoft/llm-guard`.
6. Сообщает archive path, synced capabilities, команды проверки, transport state,
   correction count и известные ограничения.

Следующий milestone нельзя начинать в этой же сессии.

## Blocked resume protocol

При blocker dashboard и итог сессии должны содержать:

- gate, на котором остановились;
- baseline и изменённые paths;
- active change и незавершённые tasks;
- Herdr agent/pane/session/result path, если применимо;
- последнюю успешную и первую неуспешную проверку;
- безопасное условие для resume.

Новая сессия сначала восстанавливает этот контекст и продолжает тот же change; она
не создаёт change-дубликат и не ставит milestone `active` заново поверх
неизвестного состояния.
