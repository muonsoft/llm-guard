## MODIFIED Requirements

### Requirement: Internal priority model расширяем без изменения pipeline
Priority table MUST быть внутренней конфигурацией resolver, поддерживать все строковые `EntityType` и задавать enclosing URL более высокий priority, чем overlapping EMAIL или generic submatch; EMAIL и validated structured numeric entities MUST иметь более высокий priority, чем неизвестные custom entities. Неизвестные entities MUST получать единый deterministic default priority.

#### Scenario: URL содержит EMAIL-like fragment
- **WHEN** URL finding охватывает userinfo, path или query с overlapping EMAIL finding
- **THEN** resolver сохраняет полный URL и удаляет EMAIL submatch независимо от порядка candidates

#### Scenario: EMAIL конфликтует с custom entity
- **WHEN** EMAIL и custom entity имеют overlapping spans
- **THEN** resolver выбирает EMAIL независимо от порядка candidates

#### Scenario: Validated numeric entity конфликтует с generic numeric candidate
- **WHEN** BANK_CARD, SNILS или INN finding пересекается с неизвестным custom numeric entity
- **THEN** resolver сохраняет built-in structured finding

#### Scenario: Два custom entity
- **WHEN** конфликтуют неизвестные entity types
- **THEN** resolver применяет общий length, confidence и lexical tie-break без panic
