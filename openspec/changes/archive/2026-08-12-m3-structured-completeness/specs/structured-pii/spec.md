## ADDED Requirements

### Requirement: Оставшийся structured PII pack доступен через публичный API
Корневой Go package MUST предоставлять immutable concurrent-safe built-in detectors со стабильными непустыми именами для `EntityPassport`, `EntityBankAccount` и `EntityDateOfBirth`; каждый detector MUST регистрироваться через существующий `WithDetector`, возвращать findings без matched value и работать в общем `Detect`/`Resolve`/`Mask`/`Restore` pipeline без entity-specific caller code.

#### Scenario: Полный structured round trip
- **WHEN** caller создаёт Guard с тремя новыми detectors и маскирует Unicode text с непересекающимися PASSPORT, BANK_ACCOUNT и DATE_OF_BIRTH
- **THEN** каждый accepted finding заменяется отдельным token, а Restore возвращает исходный text byte-for-byte

#### Scenario: Concurrent structured detection
- **WHEN** один Guard одновременно обрабатывает несколько текстов тремя новыми detectors
- **THEN** вызовы не имеют data race и не изменяют конфигурацию или результаты друг друга

### Requirement: PASSPORT требует российскую форму и явный контекст
PASSPORT detector MUST принимать российскую серию из четырёх цифр и номер из шести цифр только в contiguous форме `NNNN NNNNNN` или `NN NN NNNNNN`, когда candidate связан с явным case-insensitive RU marker `паспорт`, `паспорт РФ`, `серия` или `паспортные данные`. Он MUST сохранять span только числовой формы, нормализовать лишь разрешённые пробелы для проверки и отвергать candidate без marker, с неверной длиной, смешанными разделителями или внутри более длинной цифровой последовательности. Раздельная запись `серия NN NN, номер NNNNNN` вне одного contiguous numeric span MUST считаться неподдерживаемой в M3.

#### Scenario: Поддерживаемые формы паспорта
- **WHEN** text содержит `паспорт 45 08 123456` и `паспорт РФ 4508 654321`
- **THEN** detector возвращает два `EntityPassport` findings по полным числовым spans без context marker

#### Scenario: Раздельные серия и номер не поддерживаются
- **WHEN** text содержит `паспорт: серия 45 08, номер 123456`
- **THEN** detector не возвращает PASSPORT finding и не маскирует context marker вместе с цифрами

#### Scenario: Неоднозначные цифровые строки
- **WHEN** text содержит те же десять цифр без passport marker, номер договора, неверно сгруппированную форму или candidate как часть более длинного числа
- **THEN** detector не возвращает PASSPORT finding

### Requirement: BANK_ACCOUNT определяет пределы локальной проверки
BANK_ACCOUNT detector MUST принимать отдельный российский 20-digit account только после удаления разрешённых ASCII spaces из compact формы или пяти групп по четыре цифры. Candidate MUST иметь явный case-insensitive RU marker `р/с`, `расчетный счет`, `расчётный счёт`, `банковский счет` или `банковский счёт`; если в том же локальном реквизитном фрагменте присутствует девятизначный БИК, detector MUST дополнительно проверить стандартную контрольную сумму по последним трём цифрам БИК и account. Без БИК detector MUST считать форму только контекстно подтверждённой, а не внешне валидированной. Он MUST отвергать неверную длину, одинаковые повторяющиеся цифры, malformed/mixed separators, failed available checksum и candidate внутри более длинной цифровой последовательности.

#### Scenario: Контекстно подтверждённый счёт без БИК
- **WHEN** text содержит `расчётный счёт 40702810900000000001` без БИК
- **THEN** detector возвращает `EntityBankAccount` finding по 20 цифрам согласно документированному context-first fallback

#### Scenario: Счёт с доступным БИК
- **WHEN** реквизитный фрагмент содержит marker, 20-digit account и девятизначный БИК
- **THEN** detector возвращает BANK_ACCOUNT finding только если checksum комбинации БИК и account корректна

#### Scenario: Договорные и неконтекстные числа
- **WHEN** text содержит 20-digit contract/reference number без bank-account marker, homogeneous digits, malformed groups или checksum-invalid account рядом с БИК
- **THEN** detector не возвращает BANK_ACCOUNT finding

### Requirement: DATE_OF_BIRTH распознаётся только рядом с birth-context
DATE_OF_BIRTH detector MUST принимать календарно валидную numeric date `DD.MM.YYYY` или `DD/MM/YYYY` и русскую textual date `D <месяц в родительном падеже> YYYY [года]` только когда она непосредственно связана с case-insensitive RU marker `дата рождения`, `д.р.`, `родился` или `родилась`. Finding span MUST охватывать только дату без marker; detector MUST отвергать невозможную дату, двузначный год, неоднозначный порядок компонентов, дату без birth-context и date-like substring внутри более длинного token.

#### Scenario: Numeric и textual даты рождения
- **WHEN** text содержит `дата рождения: 12.10.1990` и `родился 3 марта 1985 года`
- **THEN** detector возвращает два `EntityDateOfBirth` findings, охватывающих только даты

#### Scenario: Обычная дата и реквизиты договора
- **WHEN** text содержит `встреча 12.10.2026`, `договор от 12.10.2026` или дату рядом с unrelated marker
- **THEN** detector не возвращает DATE_OF_BIRTH findings

#### Scenario: Невозможная дата
- **WHEN** birth-context содержит `31.02.1990` или malformed textual month
- **THEN** detector не возвращает DATE_OF_BIRTH finding

### Requirement: Structured corpus завершает per-entity coverage
Corpus evaluation MUST отдельно сообщать positive, negative и malformed counts для PASSPORT, BANK_ACCOUNT и DATE_OF_BIRTH наряду с существующими structured entities. Corpus MUST включать обычные даты, договорные реквизиты, conflicting numeric strings, Unicode boundaries и supported/unsupported forms; diagnostics MUST NOT содержать исходные или normalized sensitive values.

#### Scenario: Per-entity evaluation полного structured scope
- **WHEN** выполняется structured corpus evaluation
- **THEN** output содержит отдельные expected, detected, false-positive и false-negative counts для каждой новой entity и все expected cases совпадают без false positives
