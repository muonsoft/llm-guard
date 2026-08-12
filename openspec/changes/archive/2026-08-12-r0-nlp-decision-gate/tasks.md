## 1. Audit upstream и лицензий

- [x] 1.1 Зафиксировать executable-compatible revisions Natasha, Yargy, Razdel reference, Python и morphology dependencies вместе с командами получения source.
- [x] 1.2 Проследить `NamesExtractor` до tokenizer, morphology, predicates, parser, interpretation и dictionaries и заполнить PERSON-часть dependency graph.
- [x] 1.3 Проследить `AddrExtractor` до тех же transitive layers и заполнить ADDRESS-часть dependency graph без неизвестных required features.
- [x] 1.4 Извлечь точный inventory grammar primitives, normalized forms и grammemes, разделив required M4/M5 и out-of-scope Natasha constructs.
- [x] 1.5 Сопоставить token contracts `go-razdel` и Yargy MorphTokenizer на punctuation, initials, hyphens и address abbreviations с сохранением byte/codepoint offset differences.
- [x] 1.6 Составить inventory Natasha dictionaries, Pymorphy/OpenCorpora-derived data и других кандидатов с revision, provenance, license, attribution и distribution constraints.
- [x] 1.7 Сравнить pure-Go morphology/generated-dictionary candidates по required coverage, binary/RAM footprint, initialization, thread safety и license.

## 2. Core architecture decisions

- [x] 2.1 Составить decision matrix direct Go rules/bounded matcher против generic Yargy/Earley runtime на основе закрытого feature inventory.
- [x] 2.2 Проверить выразимость обязательных M4/M5 grammar forms bounded matcher и сохранить failing fixture, если обнаружен blocking construct.
- [x] 2.3 Зафиксировать минимальный immutable morphology token/form contract и выбранную source/distribution strategy без публичного API.
- [x] 2.4 Создать `docs/adr/0003-nlp-runtime-boundary.md` с решениями: внешний `go-razdel`, internal runtime, direct rules, отсутствие `go-natasha` в MVP и критерии будущего выделения модуля.
- [x] 2.5 Проверить текущий production dependency baseline командами `go list -deps ./...` и `go mod graph`, чтобы reference tooling не изменил Go runtime graph.

## 3. Detectors и quality baseline

- [x] 3.1 Определить versioned JSONL schema для case id, entity, input, reference matches, spans, normalized fields и intentional-difference classification.
- [x] 3.2 Добавить synthetic PERSON sample cases: supported multi-part forms, initials, declined forms и negative single-name/street-like contexts без реальных PII.
- [x] 3.3 Добавить synthetic ADDRESS sample cases: accepted compositions, reordered abbreviations и negative standalone settlement/region/street parts без реальных PII.
- [x] 3.4 Описать precision-oriented PERSON/ADDRESS metrics и правила классификации regression, intentional product difference и unsupported out-of-scope cases.

## 4. Tests и reference tooling

- [x] 4.1 Создать изолированный `tools/natasha-reference/` с pinned Python environment и CLI генерации/проверки fixtures.
- [x] 4.2 Реализовать явное преобразование Python codepoint spans в UTF-8 byte spans и self-check `input_bytes[start:end] == matched_text`.
- [x] 4.3 Сгенерировать versioned `testdata/natasha/expected-python.jsonl` на pinned reference и добавить schema/sample validation tests.
- [x] 4.4 Добавить проверку deterministic output reference harness и документированный offline verify mode по committed fixtures.
- [x] 4.5 Подтвердить, что обычные `go test ./...` не требуют Python, network, Natasha/Yargy/Pymorphy либо runtime data download.

## 5. Docs

- [x] 5.1 Создать `docs/natasha-port-scope.md` с закрытой required-feature matrix, точными reference revisions и dependency graph для PERSON/ADDRESS.
- [x] 5.2 Добавить license inventory и attribution/distribution plan для каждого разрешённого code/data source; unresolved source оформить blocker, а не допущение.
- [x] 5.3 Документировать запуск reference harness, JSONL schema, offset conversion, quality metrics и intentional differences.
- [x] 5.4 Сверить R0 exit evidence с M4/M5 scope и убрать требования повторного решения repository boundary, tokenizer или parser family.

## 6. Verification

- [x] 6.1 Выполнить reference-проверку `python3 tools/natasha-reference/reference.py verify --cases testdata/natasha/cases.jsonl --expected testdata/natasha/expected-python.jsonl` в pinned environment.
- [x] 6.2 Выполнить focused проверки harness/schema/offsets и сохранить точные команды и результаты в R0 verification evidence.
- [x] 6.3 Выполнить `go list -deps ./...`, `go test ./...`, `go vet ./...` и `go test -race ./...` без Python в обязательном Go path.
- [x] 6.4 Выполнить `openspec validate r0-nlp-decision-gate --type change --strict --no-interactive` и `openspec validate --specs --strict --no-interactive`.
