# Known limitations (MVP)

Explicit boundaries for `github.com/muonsoft/llm-guard` v0.1.0 MVP. This list
does not promise future fixes or timelines.

## Detection scope

1. **Single-name PERSON gaps** — isolated given names or surnames without a
   supported multi-token pattern may not be detected.
2. **Conservative RU-first bias** — English and mixed-language edge cases may be
   missed or only partially supported.
3. **ADDRESS composition only** — single cities, regions, streets, or postal indices
   without the supported street+house (or settlement+street+house) pattern are not
   masked as ADDRESS.
4. **No ORGANIZATION entity** — company and institution names are out of scope.
5. **No semantic/ML NER** — rule-based and structural detectors only.
6. **Custom regexp boundaries are caller-owned** — no implicit word boundaries.

## Mask, restore, and placeholders

7. **No morphological restore** — `Restore` substitutes original substrings
   byte-for-byte and does not reconcile Russian case/gender changes introduced by
   the LLM after masking.
8. **Placeholder mutation risk** — if the model edits placeholder tokens, restore
   may miss or partially fail (`restore_miss` events when observed).
9. **No conversation-aware state** — each call is stateless; callers own
   `TokenSet` lifetime across turns.

## Security and policy

10. **No zero-FN guarantee** — conservative detectors may miss sensitive data;
    operators must not treat the library as exhaustive DLP.
11. **No prompt-injection protection** — text sanitization for model abuse is out
    of scope.
12. **No content moderation** — policy focuses on PII/secrets masking, not safety
    classification.
13. **No persistent token storage** — mappings exist only in caller RAM (`TokenSet`).
14. **Unsafe observer leaks by design** — `WithUnsafeDevelopmentObserver` must not
    be enabled in production.

## Operations and quality evidence

15. **Representative evaluation corpus** — unified `testdata/evaluation/cases.jsonl`
    is a smoke/regression boundary; family corpora are authoritative per entity.
16. **Benchmarks are not SLOs** — numbers vary by hardware; see
    [benchmark-baseline.md](benchmark-baseline.md) and
    [m8-quality-benchmark-comparison.md](m8-quality-benchmark-comparison.md).
17. **No exporter server / OTel / Prometheus in core** — safe observer callbacks
    only.
18. **Pre-tag consumer check uses local replace** — proves API boundary before
    publish; module proxy fetch of an unpublished tag is a separate post-release
    step.

## Post-MVP exclusions

- OpenAI Chat/Responses adapters and HTTP proxy
- External gateway deployment model
- ML/NER dictionaries copied from Natasha/OpenCorpora
- Distributed tracing and persistent audit stores

For supply-chain boundaries on reference tooling, see
[natasha-license-inventory.md](natasha-license-inventory.md) and
[dependency-license-inventory.md](dependency-license-inventory.md).
