# Light LLM Guard for Go — детализированный план MVP

> Документ предназначен для передачи кодинговому агенту как исходная спецификация проекта.
>
> Статус: draft MVP plan  
> Приоритет: PII-first, RU-first, pure Go, embeddable library  
> Дата: 2026-08-12

---

## 1. Цель проекта

Создать open-source **Light LLM Guard для Go**, который можно:

1. встроить напрямую в Go-приложение как библиотеку;
2. использовать для предобработки данных перед отправкой в LLM;
3. использовать для постобработки ответа LLM;
4. в поздней части MVP обернуть в OpenAI-compatible adapter/proxy.

Основной сценарий MVP:

```text
Application
    |
    | original text/messages
    v
Light LLM Guard
    |
    | detect PII/secrets
    | replace sensitive fragments with stable placeholders
    v
Masked text
    |
    v
LLM
    |
    v
LLM response with placeholders
    |
    v
Light LLM Guard
    |
    | restore placeholders
    v
Application
```

Ключевая функция MVP — **обратимая локальная маскировка PII** без обращения к внешним API и без обязательного отдельного gateway.

---

## 2. Позиционирование

Рабочее позиционирование:

> Lightweight open-source LLM Guard for Go applications.  
> Local PII detection, reversible masking and secret detection with RU-first support and no mandatory external service.

Ключевые свойства:

- Go-first;
- library-first;
- pure Go в MVP;
- русский язык — основной;
- английский — вторичный;
- локальная обработка;
- CPU-only;
- низкая инфраструктурная стоимость;
- расширяемые detectors;
- возможность позднее добавить proxy/adapters;
- архитектура не должна быть ограничена только PII, но MVP остаётся PII-first.

---

## 3. Scope MVP

### 3.1. Must have

#### PII detection

Structured PII:

- `EMAIL`
- `PHONE`
- `IP_ADDRESS`
- `URL`
- `INN`
- `SNILS`
- `PASSPORT`
- `BANK_CARD`
- `BANK_ACCOUNT`
- `DATE_OF_BIRTH` — только при достаточном контексте; произвольная дата не считается PII автоматически

Rule-based / linguistic PII:

- `PERSON`
- `ADDRESS`

Custom:

- custom regexp entities;
- публичный Go-интерфейс `Detector`.

#### Masking

- замена найденных сущностей на стабильные placeholders;
- уникальная нумерация в рамках `TokenSet`;
- сохранение mapping в RAM;
- обратимое восстановление;
- детерминированная обработка пересечений findings.

Пример:

```text
Иван Петров, email: ivan@example.com

↓

{{PII_PERSON_0001}}, email: {{PII_EMAIL_0001}}
```

#### Secrets detection

Базовый набор:

- JWT;
- PEM private keys;
- common API key patterns;
- connection strings;
- несколько наиболее распространённых provider-specific token formats.

Secrets — вторичный приоритет относительно PII.

#### Audit/logging

Должны поддерживаться как минимум два профиля:

- `production/hard` — не хранить исходные PII и исходный текст;
- `development/soft` — разрешить расширенный диагностический вывод, включая исходные значения, только при явном включении.

Безопасный профиль должен быть default.

#### Metrics

Минимальная наблюдаемость должна быть предусмотрена архитектурно.

Не делать Prometheus обязательной зависимостью `guard-core`. Предпочтительно предоставить observer/metrics interface и отдельную интеграцию.

---

### 3.2. Late MVP

- OpenAI Chat Completions adapter;
- OpenAI Responses API adapter;
- standalone proxy/gateway при необходимости;
- RAM-based mapping/session layer для proxy.

Эти компоненты не должны влиять на дизайн core API.

---

### 3.3. Post-MVP

Не реализовывать на первом этапе:

- prompt injection detection;
- jailbreak detection;
- content moderation;
- ML NER;
- ONNX Runtime;
- external classification APIs;
- policy DSL;
- sophisticated policy engine;
- RAG inspection;
- MCP inspection;
- tool-call security;
- persistent token store;
- Redis/DB mapping;
- multilingual NER;
- `ORGANIZATION`;
- aggressive PERSON detection;
- semantic secret classifiers;
- LLM-as-a-judge.

---

## 4. Нефункциональные требования

### Runtime

- pure Go;
- CPU-only;
- без Python runtime;
- без внешних API;
- library должна нормально работать внутри обычного Go service.

### Performance target

Ожидаемая нагрузка:

- `< 100 RPS`;
- допустимая latency — до нескольких сотен миллисекунд;
- никаких преждевременных оптимизаций в ущерб качеству детекции.

Цель первой версии — корректность и предсказуемость.

### Thread safety

После инициализации `Guard` и built-in detectors должны быть безопасны для конкурентного использования.

Не использовать глобальное mutable state.

### State

Core библиотека не должна иметь глобального token storage.

`Mask()` возвращает объект `TokenSet`, которым владеет вызывающий код.

