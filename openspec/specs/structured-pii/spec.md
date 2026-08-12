# Structured PII Specification

## Purpose

Capability предоставляет встроенные консервативные structured PII detectors, начиная с EMAIL, которые возвращают проверяемые UTF-8 byte spans без хранения найденных значений.

## Requirements

### Requirement: EMAIL detector доступен через публичный detection API
Корневой Go package MUST предоставлять публичный способ создать immutable EMAIL detector и зарегистрировать его через существующий `WithDetector`; один detector MUST быть безопасен для concurrent calls и сообщать стабильное непустое имя.

#### Scenario: Embedded EMAIL detection
- **WHEN** caller создаёт Guard с built-in EMAIL detector и вызывает `Detect`
- **THEN** код компилируется через публичный API и возвращает EMAIL findings

#### Scenario: Concurrent EMAIL detection
- **WHEN** один Guard одновременно обрабатывает несколько текстов с EMAIL
- **THEN** вызовы не имеют data race и не изменяют конфигурацию или результаты друг друга

### Requirement: EMAIL matching консервативно принимает common ASCII mailbox
EMAIL detector MUST принимать распространённые ASCII mailbox forms с буквами или цифрами, допустимыми разделителями local-part и DNS-like domain как минимум с одной точкой; он MUST сохранять исходный регистр и возвращать span ровно по mailbox без окружающей punctuation.

#### Scenario: Common mailbox с punctuation
- **WHEN** текст содержит `Напишите (Ivan.Petrov+sales@example-domain.ru).`
- **THEN** detector возвращает один EMAIL finding, охватывающий только `Ivan.Petrov+sales@example-domain.ru`

#### Scenario: Несколько mailbox в Unicode тексте
- **WHEN** Unicode текст содержит несколько разделённых корректных mailbox
- **THEN** detector возвращает для каждого mailbox корректные UTF-8 byte offsets

### Requirement: EMAIL boundary rules отсекают неоднозначные candidates
EMAIL detector MUST отвергать local-part с leading, trailing или consecutive dot, domain label с leading или trailing hyphen, domain без dotted suffix и candidate, у которого непосредственно за границей поддерживаемого ASCII grammar находится Unicode letter/digit или допустимый mailbox-символ. Detector MUST принимать syntactically valid DNS-like suffix без эвристического allow/deny list по известным TLD prefixes. Detector MUST NOT пытаться реализовать полный RFC mailbox grammar или internationalized local-part в M1.

#### Scenario: Invalid mailbox forms
- **WHEN** input содержит `.user@example.com`, `user..name@example.com`, `user@-example.com`, `user@example`, `Жuser@example.com`, `user@example.comЖ` или `user@example.com_`
- **THEN** эти fragments не возвращаются как EMAIL findings

#### Scenario: DNS-like suffix не фильтруется по prefix
- **WHEN** input содержит syntactically valid mailbox с suffix `.company`, `.coffee` или `.community`
- **THEN** detector возвращает полный EMAIL finding без попытки угадать более короткий TLD

#### Scenario: Placeholder-like и URL text
- **WHEN** input содержит placeholder-like fragment, URL без mailbox и одиночный `@`
- **THEN** detector не создаёт ложный EMAIL finding

### Requirement: EMAIL findings совместимы с core validation
Каждый EMAIL finding MUST иметь `EntityEmail`, confidence в диапазоне `[0,1]`, стабильную detector metadata и `[Start, End)` на UTF-8 rune boundaries; finding MUST NOT содержать matched value.

#### Scenario: Core принимает EMAIL finding
- **WHEN** EMAIL detector работает через Guard на валидном UTF-8 input
- **THEN** существующая core validation принимает finding без нормализации byte offsets

#### Scenario: Отменённый context
- **WHEN** context отменён до или во время EMAIL detection
- **THEN** detector завершает работу без partial findings согласно core cancellation contract

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
