## Purpose

Capability предоставляет встраиваемый detection-only core для Go-приложений: подключение пользовательских детекторов, безопасную проверку findings и детерминированное выполнение без зависимости от LLM-провайдеров.

## ADDED Requirements

### Requirement: Публичный detection-only API
Корневой Go package MUST экспортировать строковый расширяемый `EntityType`, built-in entity constants, `Finding`, интерфейс `Detector`, `Guard`, `New`, option для регистрации detector и `Detect`. `Detector` MUST сообщать стабильное непустое имя, принимать `context.Context` и исходный текст и быть безопасным для concurrent calls после регистрации. Созданный `Guard` MUST быть immutable и безопасным для одновременных вызовов `Detect`; глобальное изменяемое состояние MUST отсутствовать.

#### Scenario: Внешний custom detector
- **WHEN** внешний Go-код создаёт `Guard` с custom detector и вызывает `Detect`
- **THEN** код компилируется через публичный API и получает findings этого detector

#### Scenario: Guard без detectors
- **WHEN** вызывается `Detect` у `Guard`, созданного без detectors
- **THEN** возвращается пустой результат без ошибки

#### Scenario: Некорректная конфигурация
- **WHEN** в `New` передан nil detector, detector с пустым именем или повторяющимся именем
- **THEN** создание `Guard` отклоняется безопасной ошибкой конфигурации

### Requirement: Finding использует проверяемые UTF-8 byte offsets
`Finding` MUST содержать entity, полуинтервал `[Start, End)`, confidence и detector metadata, но MUST NOT хранить исходное найденное значение. `Start` и `End` MUST быть byte offsets строки UTF-8. Core MUST принимать только непустую entity, конечный confidence в диапазоне `[0, 1]`, границы `0 <= Start < End <= len(text)` и обе границы на UTF-8 rune boundary. Исходный текст MUST быть валидным UTF-8.

#### Scenario: Unicode finding
- **WHEN** detector возвращает finding, границы которого точно охватывают многобайтовую UTF-8 последовательность
- **THEN** finding принимается с исходными byte offsets без преобразования в rune offsets

#### Scenario: Граница внутри rune
- **WHEN** detector возвращает `Start` или `End` внутри многобайтовой UTF-8 последовательности
- **THEN** `Detect` возвращает ошибку invalid finding без исходного значения или фрагмента текста

#### Scenario: Некорректные поля finding
- **WHEN** detector возвращает пустую entity, выходящие за текст или пустые границы, NaN, infinity либо confidence вне `[0, 1]`
- **THEN** `Detect` отклоняет результат безопасной ошибкой invalid finding

#### Scenario: Некорректный UTF-8 input
- **WHEN** `Detect` получает строку с некорректной UTF-8 последовательностью
- **THEN** input отклоняется до запуска detectors безопасной ошибкой invalid text

### Requirement: Core контролирует detector metadata
`Guard` MUST зафиксировать имя каждого detector при создании. Для каждого принятого finding core MUST установить поле detector metadata в зафиксированное имя. Core MUST разрешать detector оставить metadata пустым, но MUST отклонять любое отличающееся имя.

#### Scenario: Metadata заполняется core
- **WHEN** detector возвращает finding с пустым detector metadata
- **THEN** результат содержит имя detector, зафиксированное в `Guard`

#### Scenario: Несогласованная metadata
- **WHEN** detector возвращает finding с metadata другого detector
- **THEN** `Detect` отклоняет finding безопасной ошибкой invalid finding

### Requirement: Выполнение и агрегация детерминированы
`Detect` MUST запускать настроенные detectors независимо и конкурентно. После успешного завершения всех detectors он MUST возвращать findings в стабильном порядке: по `Start` по возрастанию, затем `End` по возрастанию, entity лексикографически, detector name лексикографически, confidence по убыванию, затем по порядку регистрации detector и исходному порядку внутри его результата. Результат MUST быть одинаковым независимо от порядка завершения goroutines.

#### Scenario: Detectors завершаются в разном порядке
- **WHEN** несколько detectors возвращают findings с различными задержками
- **THEN** последовательность результата соответствует стабильному ключу сортировки, а не порядку завершения

#### Scenario: Concurrent use одного Guard
- **WHEN** несколько goroutines одновременно вызывают `Detect` у одного `Guard`
- **THEN** все вызовы завершаются без data race и без взаимного изменения конфигурации или результатов

### Requirement: Ошибки не маскируют частичный failure
Если любой detector возвращает ошибку, `Detect` MUST отменить производный context для остальных detectors, дождаться завершения запущенных вызовов и вернуть nil findings вместе с ошибкой, сохраняющей исходную причину и безопасно идентифицирующей detector. Core MUST NOT возвращать частичные findings. Если ошибок несколько, выбор сообщаемой detector error MUST следовать порядку регистрации, а не порядку завершения.

#### Scenario: Detector error
- **WHEN** один из detectors возвращает ошибку
- **THEN** `Detect` возвращает nil findings, ошибка допускает проверку исходной причины и содержит имя ошибившегося detector

#### Scenario: Несколько detector errors
- **WHEN** несколько detectors возвращают ошибки конкурентно
- **THEN** возвращается ошибка detector с наименьшим порядком регистрации среди завершившихся с ошибкой

#### Scenario: Sensitive data в detector error
- **WHEN** detector возвращает ошибку, текст которой содержит исходное чувствительное значение
- **THEN** публичный error от `Detect` не включает текст wrapped-причины, исходный input или найденный substring, но сохраняет причину для `errors.Is` и `errors.As`

### Requirement: Context cancellation соблюдается
`Detect` MUST проверять caller context до запуска detectors, передавать каждому detector производный context и сигнализировать отмену остальным при failure. После завершения вызовов приоритет MUST иметь caller context error, если caller context был отменён.

#### Scenario: Context отменён до вызова
- **WHEN** caller context отменён до `Detect`
- **THEN** detectors не запускаются, возвращаются nil findings и ошибка, проверяемая как `context.Canceled`

#### Scenario: Context отменён во время detection
- **WHEN** caller отменяет context во время выполнения detectors и detectors завершаются по context
- **THEN** `Detect` возвращает nil findings и ошибку, проверяемую как caller context error

### Requirement: Core остаётся независимым от LLM infrastructure
Core detection package MUST NOT импортировать provider adapters, proxy, logging или persistence dependencies. Repository baseline MUST проверять package тестами, race detector и `go vet`.

#### Scenario: Repository verification
- **WHEN** выполняются milestone verification commands
- **THEN** package tests, race tests и `go vet` проходят без обязательных внешних сервисов
