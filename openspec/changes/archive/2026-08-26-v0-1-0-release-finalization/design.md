## Context

См. `proposal.md` и delta `specs/oss-distribution/spec.md`. Текущий offline
`scripts/release-check.sh` проходит на Go 1.26.2, но не включает vulnerability
database lookup. Официальный `govulncheck` на Go 1.26.2 обнаруживает девять
достижимых standard-library findings; тот же код на Go 1.26.6 проходит tests и
scan без findings. Core остаётся pure Go library без новых runtime dependencies.

## Goals / Non-Goals

**Goals:**

- Сделать Go 1.26.6 единственным current minimum во всех executable contracts.
- Добавить воспроизводимый, pinned и блокирующий vulnerability gate.
- Сохранить существующий full dry-run network-free и side-effect-free.
- Финализировать согласованные release docs и changelog до maintainer approval.

**Non-Goals:**

- Не менять detector/masking/restore semantics, public Go API или defaults.
- Не добавлять `govulncheck` в runtime/module dependency graph.
- Не переписывать historical measurement metadata, зафиксированную на Go 1.26.2.
- Не создавать tag, push, GitHub Release или announcement.
- Не обновлять внешний benchmark до выбора окончательного release commit.

## Decisions

### 1. Minimum Go повышается до 1.26.6 целиком

Выбор: обновить `go.mod`, current GoDoc/README/compatibility/release claims и все
CI/release workflow setup-go pins до 1.26.6.

Rationale: поддержка 1.26.2 противоречит security preflight, а проверка на stable
не доказывает безопасность заявленного minimum. Альтернатива — оставить 1.26.2
и рекомендовать patched compiler — отвергнута как двусмысленный contract.

Historical reports с measurement environment `go1.26.2` остаются неизменными:
это evidence прошлого запуска, а не current support claim.

### 2. Vulnerability scan — отдельный pinned online mode

Выбор: добавить `./scripts/release-check.sh vuln`, запускающий
`go run golang.org/x/vuln/cmd/govulncheck@v1.7.0 ./...`, и отдельный обязательный
CI job на Go 1.26.6.

Rationale: pin делает tool behavior воспроизводимым, а отдельный mode честно
отражает network/cache dependency. Включение scan в full mode отвергнуто, потому
что разрушает documented network-free dry-run. Добавление tool directive в
`go.mod` отвергнуто, чтобы release tooling не расширяло consumer module graph и
license inventory shipped dependencies.

### 3. Readiness — conjunction offline и security gates

Выбор: README/readiness matrix/release checklist считают repository готовым только
при зелёных full dry-run и `vuln` mode на exact minimum. CI отражает то же
разделение отдельными jobs.

Rationale: существующее правило «full script green ⇒ ready» пропускает advisory
проверку. Раздельные evidence сохраняют offline reproducibility и security
актуальность без смешения гарантий.

### 4. Changelog финализируется без ложной публикации

Выбор: перенести фактический предрелизный scope в planned `0.1.0`, оставить
`Unreleased` для изменений после release candidate и явно сохранить manual
publication boundary.

Rationale: release notes должны описывать содержимое тега до approval, но не
утверждать существование ещё не созданного release.

## Risks / Trade-offs

- [Pinned scanner со временем устареет] → версия обновляется отдельным reviewed
  change; vulnerability DB остаётся текущей при каждом запуске.
- [Online scan недоступен при outage] → offline dry-run остаётся доступным, но
  maintainer approval блокируется до получения security evidence.
- [Patch-level minimum уменьшает compatibility] → изменение явно отражается во
  всех public contracts и changelog; это безопаснее поддержки уязвимой toolchain.
- [Stable CI может использовать более новый Go] → exact 1.26.6 job остаётся
  canonical minimum evidence, stable — только forward signal.

## Migration Plan

1. Обновить implementation/docs/CI одним change и прогнать focused checks.
2. Выполнить full offline dry-run и pinned vulnerability scan на Go 1.26.6.
3. Sync delta `oss-distribution`, strict validate и archive change.
4. Maintainer выбирает release commit, обновляет external diagnostic evidence при
   необходимости и отдельно выполняет manual tag/push/publish checklist.

Rollback до publication: вернуть этот change целиком и восстановить Go 1.26.2
claims. После public tag minimum нельзя понижать без отдельного compatibility и
security review.
