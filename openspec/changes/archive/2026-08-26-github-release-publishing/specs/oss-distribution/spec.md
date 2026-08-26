## MODIFIED Requirements

### Requirement: Release dry-run и readiness evidence
Проект SHALL предоставлять локально воспроизводимый network-free release dry-run,
который проверяет build/test/vet/race/fuzz/evaluation/examples/licenses/docs без
tag, push, artifact upload или release publication. Отдельный reproducible
vulnerability gate MUST использовать pinned scanner и MAY требовать network либо
заранее заполненный module/vulnerability cache. Финальная readiness matrix MUST
требовать оба gate, сопоставлять Definition of Done и known limitations с
evidence и сравнивать quality/benchmark results с M7 baseline без превращения
benchmark в SLO. После зелёных gates явный maintainer-triggered GitHub workflow
SHALL повторно валидировать выбранный release commit, финализировать release notes,
создать точный SemVer tag и опубликовать source-only GitHub Release; публикация
MUST NOT выполняться обычным push/PR CI.

#### Scenario: Dry-run безопасен по умолчанию
- **WHEN** maintainer запускает documented release dry-run в обычном checkout
- **THEN** все offline gates выполняются без изменения git refs и без сетевой публикации или загрузки artifacts

#### Scenario: Security gate отделён от offline dry-run
- **WHEN** maintainer проверяет release candidate перед approval
- **THEN** он отдельно запускает documented pinned vulnerability scan на exact minimum Go version и получает явное evidence об отсутствии достижимых findings

#### Scenario: Release выполняется только отдельно
- **WHEN** maintainer после зелёного CI вручную запускает Release workflow с валидной SemVer на актуальном `main`
- **THEN** workflow повторно выполняет обязательные release и vulnerability gates, финализирует changelog, создаёт tag на получившемся release commit и публикует GitHub Release только после успеха validation

#### Scenario: Validation блокирует публикацию
- **WHEN** версия невалидна, dispatch относится не к актуальному `main`, changelog не готов или любой обязательный gate завершается ошибкой
- **THEN** workflow не создаёт новый tag и GitHub Release

#### Scenario: Подготовленный tag публикуется без переписывания history
- **WHEN** maintainer отдельно push-ит валидный SemVer tag с уже финализированной changelog section
- **THEN** workflow валидирует tagged commit и MAY создать соответствующий GitHub Release, но MUST NOT перемещать tag или создавать release commit на `main`

#### Scenario: Publication permissions ограничены
- **WHEN** GitHub выполняет обычный CI, release validation и publication
- **THEN** repository contents доступны только для чтения до publication job, а write permission используется только после успешной validation dependency

#### Scenario: M7 baseline сравнивается воспроизводимо
- **WHEN** выполняется финальная quality/benchmark проверка M8
- **THEN** отчёт содержит точные команды, актуальный quality outcome и сопоставление стабильных benchmark names с M7, явно отмечая hardware variance и отсутствие SLO
