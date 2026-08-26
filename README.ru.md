# llm-guard

**English version:** [README.md](README.md)

Лёгкий open-source **LLM Guard для Go** — локальный **precision-oriented
prefilter** (точностно-ориентированный предфильтр): обнаружение PII и секретов в
документированных поддерживаемых формах, обратимая маскировка и восстановление
текста в LLM-пайплайнах. Снижает риск утечки в рамках поддерживаемого охвата, но
**не заменяет** high-recall DLP, generic NER или **предметный аудит безопасности**.

**Статус:** MVP **готов к `v0.1.0`** — полный встроенный набор детекторов,
неизменяемая политика allow/mask/block, детерминированное разрешение конфликтов,
обратимая маскировка и восстановление, **безопасная наблюдаемость, не привязанная
к фреймворку** (`Noop` по умолчанию) и **воспроизводимые release gates**.
Публикация выполняется через GitHub workflow **Release** после зелёного CI; см.
[docs/release-checklist.md](docs/release-checklist.md).

**Библиотека, а не сервис:** pure Go, только CPU, встраивается в приложение.
Обработка выполняется локально — без обязательного внешнего API и отдельного
gateway. Встроенные детекторы **RU-first** для PERSON и ADDRESS; остальные
семейства следуют документированным структурным правилам.

## Установка

Снимок разработки из `main`:

```bash
go get github.com/muonsoft/llm-guard@main
```

Выпущенная версия (когда существует семантический тег):

```bash
go get github.com/muonsoft/llm-guard@v0.1.0
```

Требуется **Go 1.26.6+** — см.
[docs/compatibility-versioning.md](docs/compatibility-versioning.md).

## Пайплайн

```text
App → Guard (mask) → LLM → Guard (restore) → App
```

`TokenSet` храните в памяти вашего процесса на своей стороне границы с LLM — не
передавайте модели и недоверенным сторонам.

## Быстрый старт

Пример повторяет **сценарий `Mask` / `Restore` из публичного API** в
[example_test.go](example_test.go): один `Guard`, явная обработка ошибок и
**функция обратного вызова к LLM**, которую предоставляет вызывающий код.

```go
package myapp

import (
	"context"

	"github.com/muonsoft/llm-guard"
)

func ProcessWithLLM(
	ctx context.Context,
	prompt string,
	callLLM func(string) (string, error),
) (string, error) {
	guard, err := llmguard.New(
		llmguard.WithDetector(llmguard.NewEmailDetector()),
		llmguard.WithDetector(llmguard.NewPersonDetector()),
	)
	if err != nil {
		return "", err
	}

	masked, err := guard.Mask(ctx, prompt)
	if err != nil {
		return "", err
	}

	llmResponse, err := callLLM(masked.Text)
	if err != nil {
		return "", err
	}

	return guard.Restore(ctx, llmResponse, masked.Tokens)
}
```

Если зарегистрированы детекторы секретов, `Mask` по умолчанию блокируется для
секретов. Для обратимой маскировки используйте
`WithSecretAction(llmguard.ActionMask)` или переопределение по типу сущности через
`WithEntityAction`. Описание шаблонов — в
[docs/secret-patterns.md](docs/secret-patterns.md).

## Встроенное покрытие и значения по умолчанию

| Семейство | Конструктор(ы) | Действие по умолчанию |
|-----------|----------------|----------------------|
| EMAIL | `NewEmailDetector` | mask |
| PHONE | `NewPhoneDetector` | mask |
| URL | `NewURLDetector` | mask |
| IP-адрес | `NewIPDetector` | mask |
| ИНН | `NewINNDetector` | mask |
| СНИЛС | `NewSNILSDetector` | mask |
| PASSPORT | `NewPassportDetector` | mask |
| BANK_CARD | `NewBankCardDetector` | mask |
| BANK_ACCOUNT | `NewBankAccountDetector` | mask |
| DATE_OF_BIRTH | `NewDateOfBirthDetector` | mask |
| PERSON (RU ФИО) | `NewPersonDetector` | mask |
| ADDRESS (RU, составной) | `NewAddressDetector` | mask |
| SECRET_JWT | `NewJWTDetector` | **block** |
| SECRET_PRIVATE_KEY | `NewPEMPrivateKeyDetector` | **block** |
| SECRET_API_KEY | `NewAPIKeyDetector` | **block** |
| CONNECTION_STRING | `NewDSNDetector` | **block** |
| Custom regexp | `NewCustomRegexpDetector` | mask |

