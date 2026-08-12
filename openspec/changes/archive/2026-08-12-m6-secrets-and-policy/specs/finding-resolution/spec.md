## MODIFIED Requirements

### Requirement: Internal priority model расширяем без изменения pipeline
Priority table MUST быть внутренней конфигурацией resolver, поддерживать все строковые `EntityType` и задавать credential-bearing `CONNECTION_STRING` более высокий priority, чем overlapping URL, EMAIL либо generic submatch. Другие validated secret entities MUST иметь более высокий priority, чем URL, EMAIL и неизвестные custom entities; URL MUST сохранять более высокий priority, чем overlapping EMAIL или generic submatch; EMAIL и validated structured numeric entities MUST иметь более высокий priority, чем неизвестные custom entities. Неизвестные entities MUST получать единый deterministic default priority.

#### Scenario: Credential-bearing DSN пересекается с URL или EMAIL
- **WHEN** CONNECTION_STRING finding охватывает userinfo с overlapping URL либо EMAIL finding
- **THEN** resolver сохраняет полный CONNECTION_STRING и удаляет submatch независимо от порядка candidates

#### Scenario: Secret token пересекается с generic candidate
- **WHEN** validated JWT, private key либо provider token finding пересекается с неизвестной custom entity
- **THEN** resolver сохраняет secret finding независимо от порядка candidates

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
