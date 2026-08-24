package evaluation

import (
	"testing"

	"github.com/muonsoft/llm-guard"
)

func TestContractFailureDiagnostics_WhenSameEntityMultipleSpans_ExpectStableSort(t *testing.T) {
	t.Parallel()
	ps1, pe1 := 8, 14
	ps2, pe2 := 15, 21
	rec := SuiteRecord{
		SourceRecordID: "r1",
		InputSHA256:    "abc",
	}
	gold := map[llmguard.EntityType][]spanKey{}
	predicted := map[llmguard.EntityType][]spanKey{
		llmguard.EntityEmail: {
			{entity: llmguard.EntityEmail, start: ps2, end: pe2},
			{entity: llmguard.EntityEmail, start: ps1, end: pe1},
		},
	}
	scope := map[llmguard.EntityType]struct{}{llmguard.EntityEmail: {}}

	failures := contractFailureDiagnostics(rec, gold, predicted, scope)
	if len(failures) != 2 {
		t.Fatalf("failures len = %d, want 2", len(failures))
	}
	if failures[0].PredictedStart == nil || failures[1].PredictedStart == nil {
		t.Fatal("expected predicted offsets on FP diagnostics")
	}
	if *failures[0].PredictedStart != ps1 || *failures[1].PredictedStart != ps2 {
		t.Fatalf("sort order = [%d,%d], want [%d,%d]",
			*failures[0].PredictedStart, *failures[1].PredictedStart, ps1, ps2)
	}
}
