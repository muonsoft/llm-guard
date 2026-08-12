## Context

См. motivation в `proposal.md`. После M7 module уже имеет полный runtime scope,
race-safe core, executable package examples, evaluation CLI и benchmark baseline.
Текущий CI запускает `test`, `vet` и `race` только через `go-version-file`, а
distribution policy, внешний consumer gate, fuzz smoke, release dry-run и единый
license/notices audit отсутствуют. `go.mod` задаёт минимум `go 1.26.2`; runtime
имеет две direct production dependencies (`muonsoft/errors`, `go-razdel`), а
Python/Natasha graph остаётся development reference и не входит в module.

Изменение cross-cutting: оно затрагивает Go examples/package docs, GitHub Actions,
локальные release checks и OSS/security/supply-chain документы. Требования
зафиксированы в `specs/oss-distribution/spec.md`.

## Goals / Non-Goals

**Goals:**

- Сделать minimum-toolchain и forward-compatibility checks явными и локально
  воспроизводимыми.
- Проверять public module из настоящей внешней module boundary, а не только из
  `package llmguard_test` внутри repository.
- Иметь один side-effect-free release dry-run, на который ссылаются CI и release
  checklist.
- Закрыть code/data license и provenance evidence без добавления runtime
  dependencies или копирования reference dictionaries.
- Зафиксировать финальную DoD/limitations и regression evidence для `v0.1.0`.

**Non-Goals:**

- Не менять exported detector/masking/observer API и семантику runtime.
- Не добавлять OpenAI adapters, proxy, persistence, exporter, ML/NER или новые
  entity types.
- Не создавать tag, не выполнять push, artifact upload или GitHub release.
- Не обещать long-term support для pre-1.0 API и не превращать benchmarks в SLO.

## Decisions

### 1. Проверять minimum Go и moving stable раздельно

Основной CI использует matrix из точного minimum `1.26.2` и `stable`: minimum
доказывает заявленный нижний boundary, stable заранее выявляет forward fallout.
Race и более дорогие release gates выполняются один раз на minimum, чтобы не
дублировать стоимость без нового acceptance signal.

Альтернатива — только `go-version-file` — отклонена, потому что она не делает
forward boundary видимым. Понижение `go` directive ради широкой matrix отклонено:
это отдельное compatibility решение без evidence в M8.

### 2. External consumer — отдельный временный module

Repository хранит небольшой synthetic consumer source, а проверочная команда
копирует его во временный каталог, создаёт независимый `go.mod`, связывает
текущий checkout через local `replace` и запускает программу. Это доказывает
canonical import/public-only boundary и не публикует module.

Альтернатива — полагаться только на `package llmguard_test` examples — отклонена:
они компилируются рядом с исходным module и не ловят ошибки инструкции подключения.
Постоянный nested module отклонён как лишний второй dependency lock.

### 3. Один shell entrypoint для release dry-run, узкие reusable modes для CI

Versioned script предоставляет полную последовательность release checks и
опциональные узкие modes для consumer/fuzz/license gates. Full mode выполняет
format/diff hygiene, tests, vet, race, external consumer, evaluation, bounded
exact-target fuzz smoke, license/notices consistency и benchmark smoke. Скрипт
использует temporary directories с cleanup и не содержит git/network publication
commands.

Альтернатива — дублировать длинные команды только в YAML и Markdown — отклонена
из-за drift между локальной и CI проверкой. Новый task runner/dependency отклонён:
plain POSIX shell и Go commands достаточны для MVP.

### 4. CI и release dry-run разделены по стоимости и назначению

Pull-request workflow сохраняет быстрый test/vet matrix, отдельные race,
evaluation, consumer и bounded fuzz jobs. Отдельный release-check workflow
запускается вручную и на version-like tag только как dry-run; он не получает
`contents: write`, release environment или upload/publish steps. Фактический tag
и release остаются ручным отдельным действием после checklist.

Альтернатива — автоматический release на tag — отклонена как out of scope и как
риск случайной публикации до явного maintainer approval.

### 5. Distribution inventory отделяет shipped, test и reference-only graph

`THIRD_PARTY_NOTICES` перечисляет распространяемые/direct Go dependencies и их
license obligations. Расширенный inventory фиксирует indirect test modules,
project-authored tables/fixtures и Python reference tooling с явным решением о
том, что runtime module не redistributes Natasha/OpenCorpora data. Exact module
versions берутся из committed `go.mod`/`go.sum`; R0 provenance остаётся
нормативным источником NLP решений.

Альтернатива — перечислить только direct `go.mod` — отклонена: она скрывает test
graph и reference tooling. Копирование upstream license bodies без определения
distribution роли отклонено как неточный notice пакет.

### 6. Readiness evidence — versioned human-readable ledgers

Release checklist, compatibility policy, license inventory, final DoD matrix и
quality comparison хранят точные локальные команды и ссылки на существующие M7
evidence. Benchmark comparison проверяет стабильность имён и успешный smoke run,
но описывает hardware variance вместо числового pass/fail threshold.

Альтернатива — генерировать readiness только из CI logs — отклонена: ephemeral
logs не являются durable release evidence и не объясняют limitations.

## Risks / Trade-offs

- **[Risk] `stable` со временем станет новой major Go версией и выявит реальную
  несовместимость.** → Считать это forward-compatibility signal; minimum job
  остаётся нормативным, а изменение support policy требует changelog/spec review.
- **[Risk] Полный dry-run и несколько fuzz targets заметно увеличат время.** →
  Использовать bounded smoke time и не дублировать race/bench на всей matrix.
- **[Risk] Shell entrypoint хуже переносится на Windows.** → CI/release environment
  фиксируется как Linux; сами library и consumer example остаются pure Go.
- **[Risk] License inventory устареет при dependency change.** → Consistency gate
  сверяет ожидаемые module/provenance references, а contribution/release checklist
  требует ручного re-audit при изменении `go.mod`, data или tools.
- **[Risk] Локальный `replace` не доказывает доступность ещё не опубликованного
  tag.** → M8 проверяет API/module boundary до публикации; загрузка exact `v0.1.0`
  выполняется только отдельным post-tag release action.

## Migration Plan

1. Добавить checks, external consumer fixture и OSS/release документы без
   изменения runtime API.
2. Выполнить focused checks, полный release dry-run и broad repository suite.
3. Синхронизировать `oss-distribution` spec и архивировать M8.
4. После merge maintainer отдельно создаёт `v0.1.0` tag/release по checklist.

Rollback не требует data migration: workflow/scripts/docs можно отменить одним
commit; runtime module и caller-owned state не меняются.
