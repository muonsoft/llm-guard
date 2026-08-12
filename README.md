# llm-guard

Lightweight open-source **LLM Guard for Go** — локальное обнаружение PII и секретов, обратимая маскировка и восстановление текста в LLM-пайплайнах.

**Статус:** ранний MVP; доступны built-in detectors для EMAIL и structured pack (PHONE, IP_ADDRESS, URL, INN, SNILS, BANK_CARD, PASSPORT, BANK_ACCOUNT, contextual DATE_OF_BIRTH), `CustomRegexpDetector` для string entity, deterministic resolver, reversible masking/restore и detection-only API для custom Go detectors.

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
| [docs/milestones/](docs/milestones/) | Milestone scope, status dashboard и orchestration runbook |
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

## Использование

```go
guard, err := llmguard.New(
    llmguard.WithDetector(llmguard.NewPhoneDetector()),
    llmguard.WithDetector(llmguard.NewIPDetector()),
    llmguard.WithDetector(llmguard.NewURLDetector()),
    llmguard.WithDetector(llmguard.NewINNDetector()),
    llmguard.WithDetector(llmguard.NewSNILSDetector()),
    llmguard.WithDetector(llmguard.NewBankCardDetector()),
    llmguard.WithDetector(llmguard.NewEmailDetector()),
)
if err != nil {
    return err
}

result, err := guard.Mask(ctx, prompt)
if err != nil {
    return err
}

llmResponse := callLLM(result.Text)

restored, err := guard.Restore(ctx, llmResponse, result.Tokens)
```

`TokenSet` принадлежит caller и не раскрывает чувствительные значения через `String`, `GoString` или JSON.

### Structured forms (M3)

| Entity | Supported forms | Validation limits |
|--------|-----------------|-------------------|
| PASSPORT | `NNNN NNNNNN`, `NN NN NNNNNN` after RU marker (`паспорт`, `паспорт РФ`, `серия`, `паспортные данные`) | No checksum; separated `серия …, номер …` unsupported |
| BANK_ACCOUNT | 20 ASCII digits or five `####` groups after RU marker (`р/с`, `расчётный счёт`, …) | Context-first without BIK; same-line BIK triggers checksum when present |
| DATE_OF_BIRTH | `DD.MM.YYYY`, `DD/MM/YYYY`, or `D <month> YYYY` after birth marker (`дата рождения`, `д.р.`, `родился`, `родилась`) | Calendar validity only; ordinary/contract dates ignored |

Custom regexp boundaries are caller-owned; the adapter returns full matches without implicit word boundaries.

```go
detector, err := llmguard.NewCustomRegexpDetector(llmguard.RegexDetectorConfig{
    Name:       "employee_id",
    Entity:     llmguard.EntityType("EMPLOYEE_ID"),
    Pattern:    `EMP-[0-9]{6}`,
    Confidence: 0.9,
})
```

## Разработка

Планирование изменений — через [OpenSpec](https://github.com/Fission-AI/OpenSpec) (`openspec` CLI 1.8+). В Cursor: `/opsx-propose`.

```bash
go test ./...
agentmem skills verify
openspec doctor
```

## Лицензия

[MIT](LICENSE) — Copyright (c) 2026 MuonSoft
