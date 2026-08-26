# F1 — `v0.1.0` release finalization

| Поле | Значение |
|---|---|
| Change | `v0-1-0-release-finalization` |
| Capabilities | `oss-distribution` |
| Зависимости | D1 |
| Variant | C |
| Результат | Security-clean, проверенный release commit без tag/push/publication |

## Outcome

Закрыть найденный pre-release security blocker, согласовать minimum Go toolchain,
CI, release gates и public documentation и получить reviewed green checkpoint,
после которого maintainer может отдельно одобрить и опубликовать `v0.1.0`.

## Execution map

Один green semantic milestone передаётся одним primary Composer 2.5 job через
Herdr. Job включает script/CI/docs/tests fallout одного release contract. Sol
выполняет whole-diff review, передаёт один consolidated correction packet при
необходимости и самостоятельно запускает broad verification.

## Scope

- Повысить declared minimum и canonical release toolchain до Go 1.26.6.
- Добавить pinned `govulncheck` mode и обязательный CI security gate.
- Сохранить основной full dry-run network-free и явно отделить online scan.
- Согласовать README, GoDoc, compatibility policy, readiness matrix, release
  checklist и changelog.
- Проверить отсутствие новых runtime dependencies и release side effects.

## Out of scope

- Новые detectors, изменение public API/runtime defaults или quality thresholds.
- Перегенерация исторических benchmark/evaluation reports, где Go 1.26.2 является
  частью зафиксированного measurement environment.
- Tag, push, GitHub Release, announcement или проверка module proxy после tag.

## Acceptance

1. `go.mod`, CI и current public docs согласованно требуют Go 1.26.6+.
2. Pinned vulnerability gate выполняется на exact minimum toolchain и блокирует
   release при достижимых findings.
3. Full dry-run остаётся network-free; online vulnerability scan документирован
   и проверяется отдельно локально и в CI.
4. Все обычные, race, fuzz, consumer, license, evaluation и OpenSpec checks зелёные.
5. Changelog описывает фактический `v0.1.0` scope без утверждения, что tag уже
   опубликован.

## Verification

```bash
GOTOOLCHAIN=go1.26.6 go test ./...
GOTOOLCHAIN=go1.26.6 go vet ./...
GOTOOLCHAIN=go1.26.6 go test -race ./...
GOTOOLCHAIN=go1.26.6 ./scripts/release-check.sh
GOTOOLCHAIN=go1.26.6 ./scripts/release-check.sh vuln
openspec validate v0-1-0-release-finalization --strict --no-interactive
openspec validate --all --strict --no-interactive
```

## Exit evidence

- Archived OpenSpec change with synced `oss-distribution` spec.
- Green offline dry-run and online vulnerability scan on Go 1.26.6.
- Reviewed CI/docs/changelog diff and experiment-log row.
- Чистый release checkpoint; publication остаётся отдельным action.
