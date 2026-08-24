## 1. Core evaluation

- [x] 1.1 Добавить golden CLI/report tests, фиксирующие текущие schema v1 semantics, exact matching и legacy flags до расширения runner.
- [x] 1.2 Определить normalized suite v2 structs для source annotations, mapped entities, dispositions, lifecycle expectations и version metadata.
- [x] 1.3 Реализовать strict suite v2 loader с UTF-8 boundary, unique ID/span, disposition/reason и declared-scope validation без all-MVP coverage rule.
- [x] 1.4 Определить strict structs и loaders для source manifests, mapping policies и threshold sets с отказом на unknown fields/version mismatch.
- [x] 1.5 Выделить общий deterministic Detect → Resolve execution path, сохранив существующий v1 `Evaluate` behavior и entity order.
- [x] 1.6 Реализовать contract profile для `supported` annotations с exact entity/span TP/FP/FN, текущими rates и partial-overlap diagnostics.
- [x] 1.7 Реализовать interval-union primitives для sensitive, covered, leaked и overmatched UTF-8 byte counts без double counting.
- [x] 1.8 Реализовать exposure aggregation по source label, mapped entity и disposition, включая fully-covered spans и ignored-reason counts.
- [x] 1.9 Добавить safe deterministic JSON/Markdown formatters с profile scope, source/mapping/threshold metadata и без raw inputs/substrings.
- [x] 1.10 Расширить `cmd/llmguard-eval` mutually exclusive `-corpus`/`-suite` paths и flags `-profile`/`-thresholds`, не меняя legacy defaults и exit behavior.

## 2. Detectors и source mapping

- [ ] 2.1 Реализовать versioned direct-label mapping для EMAIL, PHONE, URL, IP_ADDRESS, INN, SNILS, CREDIT_CARD и PASSPORT без изменения production detectors.
- [ ] 2.2 Реализовать PERSON composition mapping для RedMadRobot name-component runs с явным разделением supported и exposure-only single-component annotations.
- [ ] 2.3 Реализовать ADDRESS composition mapping для STREET+HOUSE intervals и exposure-only CITY/REGION/standalone components.
- [ ] 2.4 Реализовать RedMadRobot BIO-token alignment к original UTF-8 text и normalized byte spans с fail-closed diagnostics по record ID.
- [ ] 2.5 Реализовать FactRuEval PERSON adapter и mapping на pinned source schema с сохранением исходных labels/spans.
- [ ] 2.6 Добавить deterministic checksum-aware generators для INN, SNILS, BANK_CARD и BANK_ACCOUNT valid/invalid/near-miss cases.
- [ ] 2.7 Добавить deterministic generators для PASSPORT, DATE_OF_BIRTH, PHONE, EMAIL, URL и IP Unicode/context boundary cases.
- [ ] 2.8 Добавить ADDRESS generator interface и fixed-seed synthetic implementation; подключение ФИАС/ГАР разрешать только через verified offline snapshot manifest.
- [ ] 2.9 Добавить synthetic JWT, PEM private key, API key и connection-string generators и adapter для отобранных MIT Gitleaks fixtures.

## 3. Masking и lifecycle

- [ ] 3.1 Реализовать lifecycle profile, выполняющий configured Detect → Resolve → policy → Mask → Restore на normalized cases.
- [ ] 3.2 Добавить проверку, что mask action удаляет protected original spans, а unchanged placeholders восстанавливают input byte-for-byte.
- [ ] 3.3 Добавить проверку secret block action с отсутствующим outbound text и отдельным machine-readable outcome.
- [ ] 3.4 Добавить deterministic placeholder collision, mutation, deletion и restore-miss recipes с ожидаемыми lifecycle outcomes.
- [ ] 3.5 Ограничить lifecycle failure diagnostics безопасными IDs, labels, spans, counts и hashes; добавить marker leakage assertions.

## 4. Audit и source preparation

