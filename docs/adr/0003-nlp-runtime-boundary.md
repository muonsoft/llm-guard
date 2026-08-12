# ADR 0003: Internal bounded NLP runtime with external go-razdel

## Status

Accepted (R0)

## Context

The MVP needs conservative Russian PERSON and compositional ADDRESS detection.
Natasha 0.8 provides useful reference behavior but transitively brings a generic
Yargy chart parser, Pymorphy, OpenCorpora-derived data, normalization/fact
interpretation, and dictionaries with unsuitable or unresolved production data
provenance. The product returns an entity and exact source span, not a normalized
Natasha fact.

`github.com/muonsoft/go-razdel` already exists as a standalone pure-Go tokenizer
with UTF-8 byte offsets. Its token boundaries intentionally differ from Yargy's
regex tokenizer for hyphenated words.

## Decision

- Keep `go-razdel` as a separate production dependency and preserve its source
  text and UTF-8 byte offsets through a thin internal adapter.
- Implement product-specific token annotations, PERSON/ADDRESS rules and span
  composition under `internal/` in `llm-guard`.
- Express only finite sequences, alternatives, bounded optionals/repetition and
  direct predicates required by the M4/M5 corpus. Do not port Yargy's Earley
  chart, generic grammar DSL, relation graph, interpretation tree, or fact API.
- Use project-authored immutable role/suffix/alias tables. Do not embed or
  download Natasha dictionaries, Pymorphy/OpenCorpora data, or a full generic
  morphological analyzer in the MVP.
- Keep Natasha 0.8.0 + Yargy 0.9.0 in an isolated development reference harness.
  Committed fixtures are usable offline; ordinary Go builds/tests never invoke
  Python or download runtime data.
- Do not create `go-natasha` for the MVP.

The internal annotation contract contains original byte span, folded form,
closed token kind/shape flags, lexical role bits and only the bounded
compatible-form classes demonstrated by the accepted corpus. It does not expose
lemmas, arbitrary grammemes or normalized person/address fields.

## Rejected alternatives

- **Separate `go-natasha` now:** it would promise broad compatibility, require a
  public grammar/morphology API and independent releases without a second
  consumer.
- **Embed a mini Natasha copy:** copying upstream grammars/dictionaries would
  retain irrelevant semantics and unresolved data provenance while hiding the
  same maintenance burden inside this module.
- **Port generic Yargy/Earley:** mandatory M4/M5 forms are finite bounded
  sequences; chart parsing, relation graphs and interpretation trees add cost
  without required behavior.
- **Use a full OpenCorpora Go analyzer:** audited candidates add roughly 8.4 MiB
  embedded data or require external dictionary initialization and CC BY-SA data
  distribution. The product does not need arbitrary lemma output.
- **Use only whole-text regular expressions:** this loses the token-level
  composition and byte-span controls needed for initials and nested address/name
  resolution.

## Consequences

- M4 and M5 can implement direct, immutable and concurrent-safe Go code without
  Python, ML, external APIs or runtime downloads.
- Recall outside the versioned corpus is intentionally limited; expanding role
  tables or suffix classes requires quality evidence and source provenance.
- `go-razdel` hyphenated tokens may be split into internal logical subparts, but
  all resulting spans remain slices of the original input.
- Natasha differences such as isolated-name acceptance are documented product
  choices, not parity regressions.

## Extraction criterion for a future module

Reconsider a standalone NLP runtime only through a new OpenSpec change after all
of the following are true:

1. at least two independent consumers need the same Go-native matcher;
2. its grammar/token/morphology contract has remained stable across PERSON and
   ADDRESS implementations;
3. the reusable package can be named without implying Natasha/Yargy API parity;
4. code and data licenses, versioning, compatibility and release ownership are
   explicit; and
5. extraction does not expose llm-guard-specific entity acceptance or resolver
   policy.
