## Purpose

Capability локально и консервативно обнаруживает поддерживаемые credential formats без network verification, раскрытия payload или зависимости от provider SDK.

## ADDED Requirements

### Requirement: Secret entities используют общий detector contract
Библиотека MUST предоставлять строковые entities `SECRET_JWT`, `SECRET_PRIVATE_KEY`, `SECRET_API_KEY` и `CONNECTION_STRING` и отдельные immutable built-in detectors для JWT, PEM, provider token и DSN. Findings MUST содержать только byte spans, entity, confidence и стабильное detector name; raw secret либо derived payload values MUST NOT попадать в public metadata.

#### Scenario: Общий Guard pipeline
- **WHEN** caller регистрирует все secret detectors через публичные constructors и передаёт текст с поддерживаемыми credentials
- **THEN** `Guard.Detect` возвращает валидные findings общего формата в стабильном порядке

#### Scenario: Safe metadata
- **WHEN** finding, detector либо detection error форматируется стандартными средствами Go
- **THEN** output не содержит raw credential, JWT payload, PEM body либо DSN password

### Requirement: JWT detector проверяет compact structure
JWT detector MUST принимать только три непустых base64url-сегмента с корректными token boundaries, JSON object header и непустыми payload/signature segments. Header MUST содержать строковый непустой `alg`, `alg: none` MUST отклоняться; detector MUST NOT интерпретировать, сохранять либо возвращать claims из payload и MUST NOT выполнять signature verification.

#### Scenario: Структурно корректный signed JWT
- **WHEN** text содержит bounded compact JWT с JSON header, допустимым `alg` и тремя корректными base64url segments
- **THEN** detector возвращает точный `SECRET_JWT` byte span

#### Scenario: Malformed или unsecured JWT
- **WHEN** candidate имеет неверное число segments, invalid base64url/header JSON, пустой segment либо `alg: none`
- **THEN** detector не возвращает finding и не включает candidate content в error

### Requirement: PEM detector принимает только private-key blocks
PEM detector MUST принимать bounded complete blocks с совпадающими `BEGIN`/`END` labels только для `PRIVATE KEY`, `RSA PRIVATE KEY`, `EC PRIVATE KEY`, `OPENSSH PRIVATE KEY` и `PGP PRIVATE KEY BLOCK`, с непустым syntactically valid base64 body и без захвата окружающего текста. Public output MUST NOT содержать decoded key material.

#### Scenario: Полный поддерживаемый private key block
- **WHEN** text содержит complete PEM block с поддерживаемым label и валидным body
- **THEN** detector возвращает один точный `SECRET_PRIVATE_KEY` span от `BEGIN` до `END`

#### Scenario: Public key или broken block
- **WHEN** block является public/certificate block, имеет mismatched label, invalid body либо отсутствующий footer
- **THEN** detector не возвращает finding

### Requirement: Provider token patterns имеют versioned conservative boundary
Provider detector MUST поддерживать только зафиксированные в repository version note префиксы и shapes: GitHub `ghp_` и `github_pat_`, GitLab `glpat-`, OpenAI-like `sk-`/`sk-proj-` с достаточной длиной и AWS access-key IDs `AKIA`/`ASIA` общей длины 20. Каждый pattern MUST требовать non-token boundaries и documented minimum/exact shape, чтобы embedded либо truncated values не совпадали. Version note MUST содержать дату snapshot, официальные source URLs, exact supported shapes и порядок обновления corpus; неподтверждённые либо новые provider versions MUST не расширяться эвристически.

#### Scenario: Versioned provider positives
- **WHEN** text содержит bounded token одной из явно поддерживаемых shapes
- **THEN** detector возвращает точный `SECRET_API_KEY` span и versioned corpus связывает case с family/shape

#### Scenario: Prefix-like prose и malformed tokens
- **WHEN** text содержит только prefix, слишком короткий body, invalid alphabet, embedded candidate либо unsupported provider prefix
- **THEN** detector не возвращает finding

### Requirement: DSN detector требует authority credentials
DSN detector MUST использовать conservative URL parsing и принимать только абсолютные connection schemes `postgres`, `postgresql`, `mysql`, `mongodb`, `mongodb+srv`, `redis`, `rediss`, `amqp` и `amqps`, когда authority содержит непустые username и password. Finding MUST охватывать полный DSN, включая query/fragment, но error и metadata MUST NOT содержать userinfo. Обычные `http`/`https` URL, passwordless DSN, relative strings и credentials только в query MUST не считаться connection secret.

#### Scenario: Credential-bearing connection string
- **WHEN** text содержит bounded supported DSN с непустыми username и password в authority
- **THEN** detector возвращает точный `CONNECTION_STRING` span

#### Scenario: Обычный URL или passwordless DSN
- **WHEN** text содержит `https` URL, supported scheme без password либо query-only password
- **THEN** DSN detector не возвращает finding

### Requirement: Secret detectors безопасны при cancellation и concurrency
Каждый built-in secret detector MUST проверять caller context, не использовать mutable global state и возвращать одинаковый ordered result при concurrent use одного instance. Positive, negative и malformed corpus MUST быть versioned, не содержать live credentials и проверять exact byte spans.

#### Scenario: Context отменён до detection
- **WHEN** caller context отменён до вызова secret detector
- **THEN** detector возвращает nil findings и проверяемую context error без сканирования input

#### Scenario: Concurrent corpus detection
- **WHEN** несколько goroutines используют один secret detector на одинаковом corpus input
- **THEN** результаты совпадают и race detector не сообщает shared-state races
