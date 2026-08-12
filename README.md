# llm-guard

Lightweight open-source **LLM Guard for Go** — локальное обнаружение PII и секретов, обратимая маскировка и восстановление текста в LLM-пайплайнах.

**Статус:** ранняя стадия (спецификация и инфраструктура разработки; публичного API пока нет).

## Зачем

Встроить в Go-приложение как библиотеку, чтобы:

- маскировать чувствительные данные **до** отправки в LLM;
- восстанавливать placeholders **после** ответа модели;
- не зависеть от внешних API и отдельного gateway (в MVP).

```text
App → Guard (mask) → LLM → Guard (restore) → App
```

## Принципы MVP

- **PII-first**, **RU-first** (английский — вторичный)
- **library-first**, pure Go, CPU-only
- расширяемые detectors; proxy/adapters — позже

## Документация

| Путь | Описание |
|------|----------|
| [docs/light_llm_guard_go_mvp_plan.md](docs/light_llm_guard_go_mvp_plan.md) | Черновик MVP-плана |
| [openspec/](openspec/) | Spec-driven workflow (OpenSpec) |
| [AGENTS.md](AGENTS.md) | Инструкции для coding agents |

## Модуль

```bash
go get github.com/muonsoft/llm-guard
```

```go
module github.com/muonsoft/llm-guard

go 1.26
```

## Разработка

Планирование изменений — через [OpenSpec](https://github.com/Fission-AI/OpenSpec) (`openspec` CLI 1.8+). В Cursor: `/opsx-propose`.

```bash
go test ./...
agentmem skills verify
openspec doctor
```

## Лицензия

TBD
