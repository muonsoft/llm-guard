## Context

См. мотивацию в `proposal.md`. Сейчас `internal/evaluation` обслуживает один schema v1 JSONL: loader требует positive и negative opportunities для всех 16 MVP entities, evaluator выполняет `Detect → Resolve` и сопоставляет exact `(entity, start, end)` UTF-8 byte spans, а `HasRegression` отклоняет любой FP/FN. Этот corpus из 34 synthetic cases нужен как conformance smoke, но не предназначен для heterogeneous external annotations.

Исследование candidate RU holdout показало, почему нельзя просто заменить текущий corpus внешним. В RedMadRobot `pii_benchmark` source annotations шире product contract: текущим форматам соответствуют 193/263 INN, 1/223 SNILS, 62/203 CREDIT_CARD и 16/494 PASSPORT spans; 260 PERSON runs состоят из одного name component, а STREET+HOUSE присутствуют только в 136 rows. Эти cases полезны для exposure evidence, но единый exact F1 ошибочно назвал бы intentional conservative scope detector regression.

Evaluation tooling является development/release infrastructure. Оно не должно менять public Go API, detector runtime, production dependency graph или существующий offline release gate. Исходные external values и credentials нельзя печатать в reports/errors; source licenses и redistribution rules должны проверяться до включения.

## Goals / Non-Goals

**Goals:**

- сохранить schema v1 и текущую CLI-команду как совместимый strict conformance gate;
- добавить versioned representation для suites с неполным entity coverage и source/product label mapping;
- раздельно измерять exact contract quality, незакрытую sensitive surface и Mask/Restore lifecycle;
- сделать external runs воспроизводимыми по revision, digest, mapping, generator seed и git commit;
- получить initial audited baseline и явную threshold policy для первого prefilter release;
- документировать проверяемую границу: llm-guard снижает риск для supported forms, но не является high-recall DLP.

**Non-Goals:**

- расширять detector syntax, добавлять одиночные PERSON, неполные ADDRESS, ML/NER или новые entity types;
- обещать обнаружение любой PII либо использовать external exposure score как security SLO;
- включать dataset downloader, external corpora, Python или NLP runtime в public library dependency graph;
- хранить реальные production PII, действующие credentials или TokenSet mappings;
- строить proxy/adapters либо приватный production-domain holdout в рамках этого change;
- автоматически публиковать reports или скачивать данные при `go test ./...`.

## Decisions

### 1. Сохранить v1 и добавить suite model, а не мигрировать smoke corpus

**Выбор:** schema v1 loader/evaluator и `-corpus ... -fail-on-regression` остаются совместимыми. Новые profiles используют отдельную normalized suite schema v2 и общий внутренний evaluation core, не меняя semantics старого command path.

**Rationale:** v1 кодирует полезную строгую инварианту «все известные формы проходят без FP/FN». Ослабление all-entity coverage или exact matching ради внешних datasets ухудшило бы release gate и сделало исторический baseline несопоставимым.

**Альтернатива:** расширить v1 optional fields и modes. Отклонено: один loader получил бы противоречивые validation rules и риск неявно ослабить release check.

### 2. Представить external data как source manifest, normalized suite и mapping policy

**Выбор:** разделить три versioned artifact:

1. source manifest: ID, canonical URL, immutable revision/snapshot, SHA-256, license, attribution, redistribution mode и adapter version;
2. normalized suite: source record IDs, UTF-8 text identity, source labels/spans, mapped entity, disposition и tags;
3. mapping policy: direct mappings, composition rules и machine-readable ignore reasons.

Raw external artifacts загружаются только явной dev-командой в gitignored cache. Evaluator работает только с verified local input и сам не имеет network behavior. Небольшие manifests, mapping policies, deterministic generators и approved aggregate reports коммитятся; raw corpora по умолчанию не vendor-ятся в Go module.

**Rationale:** source revision и product interpretation меняются независимо. Такое разделение позволяет пересчитать mapping без потери source truth и не выдавать изменившийся upstream за detector regression.

**Альтернатива:** закоммитить один преобразованный JSONL. Отклонено: теряются provenance и unmapped labels, растёт module zip, а review license/adapter становится невоспроизводимым.

Предпочтительный layout:

```text
testdata/evaluation/
  cases.jsonl                         # существующий v1 smoke
  manifests/*.json                   # source pins и license metadata
  mappings/*.json                    # versioned source → product policy
  generated/*.jsonl                  # small fixed-seed offline suites
  thresholds/*.json                  # approved profile boundaries
docs/evaluation/
  external-baseline.{md,json}         # safe aggregate evidence
.cache/llm-guard/evaluation/          # raw/normalized external data, gitignored
```

### 3. Использовать четыре независимых suite families

**Выбор:**