---

## 5. Архитектурные принципы

### 5.1. Core не должен знать о конкретной LLM

Запрещены зависимости вида:

```go
MaskOpenAIRequest(...)
MaskAnthropicMessage(...)
```

в основном core package.

Core работает с текстом.

LLM-specific structures реализуются только в adapters.

---

### 5.2. Detector — базовая единица расширения

Предлагаемый контракт:

```go
type Detector interface {
    Detect(ctx context.Context, text string) ([]Finding, error)
}
```

Допустимо расширить интерфейс metadata-методом:

```go
type Detector interface {
    Name() string
    Detect(ctx context.Context, text string) ([]Finding, error)
}
```

Не делать интерфейс сложнее без реальной необходимости.

---

### 5.3. Finding — единый формат результата

Базовая модель:

```go
type EntityType string

type Finding struct {
    Entity     EntityType
    Start      int
    End        int
    Confidence float64
    Detector   string
}
```

### Требование к offsets

`Start` и `End` в MVP хранить как **byte offsets UTF-8**, потому что:

- Go regexp возвращает byte indexes;
- срез `text[start:end]` работает по byte indexes;
- это упрощает masking;
- устраняет лишние преобразования.

Это обязательно задокументировать в публичном API.

При необходимости rune offsets добавить позднее отдельными helper functions.

### Не хранить `Value` в публичном Finding без необходимости

Исходное значение можно получить через:

```go
text[f.Start:f.End]
```

Это уменьшает риск случайного логирования PII.

---

## 6. Entity model

Предлагаемый минимальный набор:

```go
const (
    EntityPerson      EntityType = "PERSON"
    EntityAddress     EntityType = "ADDRESS"
    EntityEmail       EntityType = "EMAIL"
    EntityPhone       EntityType = "PHONE"
    EntityIPAddress   EntityType = "IP_ADDRESS"
    EntityURL         EntityType = "URL"
    EntityINN         EntityType = "INN"
    EntitySNILS       EntityType = "SNILS"
    EntityPassport    EntityType = "PASSPORT"
    EntityBankCard    EntityType = "BANK_CARD"
    EntityBankAccount EntityType = "BANK_ACCOUNT"
    EntityDateOfBirth EntityType = "DATE_OF_BIRTH"

    EntitySecretJWT        EntityType = "SECRET_JWT"
    EntitySecretPrivateKey EntityType = "SECRET_PRIVATE_KEY"
    EntitySecretAPIKey     EntityType = "SECRET_API_KEY"
    EntityConnectionString EntityType = "CONNECTION_STRING"
)
```

Не фиксировать enum так, чтобы custom entities невозможно было добавить.

`EntityType` должен оставаться строковым типом.

---

## 7. Pipeline

Целевой pipeline MVP:

```text
Input text
    |
    v
Detector Pipeline
    |
    +--> Structured PII detectors
    +--> PERSON detector
    +--> ADDRESS detector
    +--> CustomRegex detectors
    +--> Secrets detectors
    |
    v
Raw Findings
    |
    v
Validation / Normalization
    |
    v
Finding Resolver
    |
    v
Resolved Findings
    |
    v
Minimal Policy
    |
    v
Mask / Tokenize
    |
    +--> Masked text
    +--> TokenSet
    +--> Audit events
```

Restore:

```text
LLM output
    |
    v
Placeholder scanner
    |
    v
TokenSet
    |
    v
Restored text
```

---

## 8. Structured PII detectors

Structured detectors не должны ограничиваться regexp там, где возможно выполнить валидацию.

Общий паттерн:

```text
regexp candidate
    ↓
normalize
    ↓
validate
    ↓
Finding
```

### EMAIL

- RFC-perfect parser не требуется;
- избегать чрезмерно permissive regexp;
- поддержать распространённые реальные email.

### PHONE

Основной приоритет:

- российские номера;
- формы `+7`, `8`, разделители, пробелы, скобки;
- EN/international common formats можно поддержать на базовом уровне.

Важно избегать маскировки любых длинных числовых последовательностей как телефона.

### IP_ADDRESS

- IPv4 — must;
- IPv6 — желательно в MVP;
- использовать `net.ParseIP` / стандартные средства там, где возможно.

### URL

- аккуратно обрабатывать URL с query params;
- отдельно проверить возможные overlaps с email.

### INN

- физлица;
- юрлица;
- обязательная checksum validation.

### SNILS

- normalization;
- checksum validation;
- поддержка форматированного и неформатированного вида.

### PASSPORT

Приоритет — паспорт РФ.

Не считать любые серии цифр паспортом без формата/контекста.

### BANK_CARD

- candidate regexp;
- normalize separators;
- Luhn check;
- не логировать полное значение.

### BANK_ACCOUNT

Определить минимально поддерживаемые банковские форматы РФ.

При невозможности надёжной детерминированной валидации использовать conservative matching.

