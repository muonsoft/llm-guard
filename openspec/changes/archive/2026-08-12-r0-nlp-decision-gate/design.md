## Context

См. мотивацию в `proposal.md`. R0 предшествует M4 PERSON и M5 ADDRESS и не реализует product detectors.

Предварительный audit зафиксировал следующие ограничения reference implementation:

- `go-razdel` предоставляет `Tokenize`/`Sentenize` с полуинтервалами UTF-8 byte offsets и уже существует как отдельный модуль `github.com/muonsoft/go-razdel`.
- Natasha 0.8 `NamesExtractor` и `AddressExtractor` используют Yargy `Tokenizer` с Pymorphy forms, а не Razdel tokenizer; поэтому `go-razdel` является product tokenizer, но не обещает token-for-token совместимость с Yargy.
- В pinned Natasha 0.8 name grammar есть конечные sequences/alternatives, morphology predicates, capitalization, dictionaries и gender-number-case relations.
- Address grammar значительно больше, но не использует agreement relations или `.match()`; top-level 0.8 rule является конечной композицией и требует street + house, а остальные parts bounded optional.
- Более новые Natasha revisions переименовали extractor и расширили semantics; они не являются executable oracle этого R0.
- Natasha name rules допускают одиночные имена/фамилии, тогда как M4 намеренно требует консервативный multi-part PERSON profile.

R0 должен проверить эти наблюдения на точных pinned revisions и превратить их в durable audit evidence.

## Goals

- Зафиксировать repository и dependency boundary для NLP workstream.
- Выбрать минимальную parser/matcher family для M4/M5 без публичного grammar DSL.
- Определить проверяемый morphology/data selection gate.
- Создать воспроизводимый differential baseline, отделённый от production Go module.
- Оставить M4/M5 только bounded implementation work, а не повторное архитектурное исследование.

## Non-Goals

- Реализация `PersonDetector`, `AddressDetector`, resolver либо Mask/Restore integration.
- API-совместимый порт Natasha, Yargy, Pymorphy или всего Razdel.
- Полная parity с Natasha address grammar или поддержка произвольных пользовательских grammars.
- Публичные morphology, token matcher, fact interpretation либо extractor APIs.
- Выделение нового Go-модуля только на основании потенциального будущего reuse.

## Decisions

### 1. `go-razdel` остаётся отдельной зависимостью, NLP runtime — internal

**Выбор:** M4/M5 используют отдельный модуль `github.com/muonsoft/go-razdel` через тонкий внутренний token adapter. Product-specific morphology contracts, matcher и PERSON/ADDRESS rules размещаются под `internal/` в `llm-guard`. Отдельный `go-natasha` в MVP не создаётся.

**Rationale:** Razdel уже имеет самостоятельный, переиспользуемый контракт сегментации. Остальная нужная функциональность определяется PII-specific false-positive и composition policies и пока не образует честный публичный Natasha-compatible продукт. Internal boundary позволяет менять grammar representation после corpus evaluation без преждевременной совместимости API.

**Rejected alternatives:**

- Новый `go-natasha`: название обещает существенно более широкую совместимость, создаёт отдельное versioning/release burden и не имеет второго consumer.
- Встраивание копии Razdel в `llm-guard`: дублирует уже проверяемую библиотеку и размывает ownership токенизации.

Выделение generic runtime после MVP требует отдельного change, второго независимого consumer и стабильного Go-native contract.

### 2. Product token stream использует UTF-8 byte offsets `go-razdel`

**Выбор:** внутренний adapter сохраняет исходные `Text`, `Start`, `End` из `go-razdel`; normalization и morphology metadata добавляются отдельно и не используются для реконструкции span. Differential harness сравнивает matched text и выполняет явное преобразование Python codepoint offsets в Go byte offsets.

**Rationale:** это согласуется с canonical `Finding` contract и исключает ошибки на многобайтовом UTF-8. Token differences с Yargy принимаются как контролируемая несовместимость и покрываются punctuation/initial/address fixtures.

**Rejected alternative:** порт Yargy regex tokenizer ради parity — создаёт второй tokenizer и не улучшает product byte-span contract.

### 3. Прямые Go-правила компилируются в bounded token matcher

**Выбор:** runtime M4/M5 поддерживает только инвентаризированные операции: sequence, alternatives, predicate conjunction/negation, optional bounded elements, bounded repetition, literal/caseless/project-owned table lookup, token kind, capitalization, closed role/form flags и span capture. Grammar representation остаётся internal и может быть обычными Go combinators либо заранее собранными таблицами.

**Rationale:** проверенный subset не требует общего context-free recursion или Yargy interpretation tree. Direct matcher проще профилировать, fuzz-тестировать, делать immutable/concurrent-safe и ограничивать по времени/памяти.

**Rejected alternatives:**

- Полный Yargy/Earley port: переносит chart, BNF normalization, interpretation trees, relations и generic DSL, которые не нужны product scope.
- Набор несвязанных регулярных выражений: плохо выражает token boundaries, morphology variants и композицию ADDRESS/PERSON.

Если полный audit обнаружит обязательную рекурсию, R0 останавливается на decision gate: failing fixture и обновлённый design должны быть reviewed до реализации M4/M5.

