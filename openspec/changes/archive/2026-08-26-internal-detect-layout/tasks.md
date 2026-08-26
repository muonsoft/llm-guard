## 1. Core

- [x] 1.1 Создать `internal/detect` со `Span` и overlap helpers без импорта `llmguard`.
- [x] 1.2 Перенести `detectutil.go`, UTF-8 rune helpers и `contextmatch.go` в `internal/detect`.
- [x] 1.3 Добавить корневой unexported `spanDetector` и `builtin.go` constructors.

## 2. Detectors

- [x] 2.1 Перенести structured PII detectors (email, phone, IP, URL, INN, SNILS, passport, bank card, bank account, date of birth) в `internal/detect`.
- [x] 2.2 Перенести secret detectors (JWT, PEM, API key, DSN) в `internal/detect`.
- [x] 2.3 Подключить PERSON/ADDRESS через тот же адаптер и `internal/nlp`.
- [x] 2.4 Удалить старые корневые файлы detectors и helpers.

## 3. Tests

- [x] 3.1 Обновить whitebox fuzz helpers, чтобы mailbox/boundary checks использовали `internal/detect`.
- [x] 3.2 Прогнать `go test ./...` и `go vet ./...`.
- [x] 3.3 Прогнать `go test -race ./...`.

## 4. Docs

- [x] 4.1 Зафиксировать layout в `docs/adr/0005-internal-detect-layout.md`.
- [x] 4.2 Добавить запись в `CHANGELOG.md` Unreleased.
