## Context

M0 предоставляет immutable concurrent-safe Guard, validated UTF-8 byte spans и deterministic aggregation, но не содержит built-in detectors, resolver или caller-owned runtime state. См. `proposal.md` и три delta specs. M1 должен сохранить root package provider-agnostic и не вводить process-global mappings.

## Goals / Non-Goals

**Goals:**

- Реализовать один end-to-end EMAIL path поверх существующего detection contract.
- Сделать resolver отдельным pure этапом, пригодным для следующих entity milestones.
- Сохранить reversible state только в opaque `TokenSet`, принадлежащем caller.
- Проверить security surfaces, collision handling, Unicode byte offsets и concurrency.

**Non-Goals:**

- RFC-perfect и internationalized EMAIL parser.
- Persistent/session token storage, heuristic restore и morphology correction.
- Другие PII entities, policy actions, observer/metrics и provider adapters.

## Decisions

### 1. Built-in detector реализуется через существующий Detector contract

Публичный `NewEmailDetector()` возвращает immutable detector, который caller регистрирует через `WithDetector`. Это сохраняет один extension mechanism и не добавляет специальную mutable регистрацию. Альтернатива `WithDefaultPIIDetectors()` отложена до появления полного structured pack в M2/M3.

EMAIL matching использует compiled ASCII candidate regexp плюс явную post-validation local-part, domain labels и Unicode-aware outer boundaries. Такой двухшаговый подход консервативнее единственного permissive regexp и проще fuzz-тестировать, чем RFC parser.

### 2. Resolver является pure public function с полным deterministic key

`Resolve(text string, findings []Finding)` повторно проверяет недоверенные candidates, сортирует copy по entity priority, span length, confidence и полному lexical tie-break, greedily выбирает непересекающиеся intervals, затем сортирует result в текстовом порядке. Input slice не мутируется.

Внутренняя priority table задаёт EMAIL выше неизвестных entities и расширяется следующими milestones без публичной конфигурации. Public priority option сейчас не добавляется: это преждевременно закрепило бы policy API до появления ADDRESS/URL/secrets cases. Interval tree отвергнут как лишняя сложность для небольших prompt finding sets.

### 3. Mask связывает Detect и Resolve, а replacement идёт справа налево

`Guard.Mask` вызывает `Detect`, затем `Resolve`, создаёт occurrence tokens и заменяет spans в descending `Start` order. Это позволяет использовать канонические byte offsets напрямую и сохраняет окружающий Unicode. `MaskResult.Findings` остаётся привязанным к исходному text, а не masked text.

Каждое occurrence получает отдельный global counter token, даже для одинакового value: это сохраняет однозначную позиционную identity и не вводит дополнительный equality signal. Формат не включает entity: `{{LLMG_<32 hex namespace>_<4+ digit counter>}}`.

### 4. Namespace использует 128 бит entropy и проверяется против input

По умолчанию Guard читает 16 bytes из `crypto/rand.Reader`. `WithRandomSource(io.Reader)` существует для deterministic tests и редких embedded integrations; Guard сериализует чтения mutex, а документация возлагает cryptographic quality custom source на caller.

Перед replacement строятся все tokens выбранного namespace и проверяются как substrings исходного text. При collision генерируется новый namespace, максимум 32 попытки; exhaustion и source failures возвращают typed safe error. UUID-per-token отвергнут из-за лишнего prompt noise.

### 5. Restore использует exact simultaneous replacement

TokenSet хранит приватный ordered mapping token → value. Restore строит exact replacer из mappings и выполняет один проход, поэтому unknown/mutated tokens остаются как есть, повтор known token восстанавливается везде, а вставленное value не запускает каскадную замену. Cross-TokenSet restore исключается namespace identity.

### 6. TokenSet не предоставляет чувствительные inspection surfaces

У TokenSet нет exported fields и методов enumeration. `String()` и `GoString()` возвращают фиксированное redacted описание без count, namespace или token. Для `encoding/json` реализуется явный safe representation, чтобы будущие exported metadata случайно не открыли mappings. Errors следуют `github.com/muonsoft/errors`, имеют sentinels/typed safe context и никогда не включают input fragments.

## Risks / Trade-offs

- [Консервативный EMAIL grammar даёт false negatives для редких RFC-valid addresses] → зафиксировать supported corpus и расширять только evidence-driven cases.
- [Greedy priority resolver может потребовать более сложной policy после ADDRESS/URL] → сохранить priority и compare logic изолированными и покрыть permutation invariants.
- [Custom random source может быть слабым или блокирующим] → secure default, явная документация caller responsibility и serialized access.
- [LLM может изменить placeholder] → exact restore намеренно оставляет mutated token без раскрытия; heuristic recovery вне scope.

## Migration Plan

Изменение additive: существующий `Detect` и custom detectors сохраняют поведение. README/example обновляются на новый EMAIL flow; rollback состоит в удалении новых additive API/files без миграции persisted data, поскольку TokenSet не покидает caller lifecycle.