### DATE_OF_BIRTH

Не создавать detector вида:

```text
12.10.1990 -> DATE_OF_BIRTH
```

без контекста.

Примеры допустимого контекста:

```text
дата рождения: 12.10.1990
родился 12 октября 1990 года
д.р. 12.10.1990
```

Обычные даты:

```text
встреча 12.10.2026
договор от 12.10.2026
```

не должны считаться PII.

---

## 9. PERSON detector

### Подход MVP

Rule-based Natasha-like extractor.

ML/NER не использовать.

### Консервативная политика MVP

Высокий приоритет:

```text
Иван Петров
Петров Иван
Иван Сергеевич Петров
Петров Иван Сергеевич
Петров И. С.
И. С. Петров
Ивану Сергеевичу Петрову
Иваном Сергеевичем Петровым
```

Не требуется агрессивно распознавать:

```text
Иван
Ваня
Петров
Петрову
директор Иван
```

Одиночные имена/фамилии — post-MVP/high-recall mode.

### Требования

- корректная работа с русской морфологией;
- склонённые формы;
- инициалы;
- несколько распространённых порядков ФИО;
- корректные byte spans;
- отсутствие чрезмерного количества false positives.

### Не цель MVP

- полноценный generic NER;
- nickname resolution;
- coreference;
- cross-message identity resolution;
- определение организаций.

---

## 10. ADDRESS detector

### Подход

Natasha-like rule-based extractor.

### Политика MVP

Одиночное название города/региона не является PII:

```text
Москва
Стокгольм
Ленинградская область
```

не маскировать как ADDRESS.

Искать **композиционный адрес**.

Примеры:

```text
г. Москва, ул. Тверская, д. 18
Москва, Тверская улица, 18
ул. Ленина, дом 15, кв. 27
проспект Мира, 101
```

### Heuristic confidence

Пример базовой модели:

```text
LOCATION only              -> ignore
STREET only                -> usually ignore
STREET + HOUSE             -> ADDRESS
CITY + STREET              -> candidate
CITY + STREET + HOUSE      -> strong ADDRESS
STREET + HOUSE + APARTMENT -> strong ADDRESS
```

Это не обязательно реализовывать через numerical score; допускаются простые правила.

### Важно

Уличные названия часто совпадают с PERSON:

```text
ул. Академика Сахарова
улица Гагарина
```

Resolver должен уметь предпочесть полный ADDRESS внутреннему PERSON finding.

---

## 11. Custom entities

### 11.1. CustomRegexpDetector

Минимальный config/API:

```go
type RegexDetectorConfig struct {
    Name       string
    Entity     EntityType
    Pattern    string
    Confidence float64
}
```

Пример:

```text
EMP-123456
CONTRACT-2026-00127
INC-12345
```

### 11.2. Custom Go Detector

Пользователь библиотеки должен иметь возможность передать собственную реализацию `Detector`.

Не делать:

- dynamic `.so` plugins;
- WASM plugins;
- runtime plugin marketplace;
- сложный DSL.

---

## 12. Secrets detector

Secrets — отдельное семейство сущностей.

Минимальный набор:

- JWT;
- PEM private key;
- GitHub-like token formats;
- GitLab-like token formats;
- OpenAI-like API keys;
- AWS-like access key patterns;
- DSN / connection strings с credential components.

Не пытаться в MVP считать любую high-entropy строку secret.

Entropy detector можно добавить post-MVP, потому что он легко создаёт false positives.

---

## 13. Finding Resolver

Resolver обязателен с первой версии.

Задачи:

1. deduplication;
2. overlap resolution;
3. nested findings;
4. priority;
5. deterministic ordering.

### Примеры конфликтов

```text
ул. Академика Сахарова, 10
```

Возможные findings:

- `ADDRESS`: весь адрес;
- `PERSON`: `Сахарова`.

Ожидаем:

- оставить ADDRESS;
- удалить PERSON overlap.

Другой пример:

```text
ivan.petrov@example.com
```

Не допускать, чтобы внутренние detectors создавали PERSON/URL fragments поверх EMAIL.

### Базовые приоритеты

Начальный вариант:

```text
EMAIL          > generic submatches
URL            > generic submatches
ADDRESS        > PERSON nested inside address
BANK_CARD      > generic numeric patterns
SNILS/INN      > generic numeric patterns
PASSPORT       > generic numeric patterns
```

### Алгоритм

Требования:

- детерминированный результат;
- сложность приемлема для небольшого количества findings;
- priority table configurable внутри library;
- тесты на все виды overlap.

Не делать prematurely сложный interval tree: для типичного prompt количество findings мало.

---

## 14. Masking / tokenization

### Placeholder format

Начальный default:

```text
{{PII_PERSON_0001}}
{{PII_EMAIL_0001}}
{{PII_PHONE_0001}}
{{SECRET_API_KEY_0001}}
```

Формат должен быть:

