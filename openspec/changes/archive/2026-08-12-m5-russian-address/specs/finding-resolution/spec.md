## ADDED Requirements

### Requirement: Полный ADDRESS вытесняет вложенный PERSON
Resolver MUST назначать принятому ADDRESS finding более высокий entity priority, чем любому пересекающемуся PERSON finding. Результат MUST сохранять полный ADDRESS независимо от порядка candidates, длины либо confidence вложенного PERSON и MUST применять ту же deterministic overlap policy без address-specific post-processing в pipeline.

#### Scenario: Фамилия внутри названия улицы
- **WHEN** candidates для `ул. Академика Сахарова, 10` содержат ADDRESS на всю композицию и PERSON внутри street name
- **THEN** resolver возвращает только полный ADDRESS finding

#### Scenario: Переставленные candidates
- **WHEN** тот же ADDRESS/PERSON candidate set передан в разных permutations
- **THEN** resolver возвращает одинаковый ADDRESS finding в стабильном output order
