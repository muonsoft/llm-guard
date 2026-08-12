# Natasha port scope for the MVP

## Outcome

Natasha is a pinned development reference, not a library to port API-for-API.
M4 PERSON and M5 ADDRESS use `github.com/muonsoft/go-razdel` as an external
tokenizer and a product-specific runtime under `internal/` in this repository.
The runtime consists of direct Go rules compiled to a bounded token matcher. A
separate `go-natasha` module is not part of the MVP.

## Pinned executable reference

The reference pair was selected from the last Natasha release before 1.0 and
verified by executing both extractors in Python 3.6.15:

| Component | Revision / artifact | Role |
|---|---|---|
| Natasha | tag `0.8.0`, commit `b603af32a598105d8d42c209a422436949a18e48`; PyPI wheel SHA-256 `43f9bf04e489a28caa172cc62234efc33ef71cdf8299b51e7f8b2979a47024d5` | `NamesExtractor`, `AddressExtractor`, grammar and reference data |
| Yargy | version `0.9.0`, commit `c67041510d88981e19548715cd6fe6744d3e41e4`; PyPI wheel SHA-256 `571511ac621b848b0d828340e27549cf83b6466acccfac6ce8807317e63b34ee` | tokenizer, morphology adapter, predicates, chart parser, relations and interpretation |
| Pymorphy2 | tag `0.8`, commit `35bdb0e879419913606c1fb8d718ded265dac24b`; wheel SHA-256 `549a1141abe01093242b9e11a1f60110ee4818f7fe57c9bb363e504339f382f5` | reference-only morphological analysis |
| Pymorphy2 dictionaries | `2.4.393442.3710985`; wheel SHA-256 `70d3e33fa28108a2dfcded787c7a5946c6ab88bb669b3afb20f8c447aadec924` | reference-only OpenCorpora-derived data |
| Python | container image `python:3.6.15`; observed image digest `sha256:f8652afaf88c25f0d22354d547d892591067aa4026a7fa9a6819df9f300af6fc` | legacy executable environment with compiler for Yargy's pinned Jellyfish dependency |
| go-razdel audit | commit `5cd53c7a1d02780285406c6c9f1635a89953c27a`; upstream Razdel submodule `668dbe191a5cfd94bebf9155e2ffa5f94ff3fe33` | normative product token contract to add in M4 |

Natasha 0.8 declares an unpinned Git dependency on Yargy. Yargy commit
`c670415...` is the 0.9.0 version bump published on the same date as Natasha
0.8.0. The PyPI Yargy 0.9.0 package is byte-equivalent to that commit under the
`yargy/` package tree, and the pair was executed successfully. This closes the
otherwise floating upstream requirement.

Sources can be obtained with:

```bash
git clone --branch 0.8.0 https://github.com/natasha/natasha.git
git clone https://github.com/natasha/yargy.git
git -C yargy checkout c67041510d88981e19548715cd6fe6744d3e41e4
git clone --branch 0.8 https://github.com/pymorphy2/pymorphy2.git
git clone https://github.com/muonsoft/go-razdel.git
git -C go-razdel checkout 5cd53c7a1d02780285406c6c9f1635a89953c27a
git -C go-razdel submodule update --init --recursive
```

## Transitive extractor graph

`NamesExtractor` follows this complete path:

```text
NamesExtractor
  -> Extractor(NAME) -> normalize_text -> Yargy Parser.findall
  -> Yargy Tokenizer
       -> RussianRule -> cached Pymorphy2 0.8 analyzer
       -> Pymorphy2 dictionaries -> OpenCorpora-derived forms
  -> name grammar
       -> first/last/maybe-first/maybe-last text dictionaries
       -> eq/length/gram/dictionary/capitalization/single predicates
       -> gender-number-case relations
  -> Earley-style chart + relation graph
  -> interpretation tree -> normalized Name fact
  -> longest non-overlapping matches via IntervalTree
```

`AddressExtractor` follows this complete path:

```text
AddressExtractor
  -> Extractor(ADDRESS, ComplexGorodPipeline) -> normalize_text
  -> the same Yargy Tokenizer and Pymorphy2/OpenCorpora path
  -> ComplexGorodPipeline for one multi-token settlement case
  -> address grammar
       -> literal/alias dictionaries, normalized forms, token grammemes,
          integer range/length, capitalization, alternatives and optionals
       -> typed address-part interpretation
  -> the same chart parser and interpretation tree
  -> Address fact; in 0.8 the top-level rule requires street + house
```

The Yargy parser holds a lock around tokenization/chart construction and relies
on a process-global cached morphology analyzer. Those implementation choices are
reference behavior only and are not copied to the concurrent Go runtime.

## Exact upstream construct inventory

Static inventory for the pinned grammars:

| Construct | Name grammar | Address grammar |
|---|---:|---:|
| `rule` / sequence | 13 | 155 |
| `or_` | 5 | 65 |
| `and_` | 8 | 11 |
| `not_` | 2 | 0 |
| `optional` | 0 | 64 |
| repeatable fact attribute | 0 | 1 |
| dictionary predicate | 4 | 13 |
| morphology `gram` predicate | 5 | 9 |
| `normalized` predicate | 0 | 69 |
| case-insensitive literal/set | 0 | 52 |
| relation `.match` | 16 | 0 |
| interpretation | 29 | 84 |
| inflection during interpretation | 22 | 0 |