| Family | Initial source | Назначение | Gate |
|---|---|---|---|
| Conformance | существующие MVP/person/address/secrets fixtures | supported forms, exact spans, negative boundaries | обязательный PR/release |
| External RU | RedMadRobot `pii_benchmark`, pinned revision; FactRuEval pinned revision | contract domain shift и exposure gap | explicit/nightly; release evidence |
| Generated | project-authored checksum generators; ФИАС/ГАР-derived ADDRESS generation после license review | systematic valid/invalid/Unicode/context coverage | fixed-seed subset на PR, полный explicit |
| Secrets | project-authored synthetic JWT/PEM/API-key/DSN плюс reviewed MIT Gitleaks fixtures | secret syntax, near misses и default block | обязательный synthetic subset; external diagnostic |

Для RedMadRobot initial manifest следует закрепить проверенный candidate revision `f77ea831274daf980cc45c61a93c226be9d978d6`; raw artifact digest перепроверяется downloader перед фиксацией manifest. FactRuEval и Gitleaks получают immutable commit pins во время source audit. ФИАС/ГАР не загружается в CI: dated snapshot используется только offline generator, а committed outputs проходят отдельный provenance/redistribution review. Scanpatch остаётся optional robustness source и не используется как единственный gold benchmark.

**Rationale:** ни один открытый dataset не покрывает одновременно RU PERSON/ADDRESS, checksum PII, secrets, hard negatives и lifecycle. Families имеют разные evidence boundary и не должны искусственно сворачиваться в один corpus.

**Альтернатива:** один большой synthetic corpus. Отклонено: он хорошо проверяет известные правила, но не даёт независимого domain-shift evidence.

### 4. Mapping не превращает широкий source scope в обязательство detector

**Выбор:** direct source labels EMAIL, PHONE, URL, IP_ADDRESS, INN, SNILS, CREDIT_CARD и PASSPORT сохраняются и при совместимом синтаксисе map-ятся в одноимённые product entities. PERSON создаётся из упорядоченного contiguous run FIRST_NAME/LAST_NAME/MIDDLE_NAME только при достаточной текущему contract композиции. ADDRESS создаётся из composite interval только при STREET+HOUSE; CITY/REGION и другие части могут входить в interval, но не заменяют эту минимальную композицию. OMS, DRIVER_LICENSE, MILITARY_ID, BIRTH_CERTIFICATE и другие отсутствующие entities остаются exposure-only.

Каждая annotation получает disposition:

- `supported`: участвует в contract и exposure;
- `unsupported`: не участвует в contract TP/FN, но участвует в exposure;
- `ignored`: не участвует в scoring из-за доказанной annotation/adapter проблемы и обязательно имеет reason.

`ignored` нельзя использовать как удобный способ улучшить score. Изменение mapping version инвалидирует baseline и требует review diff по counts каждой disposition/reason.

**Rationale:** source truth отвечает «что является чувствительным», product mapping — «что обещает текущий prefilter». Смешение этих вопросов либо несправедливо штрафует bounded MVP, либо скрывает реальную утечку.

**Альтернатива:** отбросить неподдерживаемые labels до evaluation. Отклонено: exposure report перестал бы показывать границу продукта.

### 5. Считать contract exact, а exposure — по объединению byte intervals

**Выбор:** contract сохраняет exact entity/span TP/FP/FN. Для exposure строятся объединения gold sensitive intervals `G` и resolved predicted intervals `P`:

```text
sensitive_bytes        = |G|
covered_sensitive      = |G ∩ P|
leaked_sensitive       = |G \ P|
overmatched_bytes      = |P \ G|
byte_coverage          = covered_sensitive / sensitive_bytes
fully_covered_span     = каждый byte исходной gold span входит в P
```

Пересекающиеся source annotations и predictions учитываются один раз на уровне byte union. Exposure дополнительно группируется по source label, mapped entity и disposition. `ignored` исключается из `G`, но публикуется отдельным count. Prediction entity не обязана совпадать с source label для byte coverage: если prefilter скрыл чувствительные bytes под более общим типом, утечки этих bytes нет, хотя contract entity match может быть ошибочным.

Block action не получает искусственное «покрытие всего input» в detection exposure: `P` по-прежнему состоит из найденных spans. Отсутствие outbound text при block оценивается отдельно lifecycle profile.

**Rationale:** exact matching нужен для детерминированного контракта, но partial masking имеет измеримый security effect, который exact F1 не показывает. Byte-union formulas устойчивы к nested/overlapping labels.

**Альтернатива:** lenient any-overlap TP как основной score. Отклонено: один найденный byte мог бы объявить целую сущность защищённой.

### 6. Lifecycle profile использует те же normalized cases, но отдельные expectations

**Выбор:** normalized case MAY содержать expected action и response mutation recipe. Lifecycle runner создаёт стандартный MVP Guard/policy, проверяет отсутствие protected spans в mask result, byte-for-byte restore для неизменённых placeholders, отсутствие outbound text при block и ожидаемые miss/mutation outcomes. Failure diagnostics используют record ID, spans, labels и hashes, но не raw substring/input.

**Rationale:** detector quality не доказывает, что PII действительно удалена перед LLM и корректно восстановлена после ответа. Отдельный profile не искажает detection metrics policy behavior.

