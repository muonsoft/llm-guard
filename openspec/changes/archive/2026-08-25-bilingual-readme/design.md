## Context

См. `proposal.md` — Why. Текущий `README.md` уже покрывает почти весь MVP, но его
структура отражает историю M3–M8 и дублирует подробные документы. Public API и
security boundary зафиксированы в `example_test.go`, `doc.go`,
`docs/known-limitations.md` и capability `oss-distribution`; релизный тег
`v0.1.0` ещё не опубликован.

## Goals / Non-Goals

**Goals:**

- Дать consumer понятный путь `позиционирование → install → quick start →
  supported scope → security boundary → deeper docs` на двух языках.
- Сохранить смысловую эквивалентность важных claims без требования буквального
  построчного перевода.
- Сделать public examples короткими и согласованными с тестируемым Go API.

**Non-Goals:**

- Перевод `docs/`, GoDoc, policies или OpenSpec artifacts на второй язык.
- Изменение runtime behavior, detector forms или release status.
- Добавление непроверенных coverage, benchmark или security claims.

## Decisions

### 1. Английский README основной, русский — peer translation

`README.md` остаётся GitHub entry point и пишется на английском для Go/OSS
аудитории; `README.ru.md` содержит тот же информационный контракт на русском.
Альтернатива с русским основным README сохраняла бы текущую аудиторию, но хуже
работала бы как международная package landing page.

### 2. Паритет определяется обязательными claims, а не строками

Обе версии сохраняют одинаковые status, install, detector-family, default policy,
state ownership и limitation facts. Формулировки и длина могут отличаться ради
естественного языка. Буквальная синхронизация отвергнута как хрупкая и ухудшающая
читабельность.

### 3. README — маршрут, подробные matrices остаются в docs

README показывает один quick start, компактную coverage table и несколько
production-critical notes. Формы PERSON/ADDRESS, evaluation methodology,
benchmarks, release evidence и orchestration остаются в существующих документах.
Альтернатива сохранить полный reference на landing page отвергнута из-за слабой
сканируемости и дублирования source of truth.

### 4. Никаких новых зависимостей или provider adapters

Пример использует framework-neutral `Mask`/`Restore` и условный caller-owned LLM
boundary. Core API не получает OpenAI или proxy dependencies; provider-specific
интеграция остаётся за пределами MVP.

## Risks / Trade-offs

- [Языковые версии расходятся со временем] → закрепить общий обязательный набор
  claims в OpenSpec и проверять оба файла в documentation review.
- [Сокращение скрывает detector edge cases] → показывать явный prefilter warning и
  вести consumer в `docs/known-limitations.md` и family quality reports.
- [До релиза install-команда создаёт ожидание опубликованного semver] → явно
  отделить repository readiness от ещё не существующего `v0.1.0` tag.
- [Пример README перестаёт компилироваться] → основывать его на `example_test.go`
  и проверять external consumer fixture.

## Migration Plan

Заменить текущий `README.md`, добавить `README.ru.md`, проверить ссылки и public
flow, затем опубликовать оба файла одним PR. Rollback — revert documentation
commit; runtime migration не требуется.