The name grammar consumes `Name`, `Surn`, `Patr`, and `Abbr`, plus the
gender/number/case dimensions enforced by `gnc_relation`: gender
`masc/femn/neut/Ms-f/GNdr`, number `sing/plur/Sgtm/Pltm`, and case
`nomn/gent/datv/accs/ablt/loct/voct/Fixd`. The address grammar explicitly
consumes `ADJF`, `NOUN`, and `INT`; `normalized(...)` also depends on arbitrary
Pymorphy normal forms.

## Closed product feature matrix

`Required` means the capability must be expressible by the internal runtime,
not that the Yargy implementation is copied.

| Feature | M4 PERSON | M5 ADDRESS | Required implementation |
|---|---|---|---|
| finite token sequence and alternatives | yes | yes | bounded matcher nodes |
| predicate conjunction / exclusion | yes | yes | direct Go boolean predicates |
| optional punctuation or address parts | yes | yes | bounded optionals; no unbounded search |
| unbounded grammar recursion | no | no | out of scope |
| generic repeatable fact field | no | no | explicit maximum address-part count |
| literal and case-folded aliases | initials only | yes | project-owned immutable tables |
| capitalization and initial shape | yes | no | token shape flags |
| integer type, length, and range | no | yes | token kind plus numeric predicate |
| first/surname/patronymic role | yes | no | compact project-owned role/suffix rules |
| arbitrary OpenCorpora grammemes | no | no | out of scope |
| full gender-number-case relation graph | no | no | bounded compatible-form classes only where corpus requires them |
| lemma normalization in output | no | no | findings preserve source text and byte span |
| Yargy interpretation/fact tree | no | no | direct span and address-part capture |
| Natasha single-name semantics | no | no | intentional PERSON negative |
| standalone settlement/region/street | no | no | intentional ADDRESS negative |
| full Natasha geography/name dictionaries | no | no | excluded from production |
| `ComplexGorodPipeline` parity | no | no | explicit product alias when a corpus case requires it |

Every mandatory M4 form is a finite two- or three-role sequence, optionally with
period-separated initials. Every mandatory M5 form is a finite composition of
settlement, street-kind/name, house and optional building/apartment parts. No
mandatory example requires recursion, a chart, an interpretation tree, or
unbounded repetition; therefore there is no blocking construct for a bounded
matcher.

## Product token annotation contract

M4 introduces an internal value per `go-razdel` token with:

- the original text and unchanged UTF-8 byte `[start,end)` span;
- Unicode case-folded text;
- a closed kind (`word`, `integer`, `punctuation`, `other`);
- shape flags (`capitalized`, `single-letter initial`, `hyphenated`);
- closed lexical role bits needed by rules (`first`, `surname`, `patronymic`,
  address-kind aliases); and
- a small compatible-form class used only by the accepted declined-name corpus.

Annotation is a pure operation. Tables are project-owned, immutable after Guard
construction, and never expose a public morphology API. PERSON uses conservative
capitalization, compact authored role tables and deterministic Russian
name/patronymic/surname suffix classes. ADDRESS uses authored aliases such as
`ул./улица`, `пр-т/проспект`, `д./дом`, `корп.`, `стр.`, and `кв.`. No lemma or
normalized fact is needed because the detector returns the source span.

Adding a copied or generated lexicon later requires a new provenance/license
entry and corpus evidence. It is not an implicit extension point for importing
Natasha or OpenCorpora data.

## Token-contract differences

Yargy spans are Python codepoint offsets; `go-razdel` spans are UTF-8 byte
offsets. For `Иванов И.И.`, Yargy emits codepoint span `(0,6)` for `Иванов`, while
`go-razdel` emits byte span `[0,12)`. Hyphen handling also differs: Yargy emits
`Анна`, `-`, `Мария` and `Санкт`, `-`, `Петербург`; `go-razdel` emits
`Анна-Мария` and `Санкт-Петербург` as one token. Initial punctuation and compact
`ул.Тверская,д.1` remain separate tokens in both implementations.

The product adapter treats `go-razdel` as normative and may split a hyphenated
token into bounded logical subparts while retaining offsets into the original
token. The reference fixture stores both Python codepoint and converted UTF-8
byte spans and checks `input_bytes[start:end] == matched_text`.

## Quality boundary

Cases are synthetic and classified as `positive`, `negative`, or `ambiguous`.
Metrics are reported independently for PERSON and ADDRESS:

- exact-span true positives, false positives and false negatives;
- precision, recall and exact-span rate when a denominator is non-zero;
- reference-only matches classified as `regression`, `intentional_difference`,
  or `unsupported_out_of_scope`.

M4 optimizes precision and rejects isolated names/surnames. M5 accepts at least
street + house (plus optional stronger parts) and rejects isolated geographic or
street parts. Natasha output is diagnostic evidence, not product acceptance.
