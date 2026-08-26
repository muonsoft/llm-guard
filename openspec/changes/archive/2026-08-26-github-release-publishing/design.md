## Context

См. `proposal.md` — Why. Сейчас `release-check.yml` имеет read-only permissions и
на tag либо manual dispatch выполняет только dry-run. У `comm-relay` подтверждён
успешными GitHub runs другой lifecycle: ручной SemVer dispatch, validate,
release metadata commit и GitHub Release. `llm-guard` — библиотека, поэтому ей не
нужна build matrix артефактов; semantic tag является распространяемым артефактом.

## Goals / Non-Goals

**Goals:**

- Дать maintainer один явный `Release` workflow для публикации после зелёного CI.
- Гарантировать, что tag указывает только на commit, прошедший release gates, с
  финализированной changelog section.
- Сохранить least privilege и безопасную обработку workflow input.
- Оставить обычный PR/push CI и локальный dry-run без side effects.

**Non-Goals:**

- Не публиковать release в ходе этого milestone.
- Не собирать binary archives или менять Go runtime dependencies.
- Не автоматизировать external diagnostic dataset download как обязательную PR
  dependency.

## Decisions

### 1. Основной путь — `workflow_dispatch`, tag push — строгий fallback

Ручной dispatch с SemVer повторяет фактически используемый путь `comm-relay` и
сохраняет решение о публикации за maintainer. Workflow разрешает tag push для
уже подготовленного release commit, но не меняет `main` и не перемещает tag.
Альтернатива `workflow_run` после каждого зелёного CI отклонена: она могла бы
публиковать без отдельного release intent.

### 2. Release workflow сам повторяет authoritative gates

Workflow выполняет полный offline dry-run и exact-minimum pinned vulnerability
scan. Он не доверяет только conclusion отдельного CI run: повторная validation
делает запуск воспроизводимым и совпадает с моделью `comm-relay`. Dispatch также
проверяет, что исходный SHA остаётся актуальным `main` перед release commit, чтобы
конкурирующий push не попал под tag без проверки.

### 3. Changelog финализируется до создания tag

Идемпотентный script поддерживает текущую planned section `0.1.0`, будущий
непустой `[Unreleased]`, `--check-only` и строгую проверку уже подготовленного tag.
При dispatch publication job создаёт только metadata commit, push-ит его на
`main`, затем передаёт его точный SHA как target commit для GitHub Release/tag.
Альтернатива финализировать changelog после tag отклонена: source archive версии
содержал бы pre-release metadata.

### 4. Permissions выдаются на уровне job

Workflow default — `contents: read`; только publication job после `needs:
validate` получает `contents: write`. User input передаётся через environment, а
не интерполируется в shell program. Все release runs сериализуются concurrency
group. Это безопаснее top-level write permissions из исходного `comm-relay`, при
сохранении того же user flow.

### 5. Публичные документы используют state-neutral lifecycle wording

Исходники до tag и внутри tagged module не должны одновременно утверждать, что
release ещё не существует. README описывает условный install path и явный
workflow, а датированный CHANGELOG становится источником release notes.

## Risks / Trade-offs

- **[Bot push перестанет работать при будущей branch protection]** → workflow
  явно завершится до tag; maintainer сможет заранее финализировать changelog через
  PR и использовать tag fallback.
- **[Validation дублирует зелёный CI и занимает больше времени]** → это осознанная
  цена за проверку exact release SHA и networked vulnerability gate.
- **[Push release metadata не запускает новый CI из-за `GITHUB_TOKEN`]** → commit
  ограничен changelog; workflow проверяет его diff и точный SHA до tag.
- **[Два maintainer запуска конфликтуют]** → единый concurrency group не допускает
  параллельной publication.

## Migration Plan

1. Залить F2 на `main` и дождаться зелёного обычного CI.
2. В GitHub Actions открыть `Release`, выбрать `main`, задать `v0.1.0` и запустить.
3. Workflow повторно проверит candidate, создаст release metadata commit, tag и
   GitHub Release.
4. Проверить `go get github.com/muonsoft/llm-guard@v0.1.0` из чистого module.

Rollback до публикации — исправить main и перезапустить. После публичного tag
историю не переписывать; использовать documented retraction/new patch release.