- редким в обычном тексте;
- легко распознаваемым;
- детерминированным;
- устойчивым к передаче через LLM;
- без исходного PII;
- без необязательной семантической информации.

### TokenSet

Пример внутренней модели:

```go
type TokenSet struct {
    // unexported mappings
}

type TokenMapping struct {
    Token  string
    Entity EntityType
    Value  string
}
```

`Value` содержит sensitive data и не должен автоматически сериализовываться/логироваться.

Рассмотреть полностью private `TokenMapping`, предоставляя только безопасные методы.

### API

Целевой UX:

```go
result, err := guard.Mask(ctx, text)
if err != nil {
    return err
}

llmResponse := callLLM(result.Text)

restored, err := guard.Restore(ctx, llmResponse, result.Tokens)
```

Пример result:

```go
type MaskResult struct {
    Text     string
    Findings []Finding
    Tokens   *TokenSet
}
```

### Collision handling

Нужно обработать случай, когда исходный текст уже содержит:

```text
{{PII_PERSON_0001}}
```

Варианты:

- выбирать session-specific prefix;
- escape existing placeholders;
- генерировать random namespace.

Для MVP предпочтителен **random/opaque namespace внутри TokenSet**, если это не ухудшает устойчивость LLM.

Пример:

```text
{{PII_a7f3_PERSON_0001}}
```

Не использовать UUID для каждого token без необходимости — это увеличивает noise.

### Restore

Restore должен:

- заменять только placeholders, принадлежащие данному `TokenSet`;
- не выполнять unrestricted regex replacement;
- не восстанавливать неизвестные токены;
- корректно работать при повторении одного placeholder;
- документировать поведение при изменении placeholder моделью.

---

## 15. Важное ограничение: морфология после Restore

Пример:

```text
До masking:
Я отправил письмо Ивану Петрову.

Mask:
Я отправил письмо {{PII_PERSON_0001}}.

LLM:
Я поговорил с {{PII_PERSON_0001}}.

Restore:
Я поговорил с Ивану Петрову.
```

Это допустимое ограничение MVP.

Не реализовывать склонение/морфологическую генерацию на этапе restore.

Добавить это в `Known limitations`.

---

## 16. Minimal policy layer

Полноценный policy engine — post-MVP, но pipeline не должен быть жёстко привязан к `all findings = mask`.

Минимальная внутренняя action model:

```go
type Action string

const (
    ActionAllow Action = "allow"
    ActionMask  Action = "mask"
    ActionBlock Action = "block"
)
```

В MVP достаточно простого mapping:

```text
PII     -> MASK
SECRETS -> MASK или BLOCK в зависимости от config
```

Не делать YAML policy DSL на первом этапе.

---

## 17. Audit

### Safe-by-default

Default mode не должен включать:

- original text;
- PII values;
- TokenSet content.

### Production audit event

Пример:

```json
{
  "event": "guard.mask",
  "text_length": 187,
  "findings": {
    "PERSON": 1,
    "EMAIL": 2
  },
  "masked_count": 3,
  "duration_ms": 4
}
```

### Development mode

При явном opt-in допускается:

- raw findings;
- исходный fragment;
- original/masked text.

Обязательно пометить режим как unsafe for production.

### Observer interface

Рассмотреть:

```go
type Observer interface {
    OnDetection(DetectionEvent)
    OnMask(MaskEvent)
    OnRestore(RestoreEvent)
}
```

С `NoopObserver` по умолчанию.

Не заставлять core импортировать конкретный logging framework.

---

## 18. Metrics

Минимальные логические метрики:

```text
guard_requests_total
guard_findings_total{entity=...}
guard_masked_total{entity=...}
guard_blocked_total{entity=...}
guard_detection_duration_seconds
guard_mask_duration_seconds
guard_restore_total
guard_restore_miss_total
```

Реализацию Prometheus вынести в optional adapter/package.

---

## 19. Предлагаемая структура репозитория

Не воспринимать как жёсткое требование; агент может предложить компактнее.

```text
/
├── guard.go
├── options.go
├── entity.go
├── finding.go
├── result.go
│
├── detector/
│   ├── detector.go
│   ├── regexp.go
│   ├── pii/
│   │   ├── email.go
│   │   ├── phone.go
│   │   ├── ip.go
│   │   ├── url.go
│   │   ├── inn.go
│   │   ├── snils.go
│   │   ├── passport.go
│   │   ├── card.go
│   │   ├── account.go
│   │   ├── birthdate.go
│   │   ├── person.go
│   │   └── address.go
│   └── secret/
│       ├── jwt.go
│       ├── pem.go
│       ├── api_key.go
│       └── dsn.go
│
├── resolver/
│   └── resolver.go
│
├── masking/
│   ├── masker.go
│   ├── restore.go
│   └── tokens.go
│
├── audit/
│   ├── observer.go
│   └── events.go
│
├── nlp/
│   ├── ...
│   └── natasha-compatible subset
│
├── adapters/
│   └── openai/        # late MVP
│
├── testdata/
│   ├── corpus/
│   └── differential/
│
└── internal/
```

