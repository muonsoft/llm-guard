## 1. Core

- [x] 1.1 Реализовать public `EntityType`, built-in entity constants, `Finding` и безопасные программно проверяемые error contracts.
- [x] 1.2 Реализовать `Detector`, `Option`, `WithDetector` и immutable `Guard` construction с проверкой nil, имён и дубликатов.
- [x] 1.3 Реализовать concurrent `Detect` pipeline с производным context, ожиданием всех calls и детерминированным выбором failure без partial findings.
- [x] 1.4 Реализовать UTF-8 input/finding validation, нормализацию detector metadata и стабильную агрегацию по полному sort key.

## 2. Tests

- [x] 2.1 Добавить black-box public API/example tests для custom detector, empty guard и configuration failures.
- [x] 2.2 Добавить Unicode byte-span, invalid finding/input и sensitive-error redaction tests с проверкой `errors.Is/As`.
- [x] 2.3 Добавить concurrency, completion-order, multiple-error и context cancellation tests, включая race-focused сценарий одного `Guard`.

## 3. Docs и repository baseline

- [x] 3.1 Добавить package documentation и ADR о UTF-8 byte offsets и stateless core/caller-owned future state.
- [x] 3.2 Добавить минимальный GitHub Actions workflow для test, vet и race checks без внешних сервисов.

## 4. Verification

- [x] 4.1 Выполнить focused проверки `go test ./... -run 'TestGuard|TestFinding|TestDetector'` и `go test -race ./... -run 'TestGuard'`.
- [x] 4.2 Выполнить milestone проверки `go test ./...`, `go vet ./...` и `go test -race ./...`.
