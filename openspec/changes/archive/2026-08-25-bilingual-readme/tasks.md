## 1. Docs

- [x] 1.1 Переписать `README.md` как concise English OSS entry point с language switch, точным pre-release install/status boundary и одним проверяемым quick start.
- [x] 1.2 Добавить естественный русский peer translation `README.ru.md` с тем же обязательным набором product, security и limitation claims.
- [x] 1.3 Свести supported detectors, defaults, operational properties и дальнейшую документацию к компактным таблицам/ссылкам без milestone-oriented navigation.

## 2. Audit

- [x] 2.1 Сверить обе версии с `example_test.go`, `doc.go`, public constructors, `docs/known-limitations.md` и release status; устранить неподтверждённые claims.
- [x] 2.2 Проверить двусторонние language links, все относительные repository links и смысловой паритет двух README.

## 3. Tests

- [x] 3.1 Выполнить `go test ./...`, `go vet ./...`, `go test -race ./...` и `./scripts/release-check.sh consumer`; зафиксировать результаты без публикации release artifacts.
