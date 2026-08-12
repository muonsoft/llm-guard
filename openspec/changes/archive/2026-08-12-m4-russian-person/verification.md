## Verification Report: m4-russian-person

### Summary

| Dimension | Status |
|---|---|
| Completeness | 31/31 tasks complete; 7/7 requirements implemented and synced |
| Correctness | 18/18 scenarios covered; full, focused, race and corpus checks pass |
| Coherence | R0 pure-Go boundary and M4 conservative design followed |

### Completeness

- `NewPersonDetector` is an immutable value detector with stable name `person`,
  exact UTF-8 byte findings, context checks and no test-only public API.
- `internal/nlp` contains the pinned go-razdel adapter, closed token kinds, shape
  annotations, project-authored role/form tables and finite PERSON sequences.
- Mandatory direct, reversed, patronymic, declined and initials forms are covered
  alongside isolated-name, lowercase, street, product/common-word, adjective,
  embedded-token and spaced-initial negatives.
- The 25-case product corpus validates schema/linkage and exact span-set metrics:
  TP=13, FP=0, FN=0; four R0 reference-only matches are explicitly classified.
- Public resolver, Mask/Restore, concurrency, cancellation and literal-restore
  behavior are covered without PERSON-specific pipeline branches.
- Delta requirements are synced exactly into the canonical Russian PERSON spec;
  all seven main specs pass strict validation.

### Correctness mapping

| Requirement | Implementation evidence | Scenario evidence |
|---|---|---|
| Immutable RU PERSON detector | `person.go`, `internal/nlp/context.go` | registration, cancellation and concurrency tests in `person_test.go` and `internal/nlp/context_test.go` |
| Full forms and initials | `internal/nlp/person.go`, `internal/nlp/names.go`, `internal/nlp/form.go` | mandatory and declined cases in `person_test.go`, `internal/nlp/person_test.go` |
| Exact UTF-8 spans | `internal/nlp/token.go` | Unicode, multiple-person and token-kind byte-slice tests |
| Conservative acceptance | finite role composition and street/boundary checks in `internal/nlp/person.go` | negative detector tests and corpus cases 009–012, 015–020, 022–025 |
| Pure-Go NLP runtime | direct pinned `github.com/muonsoft/go-razdel` module | `go list -deps`, `go mod graph`, offline build/test evidence |
| Versioned quality boundary | `testdata/person/cases.jsonl`, `person_corpus_test.go` | exact metrics plus pinned R0 differential verification |
| Common Mask/Restore pipeline | unchanged Guard/Resolve/Mask/Restore core | PERSON round trip and overlap regression in `person_test.go`, runnable example |

### Coherence

- Runtime uses only pure Go and the audited go-razdel pseudo-version; Natasha and
  Yargy remain development-only reference fixtures.
- Authored tables and bounded suffix rules preserve the R0 provenance decision;
  no upstream dictionary or morphology data entered the production graph.
- Matching requires explicit first/surname roles or surname plus two adjacent
  initials; arbitrary capitalized pairs and isolated roles remain outside scope.
- Restore remains literal byte-for-byte and does not promise case agreement after
  an LLM moves a placeholder.

### Issues by priority

#### CRITICAL

- None.

#### WARNING

- None.

#### SUGGESTION

- None.

### Verification commands

- `go test ./... -run 'Test(Person|Name|Morph|Token)' -count=1` — PASS.
- `go test -race ./... -run 'TestPerson' -count=1` — PASS.
- `go test ./... -run 'TestPersonCorpus_WhenVersionedCases_ExpectZeroMandatoryErrors' -v -count=1` — PASS; TP=13, FP=0, FN=0, four intentional R0 differentials.
- `go test ./... -count=1` — PASS.
- `go vet ./...` — PASS.
- `go test -race ./... -count=1` — PASS.
- offline Natasha fixture verification — PASS.
- `go list -deps ./...`, `go mod graph`, `go mod tidy -diff` — PASS.
- `openspec validate m4-russian-person --strict --no-interactive` — PASS.
- `openspec validate --specs --strict --no-interactive` after sync — PASS (7/7 specs).
- Delta/main semantic comparison after replacing the operation header — exact.
- `git diff --check` — PASS.

### Orchestration evidence

- Variant C: one Composer 2.5 primary job, one consolidated review correction and
  one verification-evidence correction in the same session.
- Herdr transport recovered from the external approval-quota interruption; both
  resumed prompts settled and bounded final `agent get/read` calls succeeded.
- The initial full review found five defect groups; post-correction API/test audit
  found two evidence gaps, now resolved with no remaining blocker.

### Final assessment

Implementation, all 18 scenarios, synced canonical spec and design coherence are
green. Ready for archive.
