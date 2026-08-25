## ADDED Requirements

### Requirement: Двуязычный публичный entry point
Repository SHALL предоставлять основной англоязычный `README.md` и эквивалентный
русскоязычный `README.ru.md`; каждый файл MUST содержать заметный переход на
другой язык, canonical import path, рабочий `Mask → LLM → Restore` quick start,
поддерживаемые detector families, secure defaults, precision-oriented prefilter
boundary, известные ограничения и ссылки на подробную документацию.

#### Scenario: Consumer выбирает язык
- **WHEN** англо- или русскоязычный consumer открывает любой публичный README
- **THEN** в начале документа доступна ссылка на эквивалентную версию на другом языке

#### Scenario: Consumer оценивает интеграцию по любому README
- **WHEN** consumer читает английскую или русскую версию перед подключением library
- **THEN** обе версии сообщают одинаковые install/status facts, public flow, secret-block default, caller-owned `TokenSet`, false-negative boundary и отсутствие high-recall DLP guarantee

#### Scenario: Quick start следует проверяемому public API
- **WHEN** maintainer сверяет README quick start с repository examples и external consumer fixture
- **THEN** пример использует только canonical public module path, обрабатывает ошибки `New`, `Mask` и `Restore` и не зависит от `internal/` packages
