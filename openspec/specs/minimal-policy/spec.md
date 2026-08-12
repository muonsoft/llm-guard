# Minimal Policy Specification

## Purpose

Capability выбирает для resolved entity простое локальное действие allow, mask или block с безопасными defaults, не превращая core library в policy engine.

## Requirements

### Requirement: Action model строковая и валидируется при construction
Библиотека MUST экспортировать строковый `Action` со значениями `allow`, `mask`, `block` и immutable Guard options для secret-family и exact-entity overrides. Неизвестное action, пустая entity либо conflicting duplicate override MUST приводить к проверяемой safe configuration error при `New`; configuration error MUST NOT содержать sensitive input.

#### Scenario: Валидная family и entity configuration
- **WHEN** caller задаёт secret-family action и более точный action для одной entity
- **THEN** Guard создаётся, а exact-entity override имеет deterministic precedence над family action

#### Scenario: Невалидная configuration
- **WHEN** caller задаёт неизвестное action, пустую entity либо duplicate override
- **THEN** `New` возвращает ошибку, проверяемую как invalid configuration

### Requirement: Defaults безопасны и однозначны
Без overrides Guard MUST применять `mask` ко всем non-secret entities и `block` к `SECRET_JWT`, `SECRET_PRIVATE_KEY`, `SECRET_API_KEY` и `CONNECTION_STRING`. Secret-family override MUST применяться ко всем четырём secret entities, exact-entity override MUST иметь наивысший priority, а unknown custom entity MUST по умолчанию маскироваться.

#### Scenario: Default PII и custom action
- **WHEN** resolved set содержит PII либо unknown custom entity без policy override
- **THEN** Guard маскирует finding обратимым placeholder

#### Scenario: Default secret action
- **WHEN** resolved set содержит secret entity без policy override
- **THEN** Guard блокирует операцию и не возвращает result

### Requirement: Allow и mask не меняют resolution contract
После общего resolution action `allow` MUST оставлять исходный span byte-for-byte и не создавать token mapping, а `mask` MUST заменять span обычным collision-safe reversible token. Успешный `MaskResult.Findings` MUST содержать весь resolved set, включая allowed findings; TokenSet MUST содержать mappings только для masked findings.

#### Scenario: Mixed allow и mask
- **WHEN** resolved findings содержат allowed entity и masked entity без block
- **THEN** allowed text остаётся неизменным, masked text заменяется token, Findings содержит оба findings, а Restore раскрывает только masked mapping

#### Scenario: Все findings разрешены
- **WHEN** все resolved findings имеют action allow
- **THEN** Mask возвращает исходный text, полный resolved Findings и пустой caller-owned TokenSet без чтения namespace entropy

### Requirement: Block возвращает только safe zero result
Если хотя бы один resolved finding имеет action `block`, `Guard.Mask` MUST вернуть zero `MaskResult` и проверяемую `ErrBlocked`/typed block error. Error string, unwrap chain, exported typed fields, standard formatting и JSON-safe surfaces MUST содержать только несекретные aggregate metadata, достаточные для определения block, и MUST NOT содержать input, spans, entity sequence, raw fragment, masked text, token или JWT payload. Наличие allow/mask findings рядом MUST NOT приводить к partial result либо чтению namespace entropy.

#### Scenario: Secret блокирует mixed input
- **WHEN** resolved set содержит blocked secret вместе с mask/allow findings
- **THEN** Mask возвращает zero result, safe block error и не создаёт TokenSet или partial masked text

#### Scenario: Block error форматируется
- **WHEN** block error форматируется через `%v`, `%+v`, `%#v` либо анализируется через `errors.Is`/`errors.As`
- **THEN** block остаётся проверяемым, но ни один public representation не раскрывает sensitive content или fine-grained occurrence metadata

### Requirement: Policy immutable и concurrency-safe
Action configuration MUST быть скопирована при `New`, не предоставлять mutable global state и давать одинаковый результат при concurrent calls одного Guard. Caller context cancellation MUST иметь приоритет над policy evaluation и namespace generation.

#### Scenario: Concurrent mixed action masking
- **WHEN** несколько goroutines вызывают `Mask` одного configured Guard на одинаковом input
- **THEN** action decisions совпадают, masked invocations получают независимые TokenSet и race detector не сообщает shared-state races

#### Scenario: Cancellation до policy
- **WHEN** caller context отменён до `Mask`
- **THEN** возвращается caller context error и detectors/policy/entropy source не запускаются