Избегать искусственного создания десятков маленьких packages. Public API должен быть компактным.

---

# 20. Отдельный workstream: minimal Natasha-compatible subset

## 20.1. Цель

Не портировать Natasha целиком.

Цель:

> Реализовать минимальный pure-Go NLP/rule-extraction subset, достаточный для качественного PERSON и ADDRESS detection в Light LLM Guard.

Основной критерий — **behavioral compatibility**, а не 1:1 Python API compatibility.

---

## 20.2. Что известно из reference implementation

Актуальная структура Natasha показывает:

- `NamesExtractor` загружает grammar `natasha/grammars/name.py`;
- `AddrExtractor` использует `natasha/grammars/addr.py`;
- оба идут через Yargy parser;
- Natasha wrapper использует Yargy `MorphAnalyzer` и `MorphTokenizer`;
- Yargy — rule-based extraction framework и зависит от morphological analysis;
- адресная grammar существенно крупнее name grammar.

Следствие:

**не начинать порт с копирования extractor classes.**  
Сначала определить минимальный transitive dependency graph grammar runtime.

Reference:

- https://github.com/natasha/natasha
- https://github.com/natasha/natasha/blob/master/natasha/extractors.py
- https://github.com/natasha/natasha/blob/master/natasha/grammars/name.py
- https://github.com/natasha/natasha/blob/master/natasha/grammars/addr.py
- https://github.com/natasha/yargy
- https://github.com/natasha/razdel

---

## 20.3. Phase N0 — dependency audit

Перед кодированием агент обязан провести dependency audit.

### Задача

Построить дерево:

```text
NamesExtractor
  -> grammar/name
  -> Yargy constructs
  -> tokenizer
  -> morphology
  -> dictionaries/relations/interpretation
  -> ...

AddrExtractor
  -> grammar/addr
  -> Yargy constructs
  -> tokenizer
  -> morphology
  -> dictionaries/relations/interpretation
  -> ...
```

### Результат

Создать документ:

```text
docs/natasha-port-scope.md
```

с таблицей:

| Dependency / feature | Used by Name | Used by Addr | Required MVP | Strategy |
|---|---:|---:|---:|---|
| tokenization | yes | yes | yes | reuse existing Go Razdel port |
| morphology | yes | yes | yes | implement minimal interface / dictionary |
| grammar sequence | yes | yes | yes | implement |
| alternatives | yes | yes | yes | implement |
| optional nodes | ? | ? | ? | audit |
| predicates | ? | ? | ? | audit |
| agreement relations | ? | ? | ? | audit |
| interpretation/facts | yes | yes | likely | simplify |
| Earley parser | ? | ? | decision | investigate |

Знаки `?` должны быть заменены результатами исследования исходников.

### Critical decision gate

После audit решить:

**Option A — port minimal generic Yargy subset**

Подходит, если grammars используют достаточно широкий набор Yargy primitives.

**Option B — rewrite PERSON/ADDRESS grammars directly in Go**

Подходит, если generic parser оказывается непропорционально сложным относительно двух нужных extractors.

Не принимать решение заранее.

---

## 20.4. Phase N1 — tokenization

В проекте уже предполагается наличие Go-порта Razdel.

Задачи:

- адаптировать token representation под detector runtime;
- сохранить byte offsets;
- не копировать tokenizer Natasha, если существующего Razdel достаточно;
- добавить tests на punctuation, initials, abbreviations, addresses.

Тесты:

```text
Иван И. Петров
ул. Ленина, д. 10
г.Москва
Санкт-Петербург
Петров И.С.
```

---

## 20.5. Phase N2 — minimal morphology

Главный исследовательский вопрос — сколько morphology реально нужно grammars.

Нужно поддержать только тот набор grammatical information, который используется PERSON/ADDRESS rules.

Возможные категории:

```text
Name
Surn
Patr
NOUN
ADJF
gent
datv
accs
ablt
loct
masc
femn
sing
plur
Abbr
Geox
```

Этот список **не считать окончательным**: автоматически/вручную извлечь фактические граммемы из reference grammars.

### Требования

- нормализованная форма там, где нужна grammar;
- grammatical tags;
- несколько parse variants при неоднозначности, если это необходимо для parity;
- cached lookup;
- отсутствие Python runtime.

### Отдельное исследование

Проверить варианты:

1. порт/переиспользование OpenCorpora-compatible словаря;
2. минимальный embedded morphology dictionary;
3. использование существующей pure-Go morphology library;
4. генерация компактного словаря из открытых данных на build-time.

Зафиксировать licensing implications.

---

## 20.6. Phase N3 — rule engine / Yargy subset

Не портировать весь Yargy автоматически.

Сначала собрать **точный список constructs**, которые реально импортируют/use `name.py` и `addr.py`.

