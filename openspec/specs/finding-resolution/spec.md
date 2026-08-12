# Finding Resolution Specification

## Purpose

Capability превращает валидные candidate findings в стабильный непересекающийся набор по общей deterministic priority policy, пригодный для последующего masking.

## Requirements

### Requirement: Resolver повторно валидирует недоверенные входы
Публичный resolver MUST отклонять invalid UTF-8 text, пустую entity или detector metadata, non-finite либо выходящую из `[0,1]` confidence, пустые или выходящие за input spans и границы внутри UTF-8 rune. Ошибка MUST быть проверяемой и MUST NOT содержать input или найденный substring.

#### Scenario: Invalid candidate span
- **WHEN** caller передаёт finding с границей внутри многобайтовой rune
- **THEN** resolver возвращает nil result и safe invalid-finding error

#### Scenario: Invalid resolver text
- **WHEN** caller передаёт строку с invalid UTF-8
- **THEN** resolver возвращает nil result и ошибку, проверяемую как invalid text

### Requirement: Resolver детерминированно удаляет дубликаты и overlaps
Resolver MUST выбирать candidates общим полным ключом: internal entity priority по убыванию, длина span по убыванию, confidence по убыванию, `Start`, `End`, entity и detector лексикографически, затем исходный индекс только для полностью равных candidates. Candidate MUST быть отброшен, если его полуинтервал пересекает уже выбранный; соседние полуинтервалы MUST считаться непересекающимися. Exact duplicates MUST приводить к одному finding.

#### Scenario: Exact duplicates
- **WHEN** resolver получает несколько полностью одинаковых findings
- **THEN** result содержит ровно один экземпляр

#### Scenario: Nested findings разного priority
- **WHEN** высокоприоритетный EMAIL finding содержит или пересекает generic lower-priority finding
- **THEN** resolver сохраняет EMAIL и удаляет конфликтующий finding

#### Scenario: Equal-priority overlap
- **WHEN** два findings одинакового priority пересекаются
- **THEN** resolver выбирает более длинный, затем более уверенный candidate по полному tie-break key

#### Scenario: Adjacent spans
- **WHEN** `End` одного finding равен `Start` другого
- **THEN** resolver сохраняет оба finding

### Requirement: Internal priority model расширяем без изменения pipeline
Priority table MUST быть внутренней конфигурацией resolver, поддерживать все строковые `EntityType` и задавать EMAIL более высокий priority, чем неизвестным custom entities; неизвестные entities MUST получать единый deterministic default priority.

#### Scenario: EMAIL конфликтует с custom entity
- **WHEN** EMAIL и custom entity имеют overlapping spans
- **THEN** resolver выбирает EMAIL независимо от порядка candidates

#### Scenario: Два custom entity
- **WHEN** конфликтуют неизвестные entity types
- **THEN** resolver применяет общий length, confidence и lexical tie-break без panic

### Requirement: Resolved output имеет стабильный текстовый порядок
После conflict selection resolver MUST возвращать findings по `Start`, затем `End`, entity, detector и confidence по убыванию; output MUST быть валидным, непересекающимся и не зависеть от map iteration или concurrency timing.

#### Scenario: Candidates переданы в другом порядке
- **WHEN** один и тот же набор различимых candidates передан в разных permutations
- **THEN** resolver возвращает одинаковую последовательность findings

#### Scenario: Empty candidates
- **WHEN** resolver получает пустой список для валидного input
- **THEN** он возвращает пустой result без ошибки