PII маскируются по умолчанию. Секреты блокируют `Mask`, пока явно не настроена
маскировка. `Guard` неизменяем после создания и безопасен для параллельных вызовов
`Detect`, `Mask` и `Restore`.

## Безопасность и эксплуатация

- Наблюдатель по умолчанию — `NoopObserver`: без callbacks и побочных эффектов.
  **Безопасные наблюдатели** возвращают только **ограниченный набор метаданных** и
  не должны логировать исходный текст или содержимое `TokenSet`.
- Публичные ошибки не содержат исходный ввод и найденные чувствительные значения.
- `TokenSet` живёт в памяти вашего процесса; чувствительные значения не
  раскрываются через `String`, `GoString` или JSON.
- `Restore` подставляет исходный фрагмент **byte-for-byte** и **не** согласует
  словоформу с грамматическим контекстом, изменённым LLM.
- Если модель меняет placeholder-токены, восстановление может пропустить фрагмент
  или завершиться частично.
- **UNSAFE FOR PRODUCTION:** `WithUnsafeDevelopmentObserver` раскрывает
  **исходные данные и результаты обнаружения** только для локальной отладки.

## Известные ограничения

Полный список: [docs/known-limitations.md](docs/known-limitations.md).

- **Prefilter, не DLP** — снижает риск для документированных поддерживаемых
  форм; не гарантирует обнаружение всего PII (одиночные имена, адреса «только
  город», ИНН/СНИЛС без валидной контрольной суммы, неизвестные формы секретов).
- Консервативные **RU-first** детекторы; английский и пограничные случаи могут
  пропускаться или поддерживаться частично.
- Нет гарантии нулевых false negatives, защиты от prompt injection, модерации
  контента или постоянного хранения токенов.
- **Тестовые корпуса** — репрезентативные дымовые проверки, не production SLO;
  **замеры производительности** зависят от железа.

## Качество и разработка

```bash
go test ./...
go vet ./...
./scripts/release-check.sh
GOTOOLCHAIN=go1.26.6 ./scripts/release-check.sh vuln
./scripts/release-check.sh consumer
```

`./scripts/release-check.sh` — офлайн полный пробный прогон релиза (без сети).
Проверка уязвимостей — отдельный онлайн-gate и требует точной минимальной
toolchain (`GOTOOLCHAIN=go1.26.6`); см.
[docs/release-checklist.md](docs/release-checklist.md).

Границы качества по семействам:
[docs/person-quality-report.md](docs/person-quality-report.md),
[docs/address-quality-report.md](docs/address-quality-report.md). Методология
оценки и замеров производительности:
[docs/evaluation-baseline.md](docs/evaluation-baseline.md),
[docs/benchmark-baseline.md](docs/benchmark-baseline.md).

## Дополнительная документация

| Документ | Назначение |
|----------|------------|
| [docs/known-limitations.md](docs/known-limitations.md) | Опубликованные границы MVP |
| [docs/compatibility-versioning.md](docs/compatibility-versioning.md) | Pre-1.0 SemVer и поддержка Go |
| [docs/secret-patterns.md](docs/secret-patterns.md) | Описание шаблонов секретов |
| [docs/release-checklist.md](docs/release-checklist.md) | Чеклист релиза и GitHub Release workflow |
| [CHANGELOG.md](CHANGELOG.md) | Release notes и планируемый scope `v0.1.0` |

## Участие, безопасность, лицензия

- [CONTRIBUTING.md](CONTRIBUTING.md) — правила участия и требования к
  происхождению изменений
- [SECURITY.md](SECURITY.md) — сообщения об уязвимостях (приватно, только
  **синтетические тестовые данные**)
- [MIT](LICENSE) — Copyright (c) 2026 MuonSoft
