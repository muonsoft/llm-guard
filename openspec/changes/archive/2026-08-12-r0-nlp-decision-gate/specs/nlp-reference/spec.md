## Purpose

Capability задаёт воспроизводимый reference baseline и проверяемый архитектурный decision gate для минимального pure-Go NLP runtime, на котором последующие milestone реализуют консервативные RU PERSON и ADDRESS detectors.

## ADDED Requirements

### Requirement: Reference dependency audit является полным и воспроизводимым
Проект MUST зафиксировать точные revisions Natasha, совместимого Yargy, Razdel reference и morphology/data sources. Audit MUST проследить `NamesExtractor` и `AddrExtractor` до tokenizer, morphology, grammar constructs, predicates, parser, interpretation и dictionaries; required-feature matrix MUST не содержать неизвестных значений для решений, блокирующих M4 или M5.

#### Scenario: Повторная проверка upstream graph
- **WHEN** разработчик открывает audit после завершения R0
- **THEN** он может получить каждую reference revision и однозначно определить, какие upstream features требуются PERSON и ADDRESS scope

#### Scenario: Неизвестный обязательный dependency
- **WHEN** происхождение или назначение обязательного grammar, code либо data dependency не установлено
- **THEN** R0 остаётся незавершённым, а зависящий от него PERSON или ADDRESS milestone MUST быть заблокирован

### Requirement: Архитектурная граница NLP runtime зафиксирована
Decision gate MUST зафиксировать `github.com/muonsoft/go-razdel` как отдельную Go-зависимость токенизации с UTF-8 byte offsets и product-specific PERSON/ADDRESS runtime как внутреннюю часть `llm-guard`. MVP MUST NOT требовать отдельный `go-natasha`, публичный Natasha/Yargy-compatible API или полный generic Yargy parser; Natasha MUST использоваться как reference baseline, а не как нормативная product semantics.

#### Scenario: Проверка production architecture
- **WHEN** M4 или M5 начинает реализацию по результатам R0
- **THEN** task packet указывает внутренние product rules и `go-razdel`, не создавая отдельный Natasha-compatible Go module

#### Scenario: Возможное последующее выделение runtime
- **WHEN** после MVP появляется второй независимый consumer и стабильный Go-native grammar contract
- **THEN** выделение отдельного модуля рассматривается новым change и не считается обязательством R0

### Requirement: Parser strategy ограничена фактическим product scope
Decision gate MUST выбрать прямые Go-правила и bounded token matcher для конструкций, фактически необходимых M4/M5. Generic chart/Earley runtime допускается только если audit доказывает обязательную grammar recursion или другой construct, который нельзя покрыть bounded matcher без потери согласованного product behavior; такое отклонение MUST быть оформлено отдельным пересмотром design до product implementation.

#### Scenario: Upstream construct не нужен product scope
- **WHEN** Natasha grammar использует construct только для поведения, исключённого из M4/M5
- **THEN** runtime MUST не реализовывать construct только ради differential parity

#### Scenario: Audit обнаруживает неподдерживаемую обязательную конструкцию
- **WHEN** обязательный positive corpus невозможно выразить выбранным bounded matcher
- **THEN** R0 MUST остановить M4/M5 и пересмотреть parser decision с воспроизводимым failing fixture

### Requirement: Morphology и dictionaries имеют минимальный лицензированный контракт
R0 MUST перечислить фактически требуемые normalized forms и grammemes, определить минимальный immutable/concurrent-safe morphology contract и выбрать source/distribution strategy по измеримым критериям качества, footprint и thread safety. Каждый включаемый code или data source MUST иметь установленную лицензию, provenance и attribution; неизвестная либо несовместимая лицензия MUST запрещать включение source в production module.

#### Scenario: Добавление словаря в production plan
- **WHEN** source предлагается для PERSON или ADDRESS runtime
- **THEN** license inventory содержит его revision, происхождение, лицензию, способ дистрибуции и оценку footprint

#### Scenario: Лицензия словаря не установлена
- **WHEN** provenance или условия распространения словаря нельзя подтвердить
- **THEN** source исключается из production plan либо блокирует соответствующий milestone до замены

### Requirement: Reference harness отделён от production runtime
Проект MUST предоставить optional version-pinned reference harness, который воспроизводимо формирует или проверяет versioned JSONL fixtures для PERSON и ADDRESS. Harness MUST явно различать Python codepoint offsets и Go UTF-8 byte offsets, MUST сохранять только согласованные fixture inputs/outputs и MUST NOT входить в production import path, обязательные Go tests или transitive dependencies Go module.

#### Scenario: Reference comparison доступен разработчику
- **WHEN** разработчик запускает документированную reference-команду в pinned environment
- **THEN** harness воспроизводимо проверяет schema version, match type, span text и сопоставимые normalized fields

#### Scenario: Обычная установка Go library
- **WHEN** consumer импортирует `github.com/muonsoft/llm-guard` или выполняются обычные Go tests
- **THEN** Python, Natasha, Yargy, Pymorphy и автоматическая загрузка внешних data не требуются

### Requirement: Differential baseline допускает обоснованные product differences
R0 MUST определить per-entity precision-oriented metrics, positive/negative/ambiguous corpus classes и формат списка intentional differences. Отклонение от Natasha MUST считаться допустимым, если оно воспроизводимо, документировано и соответствует safety boundary M4/M5: одиночные имена и фамилии не принимаются как PERSON, а одиночные settlement/region/street parts не принимаются как ADDRESS без согласованной композиции.

#### Scenario: Natasha возвращает более широкий PERSON match
- **WHEN** reference принимает одиночное имя или фамилию, исключённые M4 negative corpus
- **THEN** Go baseline MUST пометить отличие как intentional и не считать его regression

#### Scenario: Natasha разбирает одиночную часть адреса
- **WHEN** reference возвращает settlement, region или street без достаточной композиции
- **THEN** Go baseline MUST применять product acceptance policy и не требовать такого ADDRESS finding

### Requirement: Exit evidence позволяет начать M4 и M5 без повторного базового исследования
R0 MUST завершиться документированными architecture decisions, закрытым dependency/license audit, воспроизводимыми sample fixtures и точными verification commands. Exit evidence MUST быть достаточен, чтобы M4 и M5 реализовывали только утверждённый bounded scope без нового решения о repository boundary, tokenizer или parser family.

#### Scenario: Передача M4 и M5 в реализацию
- **WHEN** R0 прошёл verification и архивирован
- **THEN** design, audit и fixtures однозначно задают разрешённые production dependencies, runtime boundary, morphology source strategy и intentional-difference policy
