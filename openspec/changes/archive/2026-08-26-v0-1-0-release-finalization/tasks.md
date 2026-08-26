## 1. Core

- [x] 1.1 Повысить `go` directive и current package documentation с Go 1.26.2 до Go 1.26.6, не меняя runtime API и dependency graph.
- [x] 1.2 Обновить exact-minimum pins во всех обычных, race, evaluation, fuzz, license, external и release GitHub workflows; сохранить `stable` как forward signal.

## 2. Audit

- [x] 2.1 Добавить в `scripts/release-check.sh` отдельный `vuln` mode с pinned `golang.org/x/vuln/cmd/govulncheck@v1.7.0`, явным online/cache contract и ненулевым exit при findings.
- [x] 2.2 Добавить обязательный CI vulnerability job на Go 1.26.6, использующий тот же local script mode без дублирования scanner command.
- [x] 2.3 Убедиться, что full mode остаётся network-free и не запускает vulnerability scan, а `vuln` не создаёт tag, push, upload или release side effects.

## 3. Tests

- [x] 3.1 Проверить shell syntax/help и отдельно выполнить `GOTOOLCHAIN=go1.26.6 ./scripts/release-check.sh vuln` с результатом без findings.
- [x] 3.2 Выполнить `GOTOOLCHAIN=go1.26.6 go test ./...`, `go vet ./...` и `go test -race ./...`.
- [x] 3.3 Выполнить `GOTOOLCHAIN=go1.26.6 ./scripts/release-check.sh` и подтвердить consumer/license/evaluation/fuzz/benchmark gates без изменения module inventory.

## 4. Docs

- [x] 4.1 Согласованно обновить current minimum Go и security-gate guidance в `README.md`, `README.ru.md`, GoDoc и compatibility policy, не переписывая historical measurement metadata.
- [x] 4.2 Обновить release checklist и readiness matrix: готовность требует green offline full mode плюс отдельный pinned vulnerability scan на exact minimum.
- [x] 4.3 Финализировать planned `v0.1.0` changelog scope, сохранив явное утверждение, что tag и GitHub Release ещё не опубликованы.

## 5. Verification

- [x] 5.1 Прогнать `openspec validate v0-1-0-release-finalization --strict --no-interactive`, `openspec validate --all --strict --no-interactive` и `git diff --check`; передать evidence оркестратору для sync/archive.
