## Context

См. motivation в `proposal.md`. После M4 библиотека уже имеет immutable `Detector`/`Guard`, общий deterministic resolver, byte-oriented findings и internal NLP runtime поверх `go-razdel`. R0 fixtures содержат пять обязательных ADDRESS positives и четыре conservative negatives; production graph не может зависеть от Natasha, Python, morphology data, внешних справочников или сети.

ADDRESS отличается от PERSON тем, что отдельные lexical parts недостаточны: detector должен сначала распознать bounded parts, затем принять только согласованную композицию и вернуть один maximal source span. Названия улиц могут содержать слова, похожие на PERSON, поэтому conflict resolution является частью product boundary.

## Goals / Non-Goals

**Goals:**

- построить один immutable finite scanner поверх существующих токенов и исходных byte offsets;
- явно зафиксировать supported part grammar и composition matrix до реализации;
- сохранить conservative precision boundary при punctuation, prefix/suffix street labels и extended building parts;
- дать воспроизводимую corpus/differential evidence и end-to-end ADDRESS/PERSON overlap coverage.

**Non-Goals:**

- generic location NER, morphology-compatible address facts, нормализация или разбор до ФИАС;
- полный порт Natasha/Yargy grammar, postcode extraction и безlabelные географические guesses;
- автоматическое включение ADDRESS detector в default Guard либо изменение публичных Finding/TokenSet contracts.

## Decisions

### 1. Расширить существующий internal NLP runtime bounded address scanner

`internal/nlp` получит address-specific immutable label tables, part annotations и scanner, использующий уже адаптированные `go-razdel` tokens и byte spans. Scanner не экспортирует grammar/normalization API: public wrapper преобразует только accepted spans в `EntityAddress` findings.

Так сохраняются R0 provenance boundary, единые cancellation points и UTF-8 semantics. Альтернатива с регулярным выражением на исходной строке хуже контролирует reordered labels, bounded multi-token names и maximal composition; новый parser dependency неоправдан для M5.

### 2. Использовать explicit-label grammar и mandatory street+house matrix

Accepted base всегда содержит street и следующий за ним house:

| Parts | Result |
|---|---|
| settlement/region/street/house по отдельности | reject |
| settlement + street | reject |
| street + house | accept |
| settlement + street + house | accept |
| street + house + corpus/building/apartment | accept maximal span |

Street grammar поддерживает prefix и suffix label forms для `улица`/`ул.`, `проспект`/`пр-т`, `переулок`/`пер.` и `шоссе`. Street name содержит bounded sequence из одного–четырёх word/ordinal tokens; labels являются обязательными, поэтому произвольная capitalized phrase с числом не становится адресом. `Академика Сахарова` допустимо как street name только внутри explicit street form.

House следует за street через bounded whitespace/comma separators и состоит из `дом`/`д.` плюс identifier либо из одного unlabeled identifier непосредственно в house position. Identifier ограничен цифрами с необязательной одиночной кириллической/латинской буквой, slash или hyphen segment; он не может быть частью более длинного alphanumeric token. Extended `корпус`/`корп.`, `строение`/`стр.` и `квартира`/`кв.` parts допускаются только после принятого house и требуют собственного identifier.

Settlement — необязательный prefix перед street: explicit `город`/`г.` либо bounded одно- или двухсловное capitalized/hyphenated name, отделённое запятой. Он расширяет уже доказанный street+house span, но никогда не создаёт finding самостоятельно. Такой asymmetric rule принимает обязательную форму `Москва, Тверская улица, 18`, не превращая `Москва` или `Москва, Тверская улица` в ADDRESS.

Альтернатива принимать `settlement+street` отклонена как слишком широкая для PII-first boundary. Postal index отложен: без отдельной boundary policy он конфликтует с generic numeric entities.

### 3. Отделить part parsing от maximal source-span composition

Scanner сначала строит internal candidate parts с исходными byte bounds, затем finite composer проверяет порядок и acceptance matrix. В итоговый span входят labels, meaningful internal separators и все согласованные extended parts; trailing comma, sentence punctuation и surrounding quotes не входят. При нескольких возможных starts выбирается maximal accepted composition, после чего scan продолжается с её конца; результат сортируется по source order.

Это упрощает review byte spans и исключает nested ADDRESS findings. Альтернатива выдавать каждую part наружу и собирать их в resolver отклонена: она раскрывает implementation entities и переносит domain grammar в общий conflict layer.

### 4. Реализовать ADDRESS как обычный opt-in detector

Public `NewAddressDetector() Detector` следует M4 pattern: stateless value, stable name `address`, context checks до tokenization и на bounded scan steps, fixed confidence для accepted composition. Guard configuration и detector concurrency model не меняются.

Численный heuristic score не добавляется: обязательная matrix уже является acceptance decision, а confidence не должен скрыто превращать слабые candidates в findings. Более сильные parts расширяют span, но не меняют resolver semantics.

### 5. Закрепить ADDRESS > PERSON в общей entity priority policy

Существующий resolver уже имеет central entity priority table и назначает ADDRESS более высокий rank, чем PERSON. M5 не добавляет post-processing: implementation проверяет и документирует этот contract permutation и end-to-end tests для nested PERSON-like street name.

Альтернатива подавлять PERSON внутри AddressDetector отклонена, потому что detectors выполняются независимо и не должны знать findings друг друга.

### 6. Версионировать product corpus отдельно от reference fixtures

`testdata/address/cases.jsonl` станет normative synthetic product corpus с category, expected byte spans и optional R0 reference/differential metadata. Existing pinned `testdata/natasha` fixtures остаются immutable reference baseline; новые M5 cases добавляются туда только если reference result уже воспроизводимо зафиксирован и проходит offline schema verification.

Black-box evaluator считает ADDRESS TP/FP/FN и exact-span rate и проваливает unexplained differential. `docs/address-quality-report.md` фиксирует corpus version, metrics, supported matrix, intentional differences и команды. Реальные адреса пользователей и исходные PII в fixtures не добавляются.

## Risks / Trade-offs

- [Bounded street-name grammar пропустит редкие адреса] → считать это осознанным precision-first limitation, покрыть обязательный/ambiguous corpus и расширять только отдельными reviewed cases.
- [Unlabeled house number может захватить соседнее число] → разрешать его только сразу после explicit street part и bounded separator; standalone numeric fragments отвергать.
- [Suffix street form создаёт неоднозначный start] → ограничить число name tokens, требовать known label и выбирать maximal composition только после успешного house.
- [Punctuation может расширить span за адрес] → composer включает только separators между принятыми parts и тестирует byte slice на compact/Unicode/sentence-boundary cases.
- [Resolver priority уже существует и regression может остаться незаметным] → добавить прямые permutation tests и end-to-end Guard test с overlapping ADDRESS/PERSON candidates.
- [Reference grammar шире product grammar] → каждое расхождение классифицировать в versioned corpus/report; отсутствие объяснения делает verification красным.

## Migration Plan

Изменение additive: consumer явно регистрирует новый detector. Rollback состоит в удалении регистрации либо откате новых файлов; существующие detectors, serialized findings и token format не меняются. Main specs синхронизируются перед archive только после green corpus, race и broad checks.