Типовые группы для проверки:

- sequence/rule;
- alternatives;
- predicates;
- dictionary lookup;
- normalized/caseless equality;
- morphological predicates;
- optional/repeat constructs;
- agreement relations;
- fact interpretation;
- morph pipelines;
- span extraction.

### Parser strategy

Если нужен grammar parser, сравнить:

- Earley-like implementation;
- deterministic graph/NFA style matcher;
- custom backtracking matcher;
- compiled rule graph.

Выбрать самый простой механизм, обеспечивающий parity для двух grammars.

Не портировать сложный general-purpose parser ради API compatibility, если он не нужен продукту.

---

## 20.7. Phase N4 — NamesExtractor equivalent

Target API может быть Go-native:

```go
type PersonMatch struct {
    Start int
    End   int

    First      string
    Last       string
    Patronymic string
}
```

Для Guard фактические поля имени вторичны; ключевое:

```text
correct span + PERSON entity
```

### Golden cases

Обязательный минимум:

```text
Иван Петров
Петров Иван
Иван Сергеевич Петров
Петров Иван Сергеевич
Петров И. С.
И. С. Петров
Ивану Петрову
Иваном Петровым
Ивану Сергеевичу Петрову
```

Negative:

```text
иван-чай
день Петра
улица Гагарина
проект Иван
марка Петров
```

Negative corpus расширять по мере нахождения false positives.

---

## 20.8. Phase N5 — AddrExtractor equivalent

Address grammar потенциально наиболее объёмная часть NLP workstream.

Не требуется автоматически достигать полного coverage Natasha.

Приоритетные компоненты:

- населённый пункт;
- улица;
- проспект;
- переулок;
- шоссе;
- дом;
- корпус;
- строение;
- квартира;
- почтовый индекс — если легко поддержать.

Guard-level detector должен поверх extractor применять правило композиционности.

То есть даже если extractor нашёл:

```text
Москва
```

Guard может не возвращать ADDRESS finding.

### Acceptance examples

Positive:

```text
г. Москва, ул. Тверская, д. 18
Москва, Тверская улица, дом 18
ул. Ленина, 10
проспект Мира, д. 101
ул. Ленина, д. 15, кв. 27
```

Negative:

```text
Москва
Санкт-Петербург
Ленинградская область
Тверская
улица
```

---

## 20.9. Differential tests against Python Natasha

Создать reference harness вне production runtime.

Предлагаемая схема:

```text
testdata/natasha/
    cases.jsonl
    expected-python.jsonl
```

Python script:

```text
tools/natasha-reference/
```

используется только разработчиками/CI research jobs для получения reference output.

Production Go library не зависит от Python.

### Сравнивать

- match count;
- span start/end;
- extracted type;
- normalized fields, только если они нужны Guard.

### Не требовать 100% parity

Различия допустимы, если:

- Go version безопаснее для PII use case;
- меньше false positives;
- behavior документирован;
- Guard corpus показывает лучшие product metrics.

Natasha — reference baseline, не абсолютная спецификация.

---

# 21. Evaluation corpus

Evaluation corpus создавать параллельно с кодом.

## Структура

```text
testdata/corpus/
├── structured/
│   ├── email.jsonl
│   ├── phone.jsonl
│   ├── inn.jsonl
│   ├── snils.jsonl
│   └── ...
├── person/
├── address/
├── secrets/
├── mixed/
└── negatives/
```

Пример:

```json
{
  "id": "person-001",
  "text": "Документы отправить Ивану Сергеевичу Петрову.",
  "entities": [
    {
      "type": "PERSON",
      "text": "Ивану Сергеевичу Петрову"
    }
  ]
}
```

В test runner вычислять byte offsets из annotated text либо хранить их в generated expected data.

---

## 22. Метрики качества

Основные:

- Precision;
- Recall;
- F1;
- False Positive Rate;
- False Negative Rate.

Для security use case особенно отслеживать **Recall/FNR**.

Но не добиваться recall любой ценой в MVP для PERSON: aggressive single-name policy остаётся post-MVP.

Отдельно считать метрики по entity:

```text
PERSON
ADDRESS
PHONE
EMAIL
INN
...
```

Не использовать только global aggregate F1.

---

## 23. Testing strategy

### Unit tests

Для каждого detector:

- positive;
- negative;
- malformed;
- boundaries;
- Unicode;
- punctuation;
- overlapping entities.

### Property/fuzz tests

Особенно для:

- mask/restore;
- resolver;
- byte offsets;
- structured identifiers.

Основные invariants:

```text
Restore(Mask(text)) == text
```

если masked text не изменён внешней стороной.

Также:

```text
resolved findings do not overlap
```

и:

```text
all spans are valid UTF-8 byte ranges
```

### Concurrency

```bash
go test -race ./...
```

должен проходить.

### Fuzz

Использовать Go fuzzing минимум для:

