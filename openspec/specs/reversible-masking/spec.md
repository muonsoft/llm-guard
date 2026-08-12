# Reversible Masking Specification

## Purpose

Capability локально заменяет resolved sensitive spans collision-safe placeholders и точно восстанавливает их через opaque caller-owned TokenSet без скрытого process state.

## Requirements

### Requirement: Mask выполняет полный detection и resolution pipeline
`Guard.Mask` MUST принимать context и text, вызвать configured detectors, разрешить findings общим resolver и вернуть `MaskResult` с masked text, resolved findings и новым caller-owned `TokenSet`. При отсутствии findings text MUST остаться неизменным, а пригодный для Restore TokenSet MUST оставаться caller-owned.

#### Scenario: EMAIL end-to-end masking
- **WHEN** Guard с EMAIL detector маскирует текст с корректным mailbox
- **THEN** `MaskResult.Text` содержит placeholder вместо mailbox, `Findings` содержит resolved EMAIL span, а `Tokens` позволяет восстановление

#### Scenario: Text без findings
- **WHEN** detectors не находят sensitive spans
- **THEN** Mask возвращает исходный text, пустые resolved findings и TokenSet без sensitive mappings

#### Scenario: Detection failure
- **WHEN** detector или resolver возвращает ошибку
- **THEN** Mask возвращает zero result и safe error без partial masked text или TokenSet

### Requirement: Каждый TokenSet использует collision-safe opaque namespace
Для каждого Mask invocation с findings библиотека MUST получить не менее 128 бит namespace entropy из cryptographically secure default source. Namespace MUST быть представлен без sensitive data, а generated placeholder MUST иметь редкий фиксированный envelope, namespace и монотонный occurrence counter. Если любой возможный token этого namespace уже встречается в исходном text, Mask MUST выбрать новый namespace до bounded лимита и при исчерпании вернуть safe error.

#### Scenario: Existing placeholder-like fragment
- **WHEN** input уже содержит fragment, совпадающий с token первого injected namespace
- **THEN** Mask выбирает следующий namespace и сохраняет исходный fragment без изменения

#### Scenario: Random source failure
- **WHEN** namespace source возвращает ошибку или после bounded attempts не удаётся избежать collision
- **THEN** Mask возвращает safe error без sensitive input и без partial result

#### Scenario: Независимые TokenSet
- **WHEN** два Mask invocation обрабатывают одинаковый text
- **THEN** secure default создаёт независимые namespace и tokens, а один TokenSet не восстанавливает tokens другого

### Requirement: Mask заменяет spans в reverse byte order
Mask MUST выполнять replacements от большего byte offset к меньшему, не пересчитывая UTF-8 offsets и не повреждая окружающий Unicode. Каждый resolved occurrence MUST получить отдельный token, даже если sensitive value повторяется.

#### Scenario: Mixed Unicode spans
- **WHEN** ASCII EMAIL находится между многобайтовыми Unicode fragments
- **THEN** surrounding text сохраняется byte-for-byte и только EMAIL заменяется token

#### Scenario: Повтор sensitive value
- **WHEN** один mailbox встречается в двух resolved spans
- **THEN** Mask выдаёт два разных occurrence tokens, каждый из которых отображается на одинаковое исходное value

### Requirement: Restore раскрывает только exact tokens переданного TokenSet
`Guard.Restore` MUST заменять каждое exact неизменённое вхождение known token соответствующим value, включая повтор одного token в response. Unknown token, token другого TokenSet и любой mutated token MUST оставаться без изменения и MUST NOT вызывать heuristic recovery. Replacement MUST быть однопроходным: восстановленное value не сканируется повторно как token.

#### Scenario: Round trip
- **WHEN** masked text не изменён между Mask и Restore
- **THEN** Restore возвращает исходный text byte-for-byte, включая Unicode, повторы и исходные placeholder-like fragments

#### Scenario: Known token повторён моделью
- **WHEN** response содержит один known token несколько раз
- **THEN** Restore подставляет соответствующее value во всех exact occurrences

#### Scenario: Unknown и mutated tokens
- **WHEN** response содержит unknown token или known token с изменённым символом
- **THEN** fragment остаётся без изменения и sensitive data не раскрывается

#### Scenario: Restored value похоже на другой token
- **WHEN** sensitive value само содержит token-like text
- **THEN** Restore вставляет value буквально без рекурсивной подстановки

### Requirement: TokenSet остаётся opaque и безопасно форматируется
TokenSet MUST хранить mappings только в unexported state, не предоставлять enumeration или raw-value access и реализовать redacted `String` и `GoString`. Standard formatting с `%v`, `%+v` или `%#v` и стандартный JSON serialization path MUST NOT раскрывать values, tokens или namespace.

#### Scenario: Formatting не раскрывает PII
- **WHEN** TokenSet форматируется через `fmt` с обычными и Go-syntax verbs
- **THEN** output содержит только redacted summary без mailbox, token и namespace

#### Scenario: JSON не раскрывает mappings
- **WHEN** TokenSet или MaskResult сериализуется через `encoding/json`
- **THEN** serialized output не содержит sensitive values, tokens или namespace

### Requirement: Mask и Restore уважают context и concurrent use
Mask и Restore MUST проверять caller context, не использовать глобальное изменяемое состояние и быть безопасными для concurrent calls одного Guard. Injectable namespace source MUST быть сериализован внутри Guard; caller, заменяющий secure default, MUST явно владеть качеством entropy.

#### Scenario: Cancel до Mask
- **WHEN** context отменён до Mask
- **THEN** detectors и random source не запускаются, а ошибка проверяется как caller context error

#### Scenario: Cancel до Restore
- **WHEN** context отменён до Restore
- **THEN** replacement не выполняется, а ошибка проверяется как caller context error

#### Scenario: Concurrent Mask
- **WHEN** несколько goroutines вызывают Mask одного Guard
- **THEN** каждый получает независимый TokenSet без data race или cross-restore

### Requirement: Invalid restore state отклоняется безопасно
Restore MUST отклонять nil TokenSet и invalid UTF-8 response проверяемыми safe errors; error strings, wrapping attributes и typed fields MUST NOT содержать response, token, namespace или sensitive mapping.

#### Scenario: Nil TokenSet
- **WHEN** caller вызывает Restore с nil TokenSet
- **THEN** он получает проверяемую invalid-token-set error без panic

#### Scenario: Invalid UTF-8 response
- **WHEN** Restore получает invalid UTF-8 string
- **THEN** он возвращает invalid-text error без раскрытия содержимого
