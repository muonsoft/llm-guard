# Natasha reference and NLP source license inventory

This inventory separates development reference material from sources permitted
in the production Go dependency/data graph.

| Source | Audited revision | License / provenance | Footprint and runtime properties | Decision and distribution plan |
|---|---|---|---|---|
| `github.com/muonsoft/go-razdel` | `5cd53c7a1d02780285406c6c9f1635a89953c27a` (`v0.0.0-20260425122647-5cd53c7a1d02`) | MIT; code port references MIT Razdel revision `668dbe191a5cfd94bebf9155e2ffa5f94ff3fe33` | no external dictionary; pure Go; byte offsets | permitted external production module in M4; retain license notice through normal module distribution |
| llm-guard bounded matcher, aliases, suffix rules and synthetic fixtures | repository revision | project-authored, repository MIT | small immutable tables; no runtime download | permitted production source; record provenance for every future imported list |
| Natasha | `0.8.0`, `b603af32...` | MIT code; bundled first/last-name lists have no source/provenance statement beyond repository packaging | name data is about 3.9 MB unpacked; Python reference only | code may run in isolated development harness; do not copy its dictionaries or grammar source into production |
| Yargy | `0.9.0`, `c670415...` | MIT | generic chart parser; process-level morphology cache; parser lock | development reference only; no Go runtime port |
| Pymorphy2 | `0.8`, `35bdb0e...` | MIT code | analyzer requires external compiled dictionaries | development reference only |
| `pymorphy2-dicts` / OpenCorpora data | `2.4.393442.3710985` | dictionary data CC BY-SA 3.0, OpenCorpora contributors | reference wheel 7,100,976 bytes | development reference only; not linked, embedded, generated, or redistributed in the Go module |
| `github.com/jus1d/gomorphy` | `eeb495482f9e71b1999f577b63c37e2e9d941481` | MIT code; embedded OpenCorpora v0.92 data declared CC BY-SA 4.0 | about 8.4 MiB embedded; immutable and concurrent-safe after initialization | rejected for MVP: footprint and share-alike data distribution are unnecessary for required corpus |
| `github.com/therox/gomorphy` | `1676e58fcf4a720fca211ff1de174032823d9900` | MIT code; user-supplied OpenCorpora dictionary remains CC BY-SA 4.0 | external compiled dictionary; global initialization; full source XML is roughly 500 MB according to project docs | rejected as dependency; no runtime download/global analyzer |
| `github.com/pahanini/go-opencorpora-tools` | `af8bc09d16cfc45d9403858ba022accd979b9149` | MIT code; external OpenCorpora data retains its own license | dictionary build/load tooling rather than a small detector dependency | rejected for MVP |
| `github.com/vseledkin/morph` | candidate observed during audit | GPL-3.0 | full morphology scope | rejected on license and scope |

The only third-party production NLP source approved by R0 is `go-razdel`.
Project-authored rule tables must not be populated by copying reference
dictionaries. A future imported dataset is blocked until its exact revision,
provenance, license, attribution text, generated-artifact license, compressed and
runtime footprint, initialization behavior, and concurrent-use model are added
to this inventory and reviewed.
