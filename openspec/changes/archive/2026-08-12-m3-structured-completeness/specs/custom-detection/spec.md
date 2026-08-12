## Purpose

Capability предоставляет безопасный публичный regexp-adapter и подтверждает end-to-end расширяемость pipeline для пользовательских строковых entity и Go detectors.

## ADDED Requirements

### Requirement: CustomRegexpDetector создаётся из проверяемой конфигурации
Корневой Go package MUST экспортировать `RegexDetectorConfig` с полями `Name string`, `Entity EntityType`, `Pattern string`, `Confidence float64` и constructor, возвращающий reusable immutable `CustomRegexpDetector` либо безопасную configuration error. Constructor MUST отклонять пустые name, entity или pattern, синтаксически невалидный RE2 pattern, NaN/infinity и confidence вне `[0,1]`; публичная ошибка MUST поддерживать `errors.Is(..., ErrInvalidConfig)` и MUST NOT содержать pattern или иные config values.

#### Scenario: Валидная custom regexp config
- **WHEN** caller создаёт detector для entity `EMPLOYEE_ID` и pattern `EMP-[0-9]{6}` с confidence `0.9`
- **THEN** constructor возвращает detector со стабильным configured name без error

#### Scenario: Невалидный или чувствительный pattern
- **WHEN** config содержит malformed regexp либо невалидное обязательное поле
- **THEN** constructor возвращает error, проверяемую через `ErrInvalidConfig`, без pattern, entity value или иных caller-controlled config values в публичном тексте

### Requirement: Regexp matches становятся валидными findings с точными boundaries
CustomRegexpDetector MUST применять compiled pattern к валидному input и возвращать non-overlapping findings по полному regexp match в порядке текста с configured entity, confidence и пустой metadata для заполнения core. Match offsets MUST быть исходными UTF-8 byte offsets на rune boundaries. Zero-width full matches MUST безопасно пропускаться; adapter MUST NOT добавлять implicit word-boundary policy или возвращать capture group вместо полного match.

#### Scenario: Unicode text и full match span
- **WHEN** Unicode text содержит несколько `EMP-123456`-like значений и pattern находит их полным match
- **THEN** detector возвращает findings с точными byte spans каждого полного match в порядке текста

#### Scenario: Zero-width pattern
- **WHEN** валидный regexp производит zero-width matches
- **THEN** detector пропускает их и не возвращает finding с пустым span

#### Scenario: Boundary принадлежит caller pattern
- **WHEN** pattern совпадает с substring внутри более длинного alphanumeric token
- **THEN** adapter возвращает полный regexp match без неявного расширения или отсечения, а caller может задать требуемые boundaries в pattern

### Requirement: Custom detectors проходят общий reversible pipeline
CustomRegexpDetector и внешняя реализация публичного `Detector` MUST регистрироваться через `WithDetector` и проходить общие core validation, deterministic resolution, priorities, masking и restore. Custom string entity MUST иметь fallback priority ниже известных built-in entity, поэтому overlapping built-in finding побеждает custom finding; non-overlapping custom findings MUST маскироваться и восстанавливаться без специального caller code.

#### Scenario: Custom regexp round trip
- **WHEN** Guard с CustomRegexpDetector маскирует text с entity `EMPLOYEE_ID`
- **THEN** custom finding проходит Detect и Resolve, заменяется token и восстанавливается byte-for-byte

#### Scenario: Custom Go detector round trip
- **WHEN** Guard зарегистрирован с внешней concurrent-safe реализацией `Detector`, возвращающей custom string entity
- **THEN** её валидный finding проходит тот же Mask и Restore pipeline

#### Scenario: Overlap с built-in entity
- **WHEN** custom regexp и built-in detector возвращают overlapping findings
- **THEN** resolver детерминированно сохраняет built-in finding согласно priority policy

### Requirement: Ошибки и concurrent use не раскрывают входные данные
Один CustomRegexpDetector MUST быть безопасен для concurrent calls. Отмена context MUST завершать detection согласно core contract. Невалидный finding внешнего custom Go detector MUST отклоняться существующей safe `ErrInvalidFinding` error без input или matched substring в публичном сообщении; detector name и entity подчиняются существующему публичному contract, запрещающему sensitive detector metadata.

#### Scenario: Concurrent custom regexp detection
- **WHEN** один Guard одновременно обрабатывает несколько текстов одним CustomRegexpDetector
- **THEN** вызовы не имеют data race и не изменяют compiled regexp или результаты друг друга

#### Scenario: Unsafe custom finding
- **WHEN** внешний detector возвращает zero-width, out-of-range или non-rune-boundary finding
- **THEN** Guard возвращает error, проверяемую через `ErrInvalidFinding`, без исходного input или matched substring
