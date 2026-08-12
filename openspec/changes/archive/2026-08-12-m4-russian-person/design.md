## Context

См. `proposal.md` и `specs/russian-person/spec.md`. R0 уже принял ADR 0003: production tokenization выполняет отдельный pure-Go `go-razdel`, а product-specific annotations и bounded rules остаются под `internal/`; Natasha 0.8.0/Yargy 0.9.0 используются только как pinned development reference. Текущий публичный паттерн built-in detectors — фабрика `New<Entity>Detector() Detector`, immutable value и регистрация через `WithDetector`.

M4 должен оставить внутреннюю основу, которую M5 сможет расширить для ADDRESS, но не должен заранее реализовывать generic grammar DSL, morphology API или address primitives. Нормативная revision tokenizer: commit `5cd53c7a1d02780285406c6c9f1635a89953c27a`, Go pseudo-version `v0.0.0-20260425122647-5cd53c7a1d02`.

## Goals

- Сохранить каждый accepted span как прямую композицию исходных `go-razdel` byte offsets.
- Выразить обязательные PERSON формы конечными правилами и маленькими auditable role/form tables.
- Сделать precision boundary воспроизводимой через versioned product corpus и pinned Natasha differential fixtures.
- Сохранить public API и concurrency model существующих detectors.
- Оставить минимальный reusable token-annotation seam для M5 без premature extraction.

## Decisions

### 1. Публичная фабрика следует существующему Detector API

Добавляется `NewPersonDetector() Detector`, возвращающий immutable value с именем `person`, entity `PERSON` и фиксированной confidence. Config/lexicon options в M4 не вводятся: изменяемые caller tables усложнили бы validation, provenance и concurrent ownership до появления такого требования.

Альтернативы: экспортировать `PersonDetector` struct либо конфиг словарей. Они отклонены, потому что текущие built-ins скрывают concrete implementation, а custom behavior уже покрывает общий `Detector` interface.

### 2. `internal/nlp` адаптирует `go-razdel`, но не копирует tokenizer

Тонкий adapter преобразует `razdel.Tokenize(text)` в internal value с исходными `Text`, `Start`, `End`, folded form, closed token kind, capitalization/initial/hyphen flags, lexical role bits и bounded form-class bits. Adapter проверяет инвариант `text[Start:End] == Text`; matching никогда не реконструирует offsets из folded строки.

Internal types остаются unexported и value-oriented. `go-razdel` добавляется ровно на audited pseudo-version. Логические subparts hyphenated tokens в M4 не нужны обязательному corpus и не вводятся без отдельного test case; M5 может добавить bounded split с сохранением source spans.

Альтернатива — regexp по исходной строке — отклонена: она дублирует tokenizer boundary policy и затрудняет точные initials/punctuation spans. Копирование Razdel отклонено как лишняя production source fork.

### 3. PERSON matcher является прямым набором maximal finite sequences

Matcher сканирует annotated tokens слева направо и пробует правила от наиболее специфичных/длинных к коротким:

1. `first patronymic surname`;
2. `surname first patronymic`;
3. `surname initial "." initial "."`;
4. `initial "." initial "." surname`;
5. `first surname`;
6. `surname first`.

Whitespace допускается только как gap между соседними word/initial components; точки initials входят в span, другая punctuation завершает candidate. При нескольких правилах на одном start выбирается максимальный accepted end; после принятия scanner продолжает после span, поэтому detector сам не возвращает вложенные PERSON candidates.

Role compatibility задаётся пересечением маленьких form-class bits (например, masculine nominative/dative/instrumental и необходимые feminine classes). First-name forms берутся из compact project-authored exact table; patronymics — из exact accepted forms плюс консервативных suffix classes; surname forms — из project-authored roots/forms и ограниченных русских surname suffix classes. Вся sequence обязана иметь совместимый form class, кроме initials rules, где сильным сигналом служат две буквы с точками и surname.

Альтернатива — generic NFA/Earley/grammar AST — отклонена ADR 0003: шесть finite sequences не требуют recursion или relation graph. Неограниченные suffix heuristics без first-name signal также отклонены из-за false positives.

### 4. Capitalization и outer boundaries являются обязательными gates

