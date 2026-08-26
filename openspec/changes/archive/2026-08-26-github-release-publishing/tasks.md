## 1. Release metadata tooling

- [x] 1.1 Добавить `scripts/prepare-release.sh` с SemVer validation, planned-section и `[Unreleased]` promotion, `--check-only`, `--require-final` и идемпотентным повторным запуском.
- [x] 1.2 Добавить изолированный shell test для planned, future Unreleased, invalid input, strict finalized-tag и idempotency сценариев.

## 2. GitHub Actions

- [x] 2.1 Заменить dry-run-only workflow на `Release` workflow с manual dispatch и prepared-tag path, строгим version/ref resolution и сериализацией запусков.
- [x] 2.2 Ограничить default permissions до `contents: read`, выдать `contents: write` только publication job после `needs: validate` и исключить shell interpolation пользовательского input.
- [x] 2.3 В manual path проверить актуальный `main`, финализировать changelog commit, push-нуть его и создать GitHub Release/tag на точном SHA; в tag path запретить изменение history.
- [x] 2.4 Подключить release-script syntax/tests к обычному CI, чтобы regression обнаруживался до запуска Release.

## 3. Public release documentation

- [x] 3.1 Обновить английский и русский README state-neutral описанием `v0.1.0` и explicit Release workflow без устаревающего утверждения «ещё не опубликован» внутри tag.
- [x] 3.2 Обновить CHANGELOG release workflow evidence и сохранить planned section готовой к автоматической датировке.
- [x] 3.3 Обновить release checklist, compatibility policy и readiness matrix: green CI → manual Release dispatch → automated changelog commit/tag/GitHub Release → clean-module `go get` verification.

## 4. Verification

- [x] 4.1 Выполнить `bash -n scripts/prepare-release.sh scripts/prepare-release-test.sh` и `bash scripts/prepare-release-test.sh`.
- [x] 4.2 Выполнить `go run github.com/rhysd/actionlint/cmd/actionlint@v1.7.7`, `git diff --check` и проверку OpenSpec change в strict mode.
- [x] 4.3 Выполнить `GOTOOLCHAIN=go1.26.6 ./scripts/release-check.sh` и `GOTOOLCHAIN=go1.26.6 ./scripts/release-check.sh vuln` без публикации.
