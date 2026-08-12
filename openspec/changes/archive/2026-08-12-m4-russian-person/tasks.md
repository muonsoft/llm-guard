## 1. Core runtime

- [x] 1.1 Добавить `github.com/muonsoft/go-razdel` на audited pseudo-version `v0.0.0-20260425122647-5cd53c7a1d02` и подтвердить checksum в `go.sum`.
- [x] 1.2 Реализовать internal token adapter с исходными UTF-8 byte spans, folded form, closed token kind и shape flags.
- [x] 1.3 Добавить project-authored immutable role/form tables для обязательных first-name, patronymic и surname форм без импортированных словарей.
- [x] 1.4 Реализовать bounded compatible-form classes и проверки capitalization, whitespace, punctuation, newline и outer token boundaries.
- [x] 1.5 Покрыть internal token/annotation/morphology contracts focused unit tests, включая Unicode offsets и cancellation-safe iteration points.

## 2. Detectors

- [x] 2.1 Реализовать immutable `NewPersonDetector() Detector` со стабильным именем, `EntityPerson` и safe context behavior.
- [x] 2.2 Реализовать maximal finite rules для first-last, last-first, first-patronymic-last и last-first-patronymic.
- [x] 2.3 Реализовать формы surname-initials и initials-surname с включением точек в единый span.
- [x] 2.4 Реализовать детерминированный scan нескольких PERSON occurrences без nested findings.
- [x] 2.5 Добавить detector tests для mandatory direct, declined и initials forms с точными byte slices.
- [x] 2.6 Добавить conservative regression tests для isolated roles, lowercase/common words, product/project names, street-like contexts и embedded token fragments.

## 3. Masking и pipeline

- [x] 3.1 Добавить end-to-end Guard Detect/Resolve/Mask/Restore test с одним token на полный PERSON span и byte-for-byte round trip.
- [x] 3.2 Добавить overlap regression, подтверждающий существующую resolver priority для PERSON без special-case pipeline logic.
- [x] 3.3 Добавить concurrent detector/Guard/Mask tests для общего immutable instance и независимых TokenSet.

## 4. Audit и corpus evidence

- [x] 4.1 Создать versioned synthetic `testdata/person/cases.jsonl` с обязательными R0 positive/negative IDs и расширенными boundary/false-positive cases.
- [x] 4.2 Реализовать black-box PERSON corpus evaluation с exact-span TP/FP/FN, precision, recall и exact-span rate.
- [x] 4.3 Связать differential cases с pinned `testdata/natasha` fixtures и проваливать необъяснённые reference-only differences.
- [x] 4.4 Проверить offline reference schema командой `python3 tools/natasha-reference/reference.py verify --offline --cases testdata/natasha/cases.jsonl --expected testdata/natasha/expected-python.jsonl`.
- [x] 4.5 Проверить production dependency graph через `go list -deps ./...` и `go mod graph`, подтвердив отсутствие Python/ML/morphology data dependencies.
- [x] 4.6 Обновить `docs/natasha-license-inventory.md`, если фактическая pseudo-version/checksum evidence требует уточнения, не добавляя неаудированные sources.

## 5. Tests

- [x] 5.1 Выполнить focused suite `go test ./... -run 'Test(Person|Name|Morph|Token)'`.
- [x] 5.2 Выполнить race-focused suite `go test -race ./... -run 'TestPerson'`.
- [x] 5.3 Выполнить `go test ./...`, `go vet ./...` и `go test -race ./...`.
- [x] 5.4 Повторить PERSON corpus evaluation и подтвердить нулевые FP/FN mandatory corpus после full suite.

## 6. Docs

- [x] 6.1 Обновить README built-in detector list, usage и supported PERSON boundary.
- [x] 6.2 Добавить runnable PERSON Mask/Restore example с исходным UTF-8 round trip.
- [x] 6.3 Создать `docs/person-quality-report.md` с corpus version, pinned reference revisions, metrics и intentional differences.
- [x] 6.4 Документировать limitation literal restore без morphology-aware согласования изменённого LLM-контекста.

## 7. Verification

- [x] 7.1 Выполнить `openspec validate "m4-russian-person" --strict --no-interactive`.
- [x] 7.2 Выполнить `openspec validate --specs --strict --no-interactive` после sync delta spec.
- [x] 7.3 Сверить implementation каждого requirement/scenario с `specs/russian-person/spec.md` и сохранить verification evidence без PII.