### 4. Morphology — project-owned bounded token annotations

**Выбор:** production runtime не включает полный morphology analyzer или generated OpenCorpora dictionary. Минимальный immutable contract хранит исходный byte span, case-folded text, закрытые kind/shape flags, lexical role bits и только bounded compatible-form classes, доказанные M4/M5 corpus. Annotation выполняется pure-функциями над project-owned компактными role/suffix/alias tables.

PERSON использует консервативные capitalization, role и русские name/patronymic/surname suffix classes. ADDRESS использует явные alias tables и numeric predicates. Ни одному detector не требуется возвращать lemma или normalized fact: normative result — исходный UTF-8 byte span.

**Rationale:** audited pure-Go candidates либо встраивают около 8.4 MiB OpenCorpora data под CC BY-SA, либо требуют external dictionary/global initialization. Natasha name lists не содержат достаточного provenance statement для production copy. Mandatory corpus выражается конечными surface-form/suffix predicates; полный анализатор закрепил бы неправильную abstraction и distribution burden.

**Rejected alternatives:**

- Копирование Natasha dictionaries без provenance audit.
- Полный morphological analyzer только ради всех upstream grammemes.
- Runtime download data: нарушает offline/embeddable requirement.
- `jus1d/gomorphy`: concurrent-safe, но включает лишний CC BY-SA dictionary и footprint.
- `therox/gomorphy` / `go-opencorpora-tools`: требуют external OpenCorpora data/build/init.

Результат выбора зафиксирован в `docs/adr/0003-nlp-runtime-boundary.md` и `docs/natasha-license-inventory.md`. Новый imported lexicon требует отдельного provenance/license audit; неизвестный source является blocker.

### 5. Natasha является pinned development reference, не product oracle

**Выбор:** `tools/natasha-reference/` содержит изолированный Python harness и lock/pin точных upstream revisions. Versioned JSONL содержит schema version, case id, entity, input, reference matches, spans и только необходимые normalized fields. Committed fixtures не содержат реальные PII.

**Rationale:** differential comparison полезен для обнаружения пропусков, но Natasha extractors имеют более широкую semantics, чем безопасная PII policy. Reproducible fixtures дают value без Python dependency у consumers.

**Rejected alternatives:** запуск Python в обычных Go tests или требование 100% parity.

### 6. Quality gate ориентирован на precision и documented differences

**Выбор:** corpus делится на positive, negative и ambiguous cases отдельно для PERSON/ADDRESS. Отчёт считает минимум accepted matches, false positives, false negatives и span mismatches. Каждое расхождение с reference классифицируется как regression, intentional product difference либо unsupported out-of-scope case.

Для M4 одиночные имя/фамилия являются intentional negative; для M5 одиночные settlement/region/street не образуют ADDRESS, а минимальная сильная композиция начинается со street+house либо эквивалентной утверждённой комбинации.

**Rationale:** для маскировки LLM input ложное срабатывание влияет на полезность текста, а address extractor для уже известной address-like строки нельзя напрямую считать детектором свободного текста.

## Risks / Trade-offs

- **[Risk] `go-razdel` и Yargy по-разному режут отдельные punctuation/hyphen cases** → добавить token-level differential fixtures и считать `go-razdel` нормативным для product spans.
- **[Risk] bounded matcher постепенно превратится в неявный generic Yargy port** → каждый новый primitive требует corpus evidence; out-of-scope constructs не добавляются ради parity.
- **[Risk] минимальная morphology ухудшит declined-form recall** → сравнить candidates на pinned corpus до ADR и документировать unsupported forms.
- **[Risk] большие словари увеличат binary/RAM footprint и false positives** → измерять footprint отдельно, использовать composition/capitalization gates и включать только sources с подтверждённым provenance.
- **[Risk] reference environment перестанет устанавливаться на новых Python versions** → pin interpreter/dependencies и хранить versioned expected fixtures, чтобы обычная Go verification не зависела от live Python run.
- **[Trade-off] internal runtime нельзя переиспользовать напрямую** → принять до появления второго consumer; последующее выделение выполняется отдельным совместимостным design.

## Migration Plan

1. R0 добавляет только audit/docs/reference tooling и не меняет public Go API.
2. Перед добавлением `go-razdel` или morphology/data в production dependency graph проверяется license inventory и `go list -deps ./...`.
3. M4 создаёт internal runtime в пределах ADR и интегрирует PERSON через существующий `Detector` interface.
4. M5 расширяет тот же bounded runtime только доказанными ADDRESS primitives и composition policy.
5. Rollback R0 tooling сводится к удалению development-only harness/fixtures; consumer migration отсутствует.

## Closed audit results

- Production morphology/data source: project-owned bounded token annotations and compact authored tables; no third-party morphology dictionary.
- Reference pair: Natasha `0.8.0` commit `b603af32...`, Yargy `0.9.0` commit `c670415...`, Pymorphy2 `0.8` and dictionary `2.4.393442.3710985` under Python `3.6.15`.
- Exact graph, construct/grammeme inventory, tokenizer differences and candidate matrix are recorded in `docs/natasha-port-scope.md` and `docs/natasha-license-inventory.md`.