- resolver;
- masker;
- restore;
- regexp detector boundaries.

---

## 24. Security requirements

### Запрещено по умолчанию

- логировать `TokenSet`;
- `fmt.Printf("%+v", TokenSet)` с раскрытием PII;
- сериализовать raw token mappings автоматически;
- включать original text в audit;
- возвращать secrets в error text.

### Желательно

Реализовать safe `String()`:

```go
func (t TokenSet) String() string {
    return "<redacted>"
}
```

или вообще не экспортировать внутреннее содержимое.

### Errors

Ошибки должны содержать:

```text
detector name
entity type
operation
```

но не sensitive substring.

---

# 25. Публичный API — целевой draft

Не считать окончательной сигнатурой.

```go
guard, err := llmguard.New(
    llmguard.WithDefaultPIIDetectors(),
    llmguard.WithPersonDetector(...),
    llmguard.WithAddressDetector(...),
    llmguard.WithDetector(customDetector),
)
```

Использование:

```go
result, err := guard.Mask(ctx, prompt)
if err != nil {
    return err
}

answer, err := client.Generate(ctx, result.Text)
if err != nil {
    return err
}

answer, err = guard.Restore(ctx, answer, result.Tokens)
```

Detection-only:

```go
findings, err := guard.Detect(ctx, text)
```

Это важный API: библиотека должна быть полезна без masking.

---

# 26. Этапы реализации

## Phase 0 — Architecture + repository baseline

Результат:

- module structure;
- public API draft;
- `EntityType`;
- `Finding`;
- `Detector`;
- `Guard`;
- ADR по byte offsets;
- ADR по stateless core;
- CI;
- lint;
- race tests;
- benchmark skeleton.

Acceptance:

```bash
go test ./...
go test -race ./...
```

---

## Phase 1 — Structured PII vertical slice

Реализовать:

- EMAIL;
- PHONE;
- IP;
- URL;
- INN;
- SNILS;
- BANK_CARD.

Параллельно:

- resolver;
- masker;
- TokenSet;
- restore.

В конце phase должен работать end-to-end:

```text
original -> detect -> resolve -> mask -> restore -> original
```

Acceptance:

- structured corpus;
- fuzz `Mask/Restore`;
- no overlap;
- no unsafe logging.

---

## Phase 2 — Remaining structured PII + custom detector

Добавить:

- PASSPORT;
- BANK_ACCOUNT;
- contextual DATE_OF_BIRTH;
- `CustomRegexpDetector`;
- custom `Detector` registration.

Acceptance:

- tests;
- validation;
- no breaking API changes from Phase 1 unless зафиксированы ADR.

---

## Phase 3 — Natasha subset research + PERSON

Сначала выполнить `Phase N0`.

После decision gate:

- minimal morphology/rule layer;
- NamesExtractor-compatible behavior;
- `PersonDetector`.

Acceptance:

- corpus PERSON;
- differential tests;
- conservative false-positive profile.

---

## Phase 4 — ADDRESS

Реализовать:

- minimal address grammar;
- compositional ADDRESS logic;
- resolver priority `ADDRESS > nested PERSON`.

Acceptance:

- positive/negative corpus;
- differential comparison;
- street names based on persons do not leak as extra PERSON findings.

---

## Phase 5 — Secrets + production concerns

Добавить:

- basic secret detectors;
- audit profiles;
- observer;
- metrics abstraction;
- safe defaults;
- benchmarks;
- README usage;
- security considerations.

Это первая версия, которую можно рассматривать как **OSS MVP release candidate**.

---

## Phase 6 — OSS stabilization

Перед релизом:

- public API review;
- package documentation;
- examples;
- benchmark report;
- compatibility tests на актуальных Go versions;
- LICENSE;
- THIRD_PARTY_NOTICES;
- SECURITY.md;
- CONTRIBUTING.md;
- release workflow;
- semantic versioning;
- changelog.

Особое внимание лицензиям NLP dictionaries/data.

---

## Phase 7 — Late MVP: OpenAI adapters

Только после стабилизации core.

Добавить:

- Chat Completions adapter;
- Responses API adapter;
- optional HTTP proxy.

Требования:

- не дублировать PII logic;
- adapter должен использовать public/core API;
- state первоначально RAM;
- отдельные tests request -> mask -> upstream -> restore.

Не смешивать provider models с core types.

---

# 27. Definition of Done для MVP

MVP считается готовым, когда выполнено всё ниже:

- [ ] библиотека pure Go;
- [ ] нет обязательного внешнего сервиса;
- [ ] core API не зависит от OpenAI;
- [ ] structured PII detectors работают и валидируются;
- [ ] PERSON работает rule-based для консервативного набора ФИО;
- [ ] ADDRESS работает только для композиционных адресов;
- [ ] `ORGANIZATION` отсутствует;
- [ ] custom regex entities доступны;
- [ ] пользователь может добавить собственный `Detector`;
- [ ] basic secret detection есть;
- [ ] findings проходят deterministic resolver;
- [ ] masking обратим;
- [ ] mapping хранится только в caller-owned RAM object;
- [ ] safe audit mode — default;
- [ ] исходные PII не попадают в standard logs;
- [ ] `go test -race ./...` проходит;
- [ ] есть fuzz tests mask/restore/resolver;
- [ ] есть evaluation corpus;
- [ ] metrics по PERSON/ADDRESS/structured сущностям считаются отдельно;
- [ ] Natasha subset задокументирован;
- [ ] Python Natasha используется только как development reference;
- [ ] лицензии dependencies/dictionaries проверены;
- [ ] README содержит рабочий embedded Go example.

---

# 28. Known limitations MVP

Документировать явно:

1. одиночные имена/фамилии могут не детектироваться;
2. восстановление PERSON не выполняет морфологическое согласование;
3. модель может изменить placeholder;
4. ADDRESS не пытается маскировать одиночные географические сущности;
5. ORGANIZATION не детектируется;
6. нет semantic/ML NER;
7. нет prompt-injection protection;
8. нет moderation;
9. нет persistent token storage;
10. нет conversation-aware state;
11. защита не гарантирует нулевой false-negative rate.

---

# 29. Что агент должен делать при неоднозначности

При выборе между:

- большим generic NLP framework;
- минимальным решением под Guard;

предпочитать минимальное решение, если оно:

- не ломает extensibility;
- покрывает PERSON/ADDRESS requirements;
- тестируется;
- не ухудшает security properties.

Не портировать API Natasha/Yargy только ради совместимости.

При каждом существенном архитектурном решении создать ADR:

```text
docs/adr/
```

Минимально ожидаются:

- byte offsets vs rune offsets;
- stateless core / caller-owned TokenSet;
- Natasha port strategy;
- morphology source;
- rule-engine strategy;
- audit safe defaults.

---

# 30. Первый конкретный набор задач для кодингового агента

Рекомендуемый первый execution slice:

### Task 1 — bootstrap

- создать Go module;
- определить public types;
- определить `Detector`;
- создать `Guard`;
- написать architecture README.

### Task 2 — first detector

- EMAIL detector;
- corpus tests;
- byte span tests.

### Task 3 — resolver

- deterministic overlap resolver;
- priority table;
- tests.

### Task 4 — reversible masking

- TokenSet;
- placeholder generator;
- Mask;
- Restore;
- fuzz invariant.

### Task 5 — structured detector pack

- PHONE;
- IP;
- URL;
- INN;
- SNILS;
- BANK_CARD.

После Task 5 остановиться и проверить public API.

### Task 6 — Natasha audit

Не писать parser сразу.

Сначала подготовить:

```text
docs/natasha-port-scope.md
```

с точным dependency graph и предложением A/B:

- minimal Yargy subset;
- direct Go grammar implementation.

После review продолжать PERSON implementation.

---

# 31. Основной архитектурный ориентир

Проект должен оставаться похожим на:

```text
                   Light LLM Guard
                         |
                     Guard Core
                         |
          +--------------+--------------+
          |              |              |
       PII Detector   Secrets       Custom
          |
          v
       Findings
          |
          v
       Resolver
          |
          v
    Minimal Policy
          |
          v
       Masking
          |
    +-----+------+
    |            |
 Safe Text    TokenSet
    |            |
    v            |
   LLM           |
    |            |
    +-----+------+
          |
       Restore
```

а не превращаться в:

```text
gateway + auth + DB + Redis + policy server + ML service + admin UI
```

до того, как будет доказано качество базового PII engine.

---

## Reference sources

Для реализации Natasha-compatible subset использовать исходный код как reference baseline:

- Natasha: https://github.com/natasha/natasha
- Natasha extractors: https://github.com/natasha/natasha/blob/master/natasha/extractors.py
- Natasha name grammar: https://github.com/natasha/natasha/blob/master/natasha/grammars/name.py
- Natasha address grammar: https://github.com/natasha/natasha/blob/master/natasha/grammars/addr.py
- Yargy: https://github.com/natasha/yargy
- Razdel: https://github.com/natasha/razdel

Важно: перед переносом словарей, grammar data и morphology datasets отдельно проверить их лицензии и необходимость attribution.

---

## Финальное указание агенту

Не начинать с реализации proxy и не начинать с полного порта Natasha.

Первый продуктовый milestone:

> **Embedded Go library, которая локально находит structured PII, PERSON и ADDRESS, детерминированно маскирует их, возвращает caller-owned TokenSet и восстанавливает данные после LLM.**

Главные критерии качества:

1. безопасность default-конфигурации;
2. recall по критичным PII;
3. контролируемый false-positive rate;
4. корректные spans и reversible masking;
5. компактный и стабильный public Go API;
6. отсутствие Python/ML/external API dependencies в MVP.