**Альтернатива:** считать unit tests masking достаточными. Отклонено: они не связывают полный набор entity cases с реальной policy и resolver sequence.

### 7. Разделить command paths и CI lanes

**Выбор:**

```text
PR / local default
  go test ./...
  existing llmguard-eval v1 --fail-on-regression
  fixed-seed generated/lifecycle smoke

Explicit data preparation
  fetch → verify manifest/digest/license → normalize → cache

Nightly / release evidence
  external contract + exposure + full generated/lifecycle
  safe JSON report → Markdown summary
```

Existing `cmd/llmguard-eval` сохраняет legacy path `-corpus` и получает взаимно исключающий `-suite` path с обязательным `-profile contract|exposure|lifecycle` и optional `-thresholds`; `-format` сохраняет текущие значения. Evaluation command читает только local verified artifacts и никогда не выполняет network access. Отдельный dev-only `cmd/llmguard-eval-data` получает `-action fetch|normalize`, `-manifest`, `-mapping` и explicit `-cache`: только action `fetch` имеет network behavior, а `normalize` повторно проверяет digest перед adapter conversion.

Release report фиксирует git commit, dirty state, Go version, OS/architecture, suite/source revisions, mapping/generator/threshold versions и exact commands. CI не коммитит baseline автоматически.

**Rationale:** обязательный PR path остаётся быстрым и воспроизводимым, а network/licensing failures не превращаются в consumer build failures. При этом release evidence можно повторить по manifest.

**Альтернатива:** скачивать datasets в обычном CI. Отклонено из-за availability, supply-chain и latency risks.

### 8. Thresholds вводятся после initial audit, а не подгоняются заранее

**Выбор:** первая external run формирует diagnostic baseline. Maintainer review проверяет adapter alignment, disposition counts и крупнейшие FN/FP clusters, после чего отдельный versioned threshold file фиксирует per-profile/entity/source absolute floors и/или допустимые regression deltas. Conformance и lifecycle invariants с первого дня остаются zero-regression gates. Любое изменение source pin или mapping требует нового baseline; detector нельзя настраивать на release holdout без явной смены его роли и нового holdout evidence.

**Rationale:** неизвестный empirical baseline нельзя честно заменить произвольным числом. Одновременно versioned thresholds не позволяют молча нормализовать ухудшение.

**Альтернатива:** сразу установить высокий global recall. Отклонено: он противоречит prefilter positioning и смешивает supported/unsupported scope.

### 9. Evaluation code остаётся internal и pure Go

**Выбор:** normalized types, interval metrics, lifecycle evaluator и report formatters размещаются в internal tooling. Source-specific adapters и downloader также не экспортируются из root package. Standard path реализуется на Go standard library и существующем llm-guard API; optional source preparation не добавляет production dependencies.

**Rationale:** benchmark infrastructure должна быть тестируема и переносима, но не становиться обещанием public API до стабилизации формата.

**Альтернатива:** сразу экспортировать evaluation SDK. Отклонено как отдельный post-release продуктовый scope.

## Risks / Trade-offs

- **[Public dataset становится фактическим training set]** → хранить mapping отдельно, не исправлять detectors только ради leaderboard и при detector expansion заменять/добавлять независимый holdout.
- **[Ошибочная normalization создаёт ложные FP/FN]** → golden adapter tests на Unicode, punctuation и disjoint BIO runs; audit counts до threshold approval.
- **[Exposure byte coverage выглядит лучше при чрезмерном masking]** → публиковать рядом overmatched bytes и contract precision/FPR; не использовать coverage отдельно.
- **[Source исчезает или меняет schema]** → immutable pins, digest verification, cache и fail-closed preparation без изменения approved report.
- **[License не разрешает redistribution]** → manifest-only download strategy; не vendor-ить raw/derived artifact до отдельного review.
- **[External report раскрывает данные]** → safe identifiers/counts/hashes, marker tests для formatters и запрет raw input/substrings.
- **[Большой suite замедляет PR]** → fixed-seed bounded smoke на PR, полный suite только explicit/nightly.
- **[Diagnostic metrics воспринимаются как DLP guarantee]** → обязательные profile/scope/limitations в каждом human-readable report и public docs.

## Migration Plan

1. Зафиксировать существующий v1 output golden tests и не менять release command semantics.
2. Добавить normalized schema, manifests/mapping validation и unit tests без подключения external source.
3. Реализовать contract/exposure interval metrics и safe formatters на небольшом project-authored suite.
4. Добавить lifecycle profile и fixed-seed generated smoke в offline checks.
5. Провести source license/provenance audit, зафиксировать revisions/digests и реализовать adapters по одному источнику.
6. Выполнить initial external run, проверить normalization/dispositions и сохранить safe diagnostic baseline.
7. После maintainer review зафиксировать thresholds и подключить explicit/nightly/release evidence lanes.
8. Обновить public positioning, known limitations и readiness matrix.

Rollback выполняется удалением новых optional commands/workflows и возвратом к существующему v1 gate; public API и detector behavior не мигрируют, raw cache не является repository state.
