## Purpose

Capability обнаруживает согласованные русские ФИО с консервативным false-positive profile, возвращает точные UTF-8 byte spans и безопасно участвует в локальном Mask/Restore pipeline без внешнего NLP runtime.

## ADDED Requirements

### Requirement: Библиотека предоставляет immutable RU PERSON detector
Библиотека SHALL предоставлять встроенный `NewPersonDetector`, совместимый с общим `Detector` interface, возвращающий `EntityPerson` и стабильное непустое имя detector. Detector MUST не иметь изменяемого process-global state, MUST быть безопасен для concurrent вызовов и MUST соблюдать caller context.

#### Scenario: Регистрация detector
- **WHEN** caller регистрирует `NewPersonDetector()` через `WithDetector` и создаёт `Guard`
- **THEN** Guard принимает detector без дополнительной конфигурации, Python runtime или внешнего сервиса

#### Scenario: Отменённый context
- **WHEN** PERSON detection начинается с уже отменённым context либо cancellation наблюдается во время обхода текста
- **THEN** detector возвращает context error без partial findings

#### Scenario: Concurrent detection
- **WHEN** несколько goroutines одновременно используют один PERSON detector или один immutable Guard
- **THEN** результаты детерминированы и выполнение не создаёт data race

### Requirement: PERSON поддерживает согласованные полные формы и инициалы
Detector MUST распознавать как единый PERSON span формы `имя фамилия`, `фамилия имя`, `имя отчество фамилия`, `фамилия имя отчество`, `фамилия И. О.` и `И. О. фамилия`. Он MUST поддерживать project-authored склонённые варианты обязательного corpus, включая дательный и творительный падежи, но MUST NOT возвращать нормализованный person fact или отдельные поля имени.

#### Scenario: Обязательные прямые формы
- **WHEN** input содержит `Иван Петров`, `Петров Иван`, `Иван Сергеевич Петров` или `Петров Иван Сергеевич`
- **THEN** detector возвращает ровно полный PERSON substring для каждой формы

#### Scenario: Формы с инициалами
- **WHEN** input содержит `Петров И. С.` или `И. С. Петров`
- **THEN** detector включает фамилию, оба инициала и их точки в один finding

#### Scenario: Склонённые формы
- **WHEN** input содержит `Ивану Сергеевичу Петрову` или `Иваном Сергеевичем Петровым`
- **THEN** detector возвращает полный исходный substring без лемматизации или изменения словоформы

### Requirement: PERSON spans точны для исходного UTF-8 текста
Каждый finding MUST использовать полуинтервал UTF-8 bytes `[Start, End)` исходной строки, MUST начинаться и заканчиваться на rune boundary и MUST включать только принятую name composition. Окружающая пунктуация, labels и пробелы MUST оставаться вне span; несколько непересекающихся PERSON occurrences MUST возвращаться в стабильном текстовом порядке.

#### Scenario: ФИО внутри Unicode-контекста
- **WHEN** согласованное ФИО окружено кириллическим текстом и пунктуацией
- **THEN** `text[Start:End]` byte-for-byte равно только исходному ФИО, а границы не зависят от числа Unicode code points

#### Scenario: Несколько персон
- **WHEN** input содержит два поддерживаемых ФИО, разделённых обычным текстом
- **THEN** detector возвращает два непересекающихся findings в порядке их `Start`

### Requirement: Acceptance policy остаётся консервативной
Detector MUST требовать поддерживаемую композицию как минимум из имени и фамилии либо фамилии и двух инициалов. Одиночные имена, фамилии и инициалы MUST NOT образовывать PERSON; lowercase/common-word, product/project-name и street-like contexts из versioned negative corpus MUST NOT давать findings. Capitalization, punctuation и outer token boundaries MUST запрещать совпадение внутри более длинного буквенно-цифрового token.

#### Scenario: Одиночное имя или фамилия
- **WHEN** input равен `Иван`, `Петров` либо содержит `директор Иван`
- **THEN** detector не возвращает PERSON finding

#### Scenario: Street-like context
- **WHEN** input содержит `улица Гагарина`
- **THEN** detector не принимает фамилию или всю фразу как PERSON

#### Scenario: Неименные capitalized sequences
- **WHEN** input содержит согласованный negative corpus названий продуктов, проектов и нарицательных сочетаний
- **THEN** detector не возвращает PERSON findings

#### Scenario: Вложение в token
- **WHEN** поддерживаемая последовательность примыкает к букве или цифре без token boundary
- **THEN** detector не возвращает усечённый PERSON span

### Requirement: NLP runtime ограничен воспроизводимым pure-Go scope
Production PERSON path MUST использовать нормативную tokenization с сохранением исходных byte offsets и только project-authored immutable role/form rules, разрешённые R0. Go module graph MUST NOT включать Natasha, Yargy, Pymorphy/OpenCorpora data, Python, ML runtime, runtime downloads либо неаудированные name dictionaries; внешний production source допускается только с зафиксированной revision и license inventory.

#### Scenario: Обычная Go-сборка
- **WHEN** consumer собирает или тестирует библиотеку без Python и network
- **THEN** PERSON detector компилируется и mandatory corpus выполняется локально

#### Scenario: Проверка зависимостей и данных
- **WHEN** M4 verification проверяет module graph и license inventory
- **THEN** в production path присутствуют только одобренный tokenizer и project-authored bounded rules, а reference tooling остаётся development-only

### Requirement: Versioned corpus задаёт измеримую quality boundary
Проект MUST проверять PERSON detector на synthetic versioned positive и negative cases, вычислять exact-span TP/FP/FN, precision, recall и exact-span rate отдельно для PERSON и сопоставлять результат с pinned Natasha baseline. Каждое reference-only отличие MUST иметь classification `intentional_difference`, `regression` либо `unsupported_out_of_scope`; обязательный corpus MUST иметь нулевые FP/FN, а утверждённые одиночные-name differences MUST оставаться intentional.

#### Scenario: Offline product evaluation
- **WHEN** разработчик запускает документированную PERSON corpus command без live Natasha environment
- **THEN** evaluator воспроизводимо проверяет product expectations, exact byte spans и публикует детерминированный quality summary

#### Scenario: Pinned differential comparison
- **WHEN** evaluator сравнивает product findings с `testdata/natasha/expected-python.jsonl`
- **THEN** более широкие Natasha matches для одиночных имён/фамилий учитываются как обоснованные intentional differences, а необъяснённое расхождение проваливает проверку

### Requirement: PERSON проходит общий Mask/Restore pipeline
`Guard.Mask` MUST разрешать PERSON findings общей priority policy и заменять полный PERSON span одним opaque token; `Guard.Restore` MUST восстанавливать исходную словоформу byte-for-byte. Документация MUST явно сообщать, что restore не выполняет morphology-aware согласование с текстом, изменённым LLM.

#### Scenario: PERSON round trip
- **WHEN** Guard с PERSON detector маскирует и без изменений восстанавливает текст с поддерживаемым ФИО
- **THEN** результат byte-for-byte равен исходному тексту, а masking использует один token на полный PERSON span

#### Scenario: Изменённый грамматический контекст
- **WHEN** LLM переносит PERSON token в контекст, требующий другой словоформы
- **THEN** Restore подставляет исходный substring буквально и не обещает грамматическое согласование
