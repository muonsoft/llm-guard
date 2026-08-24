# oss-distribution Specification

## Purpose

Capability определяет проверяемую OSS-дистрибуцию llm-guard: подключение внешним
Go consumer, compatibility и supply-chain disclosure, CI gates и безопасную
подготовку релиза без неявной публикации.

## Requirements

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

### Requirement: Публичное позиционирование precision-oriented prefilter
Public documentation SHALL называть llm-guard локальным precision-oriented prefilter и MUST объяснять, что library снижает риск передачи поддерживаемых PII/secrets в LLM, но не заменяет high-recall DLP, generic NER или domain-specific security review. Documentation MUST различать documented supported forms, intentional conservative exclusions и неизвестные форматы.

#### Scenario: Consumer оценивает применимость library
- **WHEN** consumer читает README, package documentation или release notes перед интеграцией
- **THEN** он видит prefilter guarantee, примеры поддерживаемого scope, риск false negatives и рекомендацию дополнительного контроля для high-risk use cases

#### Scenario: Публикуется benchmark result
- **WHEN** repository показывает contract либо exposure metrics
- **THEN** result явно обозначает profile, corpus scope и limitations и не называется доказательством полной DLP-защиты

### Requirement: Release evidence разделяет обязательные и диагностические suites
Release-readiness evidence MUST отдельно показывать статус обязательного offline conformance/lifecycle gate и результат pinned external contract/exposure suites. External report MUST быть привязан к exact git commit, source manifests, mapping policy и threshold version; отсутствие актуального external report MUST блокировать утверждение заявленного benchmark evidence, но MUST NOT добавлять network dependency обычному consumer или стандартному PR test run.

#### Scenario: Release candidate имеет полный evaluation evidence
- **WHEN** maintainer собирает readiness matrix для release commit
- **THEN** matrix содержит зелёный offline gate, воспроизводимую external command, pinned source metadata, contract/exposure results и список известных unsupported classes

#### Scenario: External report относится к другому commit или mapping
- **WHEN** report commit, mapping policy либо source manifest не совпадает с release candidate
- **THEN** report считается stale и не используется как readiness evidence до повторного запуска

### Requirement: Benchmark data имеет безопасную distribution boundary
Production Go module MUST NOT зависеть от external benchmark runtime, downloader, external corpus или optional NLP tooling. Repository MUST документировать, какие manifests, generated fixtures и reports распространяются, какие sources загружаются только явно, и какие license/attribution obligations применимы; реальные PII и действующие credentials MUST NOT включаться в public fixtures и reports.

#### Scenario: Consumer устанавливает Go module
- **WHEN** consumer импортирует `github.com/muonsoft/llm-guard`
- **THEN** external datasets, download tools, Python runtime и benchmark cache не входят в runtime dependency graph

#### Scenario: Maintainer добавляет новый benchmark source
- **WHEN** source предлагается для external либо generated suite
- **THEN** до включения фиксируются revision/digest, provenance, license, attribution, data-safety review и выбранная distribution strategy

