# M8 — OSS stabilization и `v0.1.0` readiness

| Поле | Значение |
|---|---|
| Change | `m8-oss-stabilization` |
| Capabilities | `oss-distribution`, при необходимости уточнения existing specs |
| Зависимости | M7 |
| Default variant | C |
| Результат | Проверенный OSS package, готовый к отдельному release action |

## Outcome

Стабилизировать public surface, supply-chain документы и release automation.
Milestone подтверждает готовность `v0.1.0`, но не публикует tag/release и не делает
push без отдельного явного запроса.

## Scope

- Финальный public API/package documentation review и compatibility policy.
- External-package examples и compile tests.
- Go version/CI matrix, deterministic tests и bounded fuzz smoke jobs.
- LICENSE audit, `THIRD_PARTY_NOTICES` и provenance NLP dictionaries/data.
- `SECURITY.md`, `CONTRIBUTING.md`, changelog и release checklist.
- Reproducible benchmark/quality summary со ссылкой на M7 baseline.
- Release workflow dry-run без публикации artifacts.
- Финальная проверка Definition of Done и known limitations MVP.

## Out of scope

- Фактический push/tag/GitHub release/package announcement.
- Новая функциональность, новые entity types и performance rewrites.
- OpenAI adapters/proxy; они получают отдельный post-MVP OpenSpec change.

## Планируемые задачи OpenSpec

- [ ] Провести public API/doc review и external compile examples.
- [ ] Настроить/проверить Go CI, race и fuzz smoke matrix.
- [ ] Завершить dependency/data license и notices audit.
- [ ] Добавить SECURITY и CONTRIBUTING policies.
- [ ] Добавить changelog, versioning и release checklist.
- [ ] Выполнить release workflow dry-run без публикации.
- [ ] Сверить полный MVP Definition of Done и limitations.
- [ ] Выполнить финальную quality/benchmark regression comparison.

## Acceptance

1. Новый consumer может подключить module и выполнить documented embedded example.
2. Все code/data dependencies имеют совместимую license/provenance и необходимые
   notices; неизвестные источники блокируют readiness.
3. CI/release dry-run воспроизводит test, vet, race и предусмотренные quality gates.
4. Definition of Done закрыт evidence, а known limitations явно опубликованы.
5. Нет незапланированного provider/proxy/persistence/ML scope и release не
   публикуется автоматически.

## Verification

```bash
go test ./...
go vet ./...
go test -race ./...
go test ./... -run '^$' -bench . -benchmem
openspec validate --all --strict --no-interactive
```

Дополняется точными CI lint/fuzz/example/license commands, созданными в change.
Dry-run обязан исключать tag, push и publish side effects.

## Exit evidence

- Archived `oss-distribution` spec и все main specs проходят strict validation.
- Final DoD matrix, license inventory, quality/benchmark comparison и release
  checklist.
- Reviewed green checkpoint, после которого release выполняется отдельной
  пользовательской операцией.
