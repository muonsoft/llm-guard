## Purpose

Capability определяет проверяемую OSS-дистрибуцию llm-guard: подключение внешним
Go consumer, compatibility и supply-chain disclosure, CI gates и безопасную
подготовку релиза без неявной публикации.

## ADDED Requirements

### Requirement: Внешнее подключение и package documentation
Repository SHALL документировать canonical import path, поддерживаемую Go
toolchain и рабочий embedded `Detect → Mask → Restore` flow, а отдельная проверка
MUST компилировать и запускать пример из внешнего Go module без доступа к
внутренним пакетам repository.

#### Scenario: Новый consumer запускает documented flow
- **WHEN** consumer создаёт отдельный module, подключает `github.com/muonsoft/llm-guard` и использует документированный public API
- **THEN** module компилируется и executable example подтверждает обратимый `Mask → Restore` flow

#### Scenario: Внешний пример не зависит от repository internals
- **WHEN** выполняется consumer compile check вне package tree llm-guard
- **THEN** исходник импортирует только публичный module path и не использует `internal/`, test exports или непубличные символы

### Requirement: Public API и compatibility policy
Проект SHALL публиковать pre-1.0 semantic versioning policy, список
поддерживаемых Go versions и known limitations MVP; release-readiness review MUST
подтверждать, что M8 не вводит незапланированное breaking изменение public API.

#### Scenario: Изменение оценивается до v1.0
- **WHEN** maintainer готовит изменение exported API после `v0.1.0`
- **THEN** compatibility policy объясняет допустимые semver последствия и требует отражения user-visible изменения в changelog

#### Scenario: Consumer видит границы MVP
- **WHEN** пользователь читает public documentation перед подключением library
- **THEN** он видит поддерживаемую toolchain, security boundary и ограничения PERSON, ADDRESS, placeholders, false negatives и отсутствующих post-MVP возможностей

### Requirement: Воспроизводимые CI quality gates
CI SHALL выполнять на заявленной Go matrix детерминированные unit/example tests и
vet, а отдельные gates MUST воспроизводить race, полный evaluation regression и
bounded smoke для каждого обязательного fuzz target masking/restore/resolver/custom
regexp boundary. Команды и target names SHALL быть доступны локально.

#### Scenario: Pull request проходит обычный CI
- **WHEN** pull request направлен в main
- **THEN** CI выполняет test/example и vet на всей поддерживаемой Go matrix и отдельный race gate без внешнего сервиса

#### Scenario: Fuzz smoke ограничен и адресен
- **WHEN** запускается fuzz smoke workflow или локальная release проверка
- **THEN** каждый обязательный target запускается по точному имени с bounded fuzz time и сбой любого target делает gate неуспешным

#### Scenario: Quality regression блокирует readiness
- **WHEN** evaluation corpus получает FP, FN или неполное entity coverage
- **THEN** quality gate завершается ненулевым статусом и release dry-run считается неуспешным

### Requirement: Лицензии, notices и provenance
Distribution MUST содержать project license, inventory всех production и
redistributed code/data dependencies и требуемые third-party notices. Любой
источник с неизвестной или несовместимой лицензией, provenance или attribution
MUST блокировать release readiness.

#### Scenario: Production dependency имеет disclosure
- **WHEN** maintainer сверяет `go.mod`, embedded/project-authored tables и распространяемые tooling/data artifacts
- **THEN** inventory фиксирует источник, версию, лицензию, runtime/distribution роль и требуемое notice решение для каждого элемента

#### Scenario: Неизвестный data source найден перед релизом
- **WHEN** audit обнаруживает словарь, fixture или generated data без подтверждённых source и license
- **THEN** release checklist остаётся заблокированным до удаления источника или документирования совместимой provenance и attribution

### Requirement: Публичные security и contribution policies
Repository SHALL публиковать private vulnerability reporting guidance,
поддерживаемую security scope, contribution/test requirements и запрет реальных
PII/credentials в issues, fixtures и reports.

#### Scenario: Исследователь сообщает vulnerability
- **WHEN** исследователь находит возможную утечку PII или bypass secret detector
- **THEN** `SECURITY.md` направляет его в непубличный канал и запрещает раскрывать реальные sensitive values в публичном issue

#### Scenario: Contributor отправляет изменение detector или data
- **WHEN** contribution меняет detector, corpus, dictionary или dependency
- **THEN** contribution guide требует targeted tests, provenance/license evidence и безопасные synthetic fixtures

### Requirement: Release dry-run и readiness evidence
Проект SHALL предоставлять локально воспроизводимый release dry-run, который
проверяет build/test/vet/race/fuzz/evaluation/examples/licenses/docs без tag,
push, artifact upload или release publication. Финальная readiness matrix MUST
сопоставлять Definition of Done и known limitations с evidence и сравнивать
quality/benchmark results с M7 baseline без превращения benchmark в SLO.

#### Scenario: Dry-run безопасен по умолчанию
- **WHEN** maintainer запускает documented release dry-run в обычном checkout
- **THEN** все gates выполняются без изменения git refs и без сетевой публикации или загрузки artifacts

#### Scenario: Release выполняется только отдельно
- **WHEN** dry-run и readiness matrix зелёные
- **THEN** repository сообщает готовность `v0.1.0`, но tag, push и GitHub release требуют отдельного явного действия maintainer

#### Scenario: M7 baseline сравнивается воспроизводимо
- **WHEN** выполняется финальная quality/benchmark проверка M8
- **THEN** отчёт содержит точные команды, актуальный quality outcome и сопоставление стабильных benchmark names с M7, явно отмечая hardware variance и отсутствие SLO
