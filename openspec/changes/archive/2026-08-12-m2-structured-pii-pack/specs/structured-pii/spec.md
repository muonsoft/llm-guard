## ADDED Requirements

### Requirement: Основной structured pack доступен через единый публичный detection API
Корневой Go package MUST предоставлять immutable concurrent-safe built-in detectors со стабильными непустыми именами для `EntityPhone`, `EntityIPAddress`, `EntityURL`, `EntityINN`, `EntitySNILS` и `EntityBankCard`; каждый detector MUST регистрироваться через существующий `WithDetector` и возвращать findings без matched value.

#### Scenario: Embedded structured detection
- **WHEN** caller создаёт Guard с detectors основного structured pack и вызывает `Detect`
- **THEN** код компилируется через публичный API и возвращает findings соответствующих built-in entity types

#### Scenario: Concurrent structured detection
- **WHEN** один Guard одновременно обрабатывает несколько текстов всеми detectors pack
- **THEN** вызовы не имеют data race и не изменяют конфигурацию или результаты друг друга

### Requirement: PHONE detector консервативно распознаёт телефонные формы
PHONE detector MUST принимать российские номера с префиксом `+7` или `8`, распространёнными пробелами, дефисами и скобками, а также conservative international forms с явным `+` и допустимым количеством цифр. Он MUST возвращать span без окружающей punctuation и MUST отвергать произвольные длинные цифровые строки, неоднозначный local-only набор цифр, неверное количество цифр, letters внутри candidate и candidate внутри более длинного alphanumeric token.

#### Scenario: Российские форматированные номера
- **WHEN** Unicode текст содержит `+7 (999) 123-45-67` и `8 999 123 45 67`
- **THEN** detector возвращает два PHONE findings с точными UTF-8 byte spans полных номеров

#### Scenario: Неоднозначные длинные числа
- **WHEN** input содержит unprefixed длинные числа, неверно сгруппированные или слишком короткие phone-like fragments
- **THEN** detector не возвращает PHONE findings для этих fragments

### Requirement: IP detector подтверждает IPv4 и IPv6 семантически
IP detector MUST принимать полные валидные IPv4 и IPv6 addresses, включая compressed IPv6, только после semantic parsing; brackets вокруг IPv6 endpoint MUST оставаться вне finding span. Detector MUST отвергать IPv4 octets вне диапазона, неполные IPv4, malformed IPv6, zone identifiers и address-like substring внутри более длинного identifier.

#### Scenario: Валидные IPv4 и IPv6
- **WHEN** Unicode текст содержит `192.0.2.42`, `2001:db8::1` и `[2001:db8::2]`
- **THEN** detector возвращает IP_ADDRESS findings только для lexical address values с точными byte spans

#### Scenario: Malformed IP candidates
- **WHEN** input содержит `999.1.2.3`, `1.2.3`, `2001:::1` или `fe80::1%eth0`
- **THEN** detector не возвращает findings для этих candidates

### Requirement: URL detector сохраняет значимые URL boundaries
URL detector MUST принимать absolute `http` и `https` URL с DNS-like host или IP host, optional port, userinfo, path, query и fragment; returned span MUST включать весь syntactically valid URL, но не trailing sentence punctuation. Detector MUST отвергать relative paths, unsupported schemes, host без допустимой формы, whitespace/control characters и URL-like candidate внутри более длинного alphanumeric token.

#### Scenario: URL с credentials и query
- **WHEN** text содержит `https://user:pass@example.com:8443/a?q=one#part).`
- **THEN** detector возвращает один URL finding до closing punctuation, включая userinfo, port, path, query и fragment

#### Scenario: Unsupported и malformed URL
- **WHEN** input содержит relative path, `ftp://example.com`, host без dotted DNS suffix и URL с whitespace
- **THEN** detector не возвращает URL findings для этих fragments

### Requirement: INN detector проверяет оба формата checksum
INN detector MUST принимать только отдельные 10-digit INN юридического лица и 12-digit INN физического лица с корректной официальной checksum. Он MUST отвергать неверную длину, non-digit separators, failed checksum, homogeneous repeated digits и candidate, являющийся частью более длинной цифровой последовательности.

#### Scenario: Валидные INN
- **WHEN** text содержит корректные 10-digit и 12-digit INN с безопасными boundaries
- **THEN** detector возвращает два `EntityINN` findings с точными spans

#### Scenario: Неверная INN checksum
- **WHEN** в otherwise plausible INN изменена контрольная цифра
- **THEN** detector не возвращает finding

#### Scenario: Формальная checksum без допустимого identifier
- **WHEN** 10-digit или 12-digit candidate состоит из одной повторяющейся цифры и формально проходит arithmetic checksum
- **THEN** detector не возвращает finding

### Requirement: SNILS detector нормализует формат и проверяет checksum
SNILS detector MUST принимать отдельный 11-digit SNILS в compact форме и форме `XXX-XXX-XXX YY`, нормализовать только разрешённые separators для проверки и требовать корректную checksum по первым девяти цифрам. Он MUST отвергать pre-checksum legacy range, malformed или mixed separators, failed checksum и candidate внутри более длинной цифровой последовательности.

#### Scenario: Compact и formatted SNILS
- **WHEN** text содержит один корректный SNILS compact и один в форме `XXX-XXX-XXX YY`
- **THEN** detector возвращает findings по полным исходным spans обеих форм

#### Scenario: Неверная SNILS checksum
- **WHEN** в otherwise plausible SNILS изменены контрольные цифры
- **THEN** detector не возвращает finding

### Requirement: BANK_CARD detector требует card-like форму и Luhn
BANK_CARD detector MUST принимать отдельные card-like candidates длиной от 13 до 19 цифр с согласованными пробелами или дефисами либо без separators только после успешной Luhn validation. Он MUST отвергать failed checksum, одинаковые повторяющиеся цифры, mixed/malformed separators и candidate внутри более длинной цифровой последовательности; findings и diagnostics MUST NOT содержать normalized или исходный card number.

#### Scenario: Валидные card forms
- **WHEN** text содержит Luhn-valid compact card candidate и форматированный candidate с согласованными separators
- **THEN** detector возвращает `EntityBankCard` findings по полным исходным spans

#### Scenario: False-positive boundaries
- **WHEN** input содержит Luhn-invalid number, одинаковые цифры, mixed separators или candidate как часть более длинного числа
- **THEN** detector не возвращает BANK_CARD findings

### Requirement: Structured findings используют общий reversible pipeline
Каждый accepted structured finding MUST иметь confidence в диапазоне `[0,1]`, стабильную detector metadata и `[Start, End)` на UTF-8 rune boundaries. Existing `Guard.Mask` и `Guard.Restore` MUST обрабатывать все entity pack без entity-specific caller code, а corpus evaluation MUST сообщать positive, negative и malformed counts отдельно для каждой entity, а не только aggregate.

#### Scenario: Mixed structured round trip
- **WHEN** Guard со всем structured pack маскирует Unicode text с непересекающимися PHONE, IP_ADDRESS, URL, INN, SNILS и BANK_CARD
- **THEN** каждый accepted finding заменяется отдельным token, а Restore возвращает исходный text byte-for-byte

#### Scenario: Per-entity evaluation
- **WHEN** выполняется structured corpus evaluation
- **THEN** output содержит отдельные expected, detected, false-positive и false-negative counts для каждой entity
