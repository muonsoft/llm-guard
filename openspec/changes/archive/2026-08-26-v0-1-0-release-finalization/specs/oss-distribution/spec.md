## MODIFIED Requirements

### Requirement: Public API и compatibility policy
Проект SHALL публиковать pre-1.0 semantic versioning policy, список
поддерживаемых Go versions и known limitations MVP; минимальная поддерживаемая
Go toolchain MUST содержать исправления всех известных достижимых уязвимостей
standard library на момент release-readiness review. Review MUST подтверждать,
что финализация `v0.1.0` не вводит незапланированное breaking изменение public API.

#### Scenario: Изменение оценивается до v1.0
- **WHEN** maintainer готовит изменение exported API после `v0.1.0`
- **THEN** compatibility policy объясняет допустимые semver последствия и требует отражения user-visible изменения в changelog

#### Scenario: Consumer видит границы MVP
- **WHEN** пользователь читает public documentation перед подключением library
- **THEN** он видит поддерживаемую toolchain, security boundary и ограничения PERSON, ADDRESS, placeholders, false negatives и отсутствующих post-MVP возможностей

#### Scenario: Минимальная toolchain имеет security fix
- **WHEN** maintainer проверяет release candidate официальным vulnerability scanner на exact minimum Go version
- **THEN** отсутствие достижимых findings является обязательным evidence, а найденная standard-library vulnerability блокирует release до повышения minimum на исправленную patch-version

### Requirement: Воспроизводимые CI quality gates
CI SHALL выполнять на заявленной Go matrix детерминированные unit/example tests и
vet, а отдельные gates MUST воспроизводить race, полный evaluation regression,
bounded smoke для каждого обязательного fuzz target masking/restore/resolver/custom
regexp boundary и официальный vulnerability scan на exact minimum toolchain.
Vulnerability tool version MUST быть pinned, а команды и target names SHALL быть
доступны локально.

#### Scenario: Pull request проходит обычный CI
- **WHEN** pull request направлен в main
- **THEN** CI выполняет test/example и vet на всей поддерживаемой Go matrix и отдельный race gate без внешнего сервиса

#### Scenario: Fuzz smoke ограничен и адресен
- **WHEN** запускается fuzz smoke workflow или локальная release проверка
- **THEN** каждый обязательный target запускается по точному имени с bounded fuzz time и сбой любого target делает gate неуспешным

#### Scenario: Quality regression блокирует readiness
- **WHEN** evaluation corpus получает FP, FN или неполное entity coverage
- **THEN** quality gate завершается ненулевым статусом и release dry-run считается неуспешным

#### Scenario: Vulnerability scan блокирует readiness
- **WHEN** pinned vulnerability tool на exact minimum Go version находит достижимую vulnerability в project или standard library
- **THEN** отдельный CI и local security gate завершаются ненулевым статусом и release readiness не считается зелёной

### Requirement: Release dry-run и readiness evidence
Проект SHALL предоставлять локально воспроизводимый network-free release dry-run,
который проверяет build/test/vet/race/fuzz/evaluation/examples/licenses/docs без
tag, push, artifact upload или release publication. Отдельный reproducible
vulnerability gate MUST использовать pinned scanner и MAY требовать network либо
заранее заполненный module/vulnerability cache. Финальная readiness matrix MUST
требовать оба gate, сопоставлять Definition of Done и known limitations с
evidence и сравнивать quality/benchmark results с M7 baseline без превращения
benchmark в SLO.

#### Scenario: Dry-run безопасен по умолчанию
- **WHEN** maintainer запускает documented release dry-run в обычном checkout
- **THEN** все offline gates выполняются без изменения git refs и без сетевой публикации или загрузки artifacts

#### Scenario: Security gate отделён от offline dry-run
- **WHEN** maintainer проверяет release candidate перед approval
- **THEN** он отдельно запускает documented pinned vulnerability scan на exact minimum Go version и получает явное evidence об отсутствии достижимых findings

#### Scenario: Release выполняется только отдельно
- **WHEN** offline dry-run, vulnerability gate и readiness matrix зелёные
- **THEN** repository сообщает готовность `v0.1.0`, но tag, push и GitHub release требуют отдельного явного действия maintainer

#### Scenario: M7 baseline сравнивается воспроизводимо
- **WHEN** выполняется финальная quality/benchmark проверка M8
- **THEN** отчёт содержит точные команды, актуальный quality outcome и сопоставление стабильных benchmark names с M7, явно отмечая hardware variance и отсутствие SLO
