# Эксперимент оркестрации Sol, Composer 2.5 и Herdr

## Гипотеза

Один end-to-end блок на зелёный семантический milestone уменьшит стоимость
постановки задач и число циклов review/fix по сравнению с мелкими срезами, не
увеличивая число ошибок финальной проверки. Herdr может дополнительно уменьшить
операционные задержки за счёт долгоживущей видимой сессии Composer.

## Варианты

| Вариант | Гранулярность | Транспорт |
|---|---|---|
| A | Текущие defaults `task-delegation` | `cursor-executor` |
| B — control | Один primary job на green milestone | `cursor-executor` |
| C — primary | Один primary job на green milestone | Composer через Herdr на время milestone |

Варианты нельзя смешивать внутри milestone или одновременно запускать в одном
worktree. Вариант C используется по умолчанию после успешного Herdr preflight;
B служит контролем с той же гранулярностью. При блокировке C нельзя переключаться
на B внутри milestone — следующий вариант объявляется только после явного
завершения или отказа от текущего milestone.

Каждый C packet пишет ephemeral handoff в
`.agent-orchestration/results/<milestone>.md`. Sol классифицирует transport как
`healthy`, `degraded` или `blocked` по правилам project overlay и принимает
результат только после полного review diff и независимой проверки.

## Что измерять

- число primary jobs и correction/resume jobs;
- изменённые файлы и строки как наблюдаемую сложность, не квоту;
- wall time Composer;
- wall time постановки, ожидания, review и независимой проверки Sol;
- результат focused и final verification;
- дефекты, найденные после первого полного review;
- transport state Herdr: `healthy`, `degraded` или `blocked`;
- control-call timeouts, время восстановления и orphaned panes;
- блокировки и ручные вмешательства.

## Критерий успеха

C сравнивается прежде всего с B: оба варианта используют одинаковые крупные
milestone-срезы, поэтому разница показывает стоимость и пользу Herdr. C успешен,
если уменьшает orchestration overhead без роста blocked milestones, failed final
verification и дефектов после первого полного review. Вывод делать не раньше чем
после трёх сопоставимых milestones на вариант.

## Журнал результатов

| Дата | Milestone | Вариант | Primary | Corrections | Файлы / строки | Composer | Orchestration + review | Herdr state / recovery | Final verify | Дефекты после review | Примечания |
|---|---|---:|---:|---:|---|---|---|---|---|---:|---|
