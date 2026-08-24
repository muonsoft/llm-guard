## ADDED Requirements

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
