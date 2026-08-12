## 1. Core и public package boundary

- [x] 1.1 Инвентаризировать exported API и подтвердить отсутствие breaking изменений runtime surface в M8.
- [x] 1.2 Уточнить package GoDoc: назначение, concurrency, byte spans, caller-owned TokenSet и security boundary.
- [x] 1.3 Подготовить минимальный external-consumer source только на canonical public import path.
- [x] 1.4 Добавить временный external-module check с local replace и гарантированным cleanup.
- [x] 1.5 Проверить внешний flow командой `./scripts/release-check.sh consumer`.

## 2. Tests и CI automation

- [x] 2.1 Добавить единый side-effect-free `scripts/release-check.sh` с документированными modes и fail-fast поведением.
- [x] 2.2 Реализовать `license` mode, сверяющий `go list -m all` с committed inventory/notices.
- [x] 2.3 Реализовать `fuzz` mode с точными targets `FuzzMaskRestoreRoundTrip`, `FuzzResolveInvariants` и `FuzzCustomRegexpDetectorInvariants` и bounded временем.
- [x] 2.4 Реализовать full dry-run с format/diff hygiene, test, vet, race, consumer, evaluation, license, fuzz и benchmark smoke gates.
- [x] 2.5 Расширить pull-request CI matrix до minimum Go `1.26.2` и `stable` для tests/vet.
- [x] 2.6 Добавить отдельные CI jobs для race, external consumer, evaluation regression и bounded fuzz smoke без внешних сервисов.
- [x] 2.7 Добавить manual/version-tag release-check workflow только с read permissions и без publish/upload steps.
- [x] 2.8 Проверить shell syntax командой `sh -n scripts/release-check.sh`.
- [x] 2.9 Проверить exact-target fuzz smoke командой `./scripts/release-check.sh fuzz`.

## 3. Audit, dependencies и provenance

- [x] 3.1 Сверить direct/indirect modules из `go.mod`/`go.sum`, их версии, лицензии и production/test роль.
- [x] 3.2 Сверить project-authored detector tables, corpora, fuzz seeds и generated/reference fixtures с известным provenance.
- [x] 3.3 Сверить Python Natasha/Yargy/OpenCorpora tooling и зафиксировать reference-only/non-redistributed boundary.
- [x] 3.4 Добавить `THIRD_PARTY_NOTICES` с distribution obligations для shipped dependencies и ссылками на полный audit.
- [x] 3.5 Добавить полный dependency/data license inventory с blocking policy для неизвестного source/license.
- [x] 3.6 Выполнить consistency gate командой `./scripts/release-check.sh license`.
- [x] 3.7 Подтвердить отсутствие новых runtime dependencies командой `go list -m all` и review diff `go.mod go.sum`.

## 4. Docs, security и release readiness

- [x] 4.1 Добавить `SECURITY.md` с private reporting guidance, supported scope и запретом real PII/secrets в публичных reports.
- [x] 4.2 Добавить `CONTRIBUTING.md` с OpenSpec, test, synthetic fixture и dependency/data provenance требованиями.
- [x] 4.3 Добавить compatibility/versioning policy для pre-1.0 API, minimum Go и forward checks.
- [x] 4.4 Добавить `CHANGELOG.md` с unreleased policy и составом планируемого `v0.1.0` без заявления о публикации.
- [x] 4.5 Добавить release checklist с preflight, dry-run, manual approval и отдельными tag/push/release шагами.
- [x] 4.6 Обновить README: OSS-ready status, supported Go, external consumer path, policies, limitations и release boundary.
- [x] 4.7 Сформировать финальную MVP Definition of Done matrix со ссылками на implementation/test/spec evidence.
- [x] 4.8 Опубликовать полный known-limitations список из MVP plan, включая placeholder mutation и отсутствие zero-FN guarantee.
- [x] 4.9 Сформировать M8 quality/benchmark comparison с точными командами, M7 baseline и no-SLO/hardware caveat.

## 5. Verification

- [x] 5.1 Выполнить `go test ./...` и `go vet ./...`.
- [x] 5.2 Выполнить `go test -race ./...`.
- [x] 5.3 Выполнить `go run ./cmd/llmguard-eval -corpus ./testdata/evaluation/cases.jsonl -format json -fail-on-regression`.
- [x] 5.4 Выполнить `go test ./... -run '^$' -bench . -benchmem` и сравнить стабильные benchmark names с M7.
- [x] 5.5 Выполнить полный `./scripts/release-check.sh` и подтвердить отсутствие tag/push/publish side effects.
- [x] 5.6 Выполнить `openspec validate m8-oss-stabilization --strict --no-interactive` и `git diff --check`.
