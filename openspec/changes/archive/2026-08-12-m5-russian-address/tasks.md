## 1. Core runtime

- [x] 1.1 Зафиксировать internal part kinds и immutable label tables для settlement, street, house, corpus, building и apartment.
- [x] 1.2 Реализовать bounded street-name annotation для prefix/suffix `улица`, `проспект`, `переулок` и `шоссе` forms.
- [x] 1.3 Реализовать bounded house identifier с explicit и unlabeled forms, outer token boundaries и cancellation points.
- [x] 1.4 Реализовать optional settlement prefix, который расширяет только доказанную street+house композицию.
- [x] 1.5 Реализовать trailing corpus/building/apartment parts и maximal accepted source span.
- [x] 1.6 Покрыть internal scanner tests для byte offsets, compact punctuation, multiple candidates и cancellation без partial spans.

## 2. Detector

- [x] 2.1 Реализовать immutable `NewAddressDetector() Detector` со стабильным именем и `EntityAddress`.
- [x] 2.2 Покрыть mandatory direct forms `г. Москва, ул. Тверская, д. 18`, `Москва, Тверская улица, дом 18`, `ул. Ленина, 10` и `проспект Мира, д. 101`.
- [x] 2.3 Покрыть extended `корпус`, `строение` и `квартира` forms одним maximal finding.
- [x] 2.4 Добавить compact punctuation, Unicode context, sentence boundary и two-address detector tests с exact byte slices.
- [x] 2.5 Добавить conservative negative tests для isolated settlement/region/street/number, settlement+street и embedded token fragments.
- [x] 2.6 Добавить context cancellation и shared-detector concurrency tests.

## 3. Resolution и masking

- [x] 3.1 Добавить resolver permutation regression для ADDRESS над nested PERSON независимо от confidence и candidate order.
- [x] 3.2 Добавить Guard Detect/Resolve regression для `ул. Академика Сахарова, 10` без PERSON leakage.
- [x] 3.3 Добавить end-to-end ADDRESS Mask/Restore test с одним token на полный span и byte-for-byte round trip.
- [x] 3.4 Добавить concurrent Guard/Mask test с независимыми caller-owned TokenSet.

## 4. Corpus и audit evidence

- [x] 4.1 Создать versioned synthetic `testdata/address/cases.jsonl` с positive, negative и ambiguous classes и exact byte spans.
- [x] 4.2 Реализовать black-box ADDRESS corpus evaluator с TP/FP/FN, precision, recall и exact-span rate.
- [x] 4.3 Связать R0 cases с pinned `testdata/natasha` fixtures и проваливать unexplained differentials.
- [x] 4.4 Проверить offline reference schema командой `python3 tools/natasha-reference/reference.py verify --offline --cases testdata/natasha/cases.jsonl --expected testdata/natasha/expected-python.jsonl`.
- [x] 4.5 Проверить production dependency graph через `go list -deps ./...`, `go mod graph` и `go mod tidy -diff`.

## 5. Tests

- [x] 5.1 Выполнить focused suite `go test ./... -run 'Test(Address|Addr|Resolve.*Address|Mixed)' -count=1`.
- [x] 5.2 Выполнить race-focused suite `go test -race ./... -run 'TestAddress' -count=1`.
- [x] 5.3 Повторить ADDRESS corpus evaluation и подтвердить нулевые FP/FN mandatory corpus.
- [x] 5.4 Выполнить `go test ./... -count=1`, `go vet ./...` и `go test -race ./... -count=1`.

## 6. Docs

- [x] 6.1 Обновить README built-in detector list, opt-in usage и conservative composition boundary.
- [x] 6.2 Добавить runnable ADDRESS Mask/Restore example с исходным UTF-8 round trip.
- [x] 6.3 Создать `docs/address-quality-report.md` с corpus version, composition matrix, metrics, pinned reference revisions и intentional differences.
- [x] 6.4 Документировать unsupported postal index, standalone location, geocoding и normalization boundaries.

## 7. Verification

- [x] 7.1 Выполнить `openspec validate "m5-russian-address" --strict --no-interactive`.
- [x] 7.2 Выполнить `openspec validate --specs --strict --no-interactive` после sync delta specs.
- [x] 7.3 Сверить implementation каждого requirement/scenario с `specs/russian-address/spec.md` и `specs/finding-resolution/spec.md` и сохранить safe verification evidence.
- [x] 7.4 Зафиксировать review ADDRESS/PERSON overlap и отсутствие CRITICAL/WARNING перед archive.