Word roles применяются только к кириллическим capitalized tokens; all-caps, lowercase, mixed alphanumeric и embedded token fragments не принимаются. Candidate не пересекает newline и не захватывает surrounding label/punctuation. Одиночные role tokens никогда не создают finding. Street-like exclusion проверяет непосредственно предшествующие tokens (`улица`, `ул.`, `проспект`, `пр-т`) для surname-only-like candidate paths; основные false positives дополнительно фиксируются negative corpus, а не расширяемым blacklist произвольного текста.

Альтернатива — принимать любые два capitalized слова — отклонена: это маскировало бы названия продуктов/проектов. Альтернатива — большой импортированный name dictionary — заблокирована R0 provenance policy.

### 5. Product corpus отделён от reference output, но связывается stable IDs

Новый `testdata/person/cases.jsonl` хранит schema version, stable id, synthetic input, ожидаемые exact byte spans либо `no_match`, category и при необходимости R0 reference case id. Он включает обязательные восемь positive и четыре negative R0 cases, а также embedded Unicode/punctuation, multiple persons, capitalization, product/project/common-word, street-like и token-boundary regressions.

Black-box corpus test запускает public `NewPersonDetector`, валидирует exact slices и вычисляет TP/FP/FN, precision, recall и exact-span rate. Для связанных R0 IDs он загружает `testdata/natasha/cases.jsonl` и `expected-python.jsonl`, проверяет pinned inputs/spans и требует объяснённую classification каждого reference-only match. Детерминированные итоги и exact revisions фиксируются в `docs/person-quality-report.md`; обычный test не запускает Python.

Live/offline reference schema проверяется существующими командами harness. Это сохраняет Natasha диагностической, а product corpus нормативным.

Альтернатива — генерировать product expected из Natasha — отклонена: intentional single-name differences и byte/codepoint semantics сделали бы reference фактической product specification.

### 6. End-to-end использует существующие resolver и Mask/Restore без special case

PERSON finding проходит общий `Resolve`; текущая priority table уже содержит `EntityPerson`. `Guard.Mask` заменяет весь resolved span одним token, `Restore` возвращает сохранённые bytes. README и example описывают фабрику и явно предупреждают: если LLM переносит token в иной падежный контекст, restore не склоняет исходное значение.

Альтернатива — сохранять normalized name и inflect при restore — отклонена как новый публичный/data contract, противоречащий opaque `TokenSet` и M4 non-goals.

## Non-Goals

- Одиночные имена/фамилии, nicknames, generic NER, organizations и coreference.
- Нормализованные person fields, lemmas, arbitrary grammemes или morphology-aware restore.
- Natasha/Yargy API compatibility, full OpenCorpora analyzer или runtime data download.
- ADDRESS rules, generic reusable parser и выделение отдельного NLP module.
- Caller-configurable dictionaries либо обещание recall за пределами versioned corpus.

## Risks / Trade-offs

- **[Risk] Маленькие authored tables ограничивают recall за пределами corpus** → фиксировать границу в quality report; расширять только новым corpus evidence и provenance review.
- **[Risk] Suffix rules принимают capitalized нарицательные сочетания** → требовать first-name/role composition и compatible form classes, держать versioned negative corpus, не принимать generic two-capitalized-word pattern.
- **[Risk] Tokenizer и initials punctuation дают off-by-byte errors** → строить span только из исходных token offsets и проверять Unicode/punctuation fixtures плюс `text[start:end]` invariant.
- **[Risk] Corpus evaluator случайно начнёт требовать Python** → Go test читает committed JSONL; Python commands остаются отдельной reference verification.
- **[Trade-off] Restore не согласует падеж после изменения контекста** → документировать literal byte-for-byte behavior и не обещать morphology semantics.

## Migration Plan

1. Добавить pinned `go-razdel`, internal annotations и PERSON rules без изменения существующих detector behavior.
2. Добавить public factory, corpus/differential/concurrency и Mask/Restore tests.
3. Обновить README/example, quality report и dependency/license evidence.
4. Проверить module graph, focused/race/full suites и pinned reference offline verification.
5. Rollback удаляет фабрику, internal PERSON runtime, dependency и новые evidence files; persisted state или consumer migration отсутствуют.
