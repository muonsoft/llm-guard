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
| 2026-08-12 | M0 | C | 1 | 2 | 26 / +1922 −2 | 7m43s | ≈13m | healthy / 0 timeouts / recovery не требовался / 0 orphaned panes | `go test`, `go vet`, `go test -race`, OpenSpec strict — PASS | 1 | Первый review выявил dropped standalone context errors; verification correction выявил missing validation cancel marker. |
| 2026-08-12 | R0 | C | 1 | 2 | 22 / ≈+2031 −9 | ≈18m | ≈41m | healthy / 2 bounded wait timeouts / same session resumed after canceling an optional network build / 0 orphaned panes | reference live+offline, 27 tests, `go test`, `go vet`, `go test -race`, OpenSpec strict — PASS | 2 | Full review hardened schema/entity/order/digest contracts; verification correction closed invalid UTF-8 and unhashable-type leaks. Normal-PyPI Docker build was network-limited; digest-pinned offline live run passed. |
| 2026-08-12 | M1 | C | 1 | 2 | 26 / +2500 −42 | ≈23m | ≈34m | healthy / 0 timeouts / recovery не требовался / 0 orphaned panes | focused + 3×20s fuzz, `go test`, `go vet`, `go test -race`, OpenSpec strict — PASS | 2 | Full review removed invalid TLD-prefix guessing, fixed broad helper failure, hardened typed-nil entropy input, fuzz invariants and concurrent tests; verification correction closed two evidence gaps. |
| 2026-08-12 | M2 | C | 1 | 2 | 29 / +2651 −6 | ≈10m | ≈47m | healthy / 0 control timeouts / recovery не требовался / 0 orphaned panes | focused + 20s structured fuzz, `go test`, `go vet`, `go test -race`, OpenSpec strict 5/5 — PASS | 4 | Full review hardened PHONE grouping, maximal IP/IPv6 URL handling, numeric checksum/boundaries and the corpus evaluator; verification correction closed prefix-extension and fuzz-routing gaps. |
| 2026-08-12 | M3 | C | 1 | 2 | 25 / +2400 −11 | ≈11m | ≈28m | degraded / 1 bounded post-correction read timeout / recovered on single liveness retry / 0 orphaned panes | focused + 2×2s fuzz, `go test`, `go vet`, `go test -race`, OpenSpec strict 6/6 — PASS | 2 | Full review hardened BIK locality/grammar, DATE_OF_BIRTH token boundaries, regexp metadata/error evidence and sensitive test output; verification correction fixed exact BIK-order/evidence gaps. |
| 2026-08-12 | M4 | C | 1 | 2 | 28 / +2140 −5 | ≈13m | ≈35m active plus interrupted wait | degraded then recovered / external approval quota interrupted control access; same session resumed; one expected file-delete confirmation; final bounded get/read healthy; 0 orphaned panes | focused, corpus, offline reference, `go test`, `go vet`, `go test -race`, OpenSpec strict 7/7 — PASS | 5 | Full review fixed cancellation traversal, token contracts, surname false positives, corpus accounting and test/module hygiene; verification correction removed test-only public API and closed four-kind evidence. |
| 2026-08-12 | M5 | C | 1 | 2 | 24 / +1958 −3 | ≈19m | ≈50m включая blocker/resume | degraded then recovered / initial 3 control timeouts; same change resumed after bounded list/split recovery; one stalled Cursor UI operation recovered with `Esc`; final get/read healthy; 0 orphaned panes | focused, corpus, offline reference, dependency graph, `go test`, `go vet`, `go test -race`, OpenSpec strict 8/8 — PASS | 1 | Full review fixed compound house identifiers, suffix abbreviations, malformed separators, traversal cancellation, TokenSet evidence and cache hygiene; verification correction rejected multi-segment identifiers. Change synced and archived without switching variant. |
| 2026-08-12 | M6 | C | 1 | 2 | 31 / +2563 −15 | ≈15m | ≈40m | degraded then recovered / 4 external approval-review control timeouts; single post-review liveness retry and final get/read healthy; 0 orphaned panes | focused + 4×2s fuzz, `go test`, `go vet`, `go test -race`, OpenSpec strict 10/10 — PASS | 2 | Full review fixed duplicate policy options, JWT payload decoding, PEM line boundaries, IPv6 DSN trimming, provider shapes/order, corpus/leakage evidence and README structure; verification correction closed raw-candidate diagnostics and schema coverage guarantees. Specs synced and change archived. |
