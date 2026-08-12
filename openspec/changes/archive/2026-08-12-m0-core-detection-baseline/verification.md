# Verification Report: m0-core-detection-baseline

## Summary

| Dimension | Status |
|---|---|
| Completeness | 11/11 tasks; 7/7 requirements implemented |
| Correctness | 17/17 scenarios covered by code, tests or repository checks |
| Coherence | Design decisions followed; no issues |

## Completeness

- Все 11 checklist items в `tasks.md` отмечены после независимого review и выполнения focused/broad checks.
- Public contracts находятся в `entity.go`, `finding.go`, `detector.go`, `errors.go` и `guard.go`.
- Validation и deterministic aggregation находятся в `validate.go` и `guard.go`.
- Package/example/race evidence находится в `doc.go`, `example_test.go`, `detector_test.go`, `finding_test.go` и `guard_test.go`.
- ADR и CI evidence находятся в `docs/adr/` и `.github/workflows/ci.yml`.

## Correctness

- Public API, empty Guard и invalid configuration покрыты `detector_test.go:36` и `guard_test.go:50`–`guard_test.go:121`.
- Unicode byte ranges, invalid spans/entity/confidence/input и metadata покрыты `finding_test.go:38`–`finding_test.go:205` и `guard_test.go:122`–`guard_test.go:134`.
- Completion-order independence, full sort, concurrent Guard и race behavior покрыты `guard_test.go:257`–`guard_test.go:320` и `guard_test.go:410`–`guard_test.go:458`.
- Nil-on-failure, deterministic multiple errors, safe cause redaction и sibling cancellation покрыты `guard_test.go:187`–`guard_test.go:255`, `guard_test.go:321`–`guard_test.go:375` и `guard_test.go:477`–`guard_test.go:636`.
- Caller context cancellation покрыта `guard_test.go:135`–`guard_test.go:185`; standalone detector context errors отдельно проверены `guard_test.go:519`–`guard_test.go:550`.
- Example компилируется в `example_test.go:27`; repository commands проходят без внешних сервисов.

## Coherence

- Root `llmguard` package и закрытая option model соответствуют design.
- Detector names фиксируются в `New`; Guard configuration после construction не меняется.
- Один goroutine на detector пишет в индексированный slot; метод ждёт все calls и отдаёт приоритет caller context.
- UTF-8 validation выполняется до detector calls, findings копируются и сортируются по полному contract key.
- Typed errors сохраняют `errors.Is/As`, а публичное форматирование не включает raw detector cause или input.
- Runtime dependencies ограничены `github.com/muonsoft/errors`; provider/proxy/logging/persistence imports отсутствуют.

## Issues

- CRITICAL: 0
- WARNING: 0
- SUGGESTION: 0

## Verification evidence

```text
go test ./... -run 'TestGuard|TestFinding|TestDetector' -count=1                 PASS
go test -race ./... -run 'TestGuard' -count=1                                  PASS
go test ./... -run 'TestGuard_When(InvalidFinding|MultipleDetectorErrors|SiblingFailure|LaterSubstantive|DetectorReturnsContext|DetectorReturnsDeadline)' -count=100  PASS
go test ./... -count=1                                                         PASS
go vet ./...                                                                   PASS
go test -race ./... -count=1                                                   PASS
openspec validate m0-core-detection-baseline --strict --no-interactive         PASS
```

## Final assessment

Все проверки пройдены. Change готов к sync и archive.