- [ ] 4.1 Реализовать dev-only `cmd/llmguard-eval-data` actions `fetch` и `normalize` с explicit cache path и без неявного network access evaluator.
- [ ] 4.2 Добавить digest verification до и после download, atomic cache placement и запрет normalization при manifest/schema mismatch.
- [ ] 4.3 Зафиксировать RedMadRobot manifest на revision `f77ea831274daf980cc45c61a93c226be9d978d6`, подтвердить artifact SHA-256, MIT license, attribution и manifest-only distribution.
- [ ] 4.4 Провести license/provenance review FactRuEval, закрепить immutable commit/digests и определить разрешённую cache/report boundary.
- [ ] 4.5 Провести license/provenance review выбранных Gitleaks fixtures, закрепить commit/digests и исключить действующие либо сомнительные credentials.
- [ ] 4.6 Зафиксировать правила dated ФИАС/ГАР snapshot, attribution и derived-output review либо оставить source отключённым при неподтверждённой redistribution policy.
- [ ] 4.7 Добавить `.cache/llm-guard/evaluation/` в ignore policy и проверить, что raw external corpora не попадают в module/release artifacts.
- [ ] 4.8 Выполнить initial diagnostic run, проверить adapter alignment и disposition counts и сохранить безопасный aggregate baseline без raw values.
- [ ] 4.9 После review initial baseline создать `prefilter-v1` threshold set с per-profile/entity/source boundaries и documented rationale.

## 5. Tests и CI

- [ ] 5.1 Покрыть suite/manifest/mapping/threshold loaders table-driven tests для valid cases, unknown fields, invalid UTF-8 spans, duplicates и missing provenance.
- [x] 5.2 Покрыть contract evaluator exact, overlap, subset-scope и zero-denominator scenarios без изменения v1 regression tests.
- [x] 5.3 Покрыть exposure interval math nested/overlapping/disjoint/partial spans и supported/unsupported/ignored partitions.
- [ ] 5.4 Добавить adapter golden tests для Cyrillic byte offsets, punctuation, BIO runs, disjoint passport parts и unmapped sensitive labels.
- [ ] 5.5 Добавить fixed-seed reproducibility и valid/invalid invariant tests для всех structured/address/secret generators.
- [ ] 5.6 Добавить lifecycle tests для reversible mask, block, collision, mutation, deletion, restore miss и safe diagnostics.
- [x] 5.7 Добавить CLI tests для legacy compatibility, mutually exclusive flags, diagnostic no-threshold status, threshold failures и formatter safety.
- [ ] 5.8 Добавить bounded generated/lifecycle smoke в `.github/workflows/ci.yml` без network и external cache.
- [ ] 5.9 Добавить scheduled/workflow-dispatch external evaluation workflow с cache, pinned fetch, normalization и uploaded report artifact без auto-commit.
- [ ] 5.10 Обновить `scripts/release-check.sh`, чтобы offline gates оставались network-free и readiness проверяла наличие external report metadata для exact release commit.
- [ ] 5.11 Выполнить verification: `gofmt -w` для изменённых Go files, `go test ./...`, `go test -race ./...` и `go vet ./...`.
- [ ] 5.12 Выполнить offline gates: `go run ./cmd/llmguard-eval -corpus ./testdata/evaluation/cases.jsonl -format markdown -fail-on-regression` и `./scripts/release-check.sh`.
- [ ] 5.13 Воспроизвести pinned external evidence через `go run ./cmd/llmguard-eval-data -action fetch -manifest ./testdata/evaluation/manifests/redmadrobot-pii-benchmark-v1.json -cache ./.cache/llm-guard/evaluation`, затем `normalize` и `go run ./cmd/llmguard-eval -suite ./.cache/llm-guard/evaluation/normalized/redmadrobot-pii-benchmark-v1.jsonl -profile exposure -format json`.

## 6. Documentation

- [ ] 6.1 Обновить README и package documentation: precision-oriented prefilter promise, supported/unsupported examples и явный отказ от high-recall DLP guarantee.
- [ ] 6.2 Обновить `docs/known-limitations.md` границами PERSON, ADDRESS, checksum-invalid PII, unknown secrets и рекомендациями для high-risk integrations.
- [ ] 6.3 Документировать suite v2 schema, mapping dispositions, contract/exposure formulas, zero-denominator rules и safe diagnostics policy.
- [ ] 6.4 Документировать source manifests, cache/fetch/normalize commands, licenses, attribution и процедуру добавления нового benchmark source.
- [ ] 6.5 Опубликовать initial safe baseline report с exact commands, revisions, mapping/threshold versions, environment и profile limitations.
- [ ] 6.6 Обновить release checklist и MVP readiness matrix, разделив mandatory offline gates, external evidence freshness и diagnostic exposure results.
