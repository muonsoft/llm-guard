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
