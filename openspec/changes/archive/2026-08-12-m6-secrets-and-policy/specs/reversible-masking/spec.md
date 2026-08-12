## MODIFIED Requirements

### Requirement: Mask выполняет полный detection и resolution pipeline
`Guard.Mask` MUST принимать context и text, вызвать configured detectors, разрешить findings общим resolver, применить immutable action configuration и при отсутствии block вернуть `MaskResult` с transformed text, полным resolved findings и новым caller-owned `TokenSet`. Mask findings MUST заменяться collision-safe placeholders, allow findings MUST оставаться неизменными без token mapping. При отсутствии masked findings text MUST сохранять все исходные bytes, а пригодный для Restore пустой TokenSet MUST оставаться caller-owned. При block либо любом detection/resolution/policy failure Mask MUST возвращать zero result и safe error без partial transformed text или TokenSet.

#### Scenario: EMAIL end-to-end masking
- **WHEN** Guard с EMAIL detector маскирует текст с корректным mailbox и default policy
- **THEN** `MaskResult.Text` содержит placeholder вместо mailbox, `Findings` содержит resolved EMAIL span, а `Tokens` позволяет восстановление

#### Scenario: Text без findings
- **WHEN** detectors не находят sensitive spans
- **THEN** Mask возвращает исходный text, пустые resolved findings и TokenSet без sensitive mappings

#### Scenario: Allowed finding
- **WHEN** resolved finding настроен как allow и blocked findings отсутствуют
- **THEN** Mask сохраняет его исходные bytes, включает finding в resolved result и не создаёт для него mapping

#### Scenario: Blocked finding
- **WHEN** хотя бы один resolved finding настроен как block
- **THEN** Mask возвращает zero result и проверяемую safe block error без entropy read или partial output

#### Scenario: Detection failure
- **WHEN** detector или resolver возвращает ошибку
- **THEN** Mask возвращает zero result и safe error без partial masked text или TokenSet
