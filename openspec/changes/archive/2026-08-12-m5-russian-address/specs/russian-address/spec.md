## Purpose

Capability обнаруживает только согласованные композиционные русские адреса, возвращает точные UTF-8 byte spans и безопасно участвует в локальном Mask/Restore pipeline без внешнего геокодинга или NLP runtime.

## ADDED Requirements

### Requirement: Библиотека предоставляет immutable RU ADDRESS detector
Библиотека SHALL предоставлять встроенный `NewAddressDetector`, совместимый с общим `Detector` interface, возвращающий `EntityAddress` и стабильное непустое имя detector. Detector MUST не иметь изменяемого process-global state, MUST быть безопасен для concurrent вызовов и MUST соблюдать caller context.

#### Scenario: Регистрация detector
- **WHEN** caller регистрирует `NewAddressDetector()` через `WithDetector` и создаёт `Guard`
- **THEN** Guard принимает detector без дополнительной конфигурации, Python runtime, сети или внешнего сервиса

#### Scenario: Отменённый context
- **WHEN** ADDRESS detection начинается с уже отменённым context либо cancellation наблюдается во время обхода текста
- **THEN** detector возвращает context error без partial findings

#### Scenario: Concurrent detection
- **WHEN** несколько goroutines одновременно используют один ADDRESS detector или один immutable Guard
- **THEN** результаты детерминированы и выполнение не создаёт data race

### Requirement: ADDRESS принимает поддерживаемые street и house композиции
Detector MUST принимать композицию, содержащую поддерживаемые street и house parts. Street part MUST поддерживать prefix и suffix forms для `улица`/`ул.`, `проспект`/`пр-т`, `переулок`/`пер.` и `шоссе`; house part MUST поддерживать явные `дом`/`д.` и bounded unlabeled number после street. Принятая композиция MAY также содержать settlement, `корпус`/`корп.`, `строение`/`стр.` и `квартира`/`кв.` parts в распространённом порядке. Порядок settlement относительно street и формы compact punctuation MUST определяться versioned corpus; postal index не входит в обязательный M5 scope.

#### Scenario: Обязательные полные формы
- **WHEN** input содержит `г. Москва, ул. Тверская, д. 18`, `Москва, Тверская улица, дом 18`, `ул. Ленина, 10` или `проспект Мира, д. 101`
- **THEN** detector возвращает ровно один ADDRESS finding на всю поддерживаемую композицию

#### Scenario: Extended building parts
- **WHEN** input содержит `ул. Ленина, д. 15, корп. 2, стр. 1, кв. 27`
- **THEN** accepted ADDRESS span включает house и все следующие согласованные building/apartment parts

#### Scenario: Compact punctuation
- **WHEN** input содержит `ул.Тверская,д.1`
- **THEN** detector принимает ту же street+house композицию без требования пробела после сокращения или запятой

### Requirement: Композиционная acceptance policy остаётся консервативной
Public ADDRESS finding MUST требовать одновременно street и house parts. Одиночные settlement, region, street либо house-like number, а также settlement+street без house MUST NOT образовывать ADDRESS. Detector MUST применять bounded labels, numbers и outer token boundaries, чтобы не принимать усечённые parts внутри более длинных буквенно-цифровых tokens.

#### Scenario: Одиночная географическая часть
- **WHEN** input равен `Москва`, `Санкт-Петербург`, `Ленинградская область`, `Тверская` или `улица Ленина`
- **THEN** detector не возвращает ADDRESS finding

#### Scenario: Неоднозначная композиция без дома
- **WHEN** input содержит `Москва, Тверская улица`
- **THEN** detector не возвращает ADDRESS finding

#### Scenario: House-like fragment вне адреса
- **WHEN** число или building label не следует за принятой street part либо часть встроена в более длинный token
- **THEN** detector не возвращает усечённый ADDRESS finding

### Requirement: ADDRESS spans точны для исходного UTF-8 текста
Каждый finding MUST использовать полуинтервал UTF-8 bytes `[Start, End)` исходной строки, MUST начинаться и заканчиваться на rune boundary и MUST включать только принятую address composition. Внутренняя согласованная пунктуация и labels MUST входить в span, а внешняя пунктуация и surrounding text MUST оставаться вне; несколько непересекающихся ADDRESS occurrences MUST возвращаться в стабильном текстовом порядке.

#### Scenario: Адрес внутри Unicode-контекста
- **WHEN** поддерживаемый адрес окружён кириллическим текстом, кавычками или punctuation
- **THEN** `text[Start:End]` byte-for-byte равно только исходной address composition

#### Scenario: Несколько адресов
- **WHEN** input содержит два поддерживаемых адреса, разделённых обычным текстом
- **THEN** detector возвращает два непересекающихся findings в порядке их `Start`

### Requirement: NLP runtime ограничен воспроизводимым pure-Go scope
Production ADDRESS path MUST использовать только принятый R0 tokenizer с исходными byte offsets и project-authored immutable annotations/composition rules. Go module graph MUST NOT включать Natasha, Yargy, Pymorphy/OpenCorpora data, Python, ML runtime, runtime downloads, внешние address databases либо geocoding calls.

#### Scenario: Обычная Go-сборка
- **WHEN** consumer собирает и тестирует библиотеку без Python, сети и внешних данных
- **THEN** ADDRESS detector и mandatory corpus выполняются локально

#### Scenario: Проверка production graph
- **WHEN** M5 verification проверяет dependencies и license inventory
- **THEN** reference tooling остаётся development-only, а production path содержит только разрешённый R0 runtime и project-authored rules

### Requirement: Versioned corpus задаёт измеримую ADDRESS quality boundary
Проект MUST проверять detector на synthetic versioned positive, negative и ambiguous cases, вычислять exact-span TP/FP/FN, precision, recall и exact-span rate отдельно для ADDRESS и сопоставлять результат с pinned Natasha baseline. Каждое reference-only либо product-only отличие MUST иметь classification `intentional_difference`, `regression` либо `unsupported_out_of_scope`; обязательный corpus MUST иметь нулевые FP/FN.

#### Scenario: Offline product evaluation
- **WHEN** разработчик запускает документированную ADDRESS corpus command без live Natasha environment
- **THEN** evaluator воспроизводимо проверяет product expectations, exact byte spans и выдаёт детерминированный quality summary

#### Scenario: Pinned differential comparison
- **WHEN** evaluator сравнивает product findings с versioned Natasha fixtures
- **THEN** более широкие matches одиночных address parts либо unsupported grammar учитываются только как документированные differences, а необъяснённое расхождение проваливает проверку

### Requirement: ADDRESS проходит общий Mask/Restore pipeline
`Guard.Mask` MUST разрешать ADDRESS findings общей priority policy и заменять полный accepted ADDRESS span одним opaque token; `Guard.Restore` MUST восстанавливать исходный адрес byte-for-byte. Detector MUST NOT сохранять исходный текст в process-global state.

#### Scenario: ADDRESS round trip
- **WHEN** Guard с ADDRESS detector маскирует и без изменений восстанавливает текст с поддерживаемым адресом
- **THEN** результат byte-for-byte равен исходному тексту, а masking использует один token на полный ADDRESS span

#### Scenario: Shared detector без shared tokens
- **WHEN** concurrent callers используют общий detector или Guard для независимых address texts
- **THEN** каждому вызову принадлежит отдельный TokenSet, и исходные значения не смешиваются между вызовами
