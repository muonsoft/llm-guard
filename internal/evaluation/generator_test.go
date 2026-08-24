package evaluation_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/muonsoft/llm-guard/internal/evaluation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommittedSmokeMatchesGeneratorSeed(t *testing.T) {
	t.Parallel()
	path := filepath.Join(repoRoot(t), "testdata", "evaluation", "generated", "smoke.jsonl")
	committed, err := evaluation.LoadSuite(path)
	require.NoError(t, err)
	generated := evaluation.GenerateSmokeSuite(evaluation.GeneratorSeed)
	require.Len(t, committed.Records, len(generated))
	for i := range generated {
		assert.Equal(t, generated[i].SourceRecordID, committed.Records[i].SourceRecordID)
		assert.Equal(t, generated[i].InputSHA256, committed.Records[i].InputSHA256)
		assert.Equal(t, len(generated[i].Annotations), len(committed.Records[i].Annotations))
	}
}

// Regenerate committed smoke with:
//
//	REGEN_SMOKE=1 go test ./internal/evaluation -run TestRegenerateSmokeJSONL -count=1
func TestRegenerateSmokeJSONL(t *testing.T) {
	if os.Getenv("REGEN_SMOKE") != "1" {
		t.Skip("set REGEN_SMOKE=1 to rewrite testdata/evaluation/generated/smoke.jsonl")
	}
	records := evaluation.GenerateSmokeSuite(evaluation.GeneratorSeed)
	path := filepath.Join(repoRoot(t), "testdata", "evaluation", "generated", "smoke.jsonl")
	require.NoError(t, evaluation.WriteSuiteJSONL(path, records))
}

func TestWriteSmokeJSONL_WhenTempDir_ExpectRoundTrip(t *testing.T) {
	t.Parallel()
	records := evaluation.GenerateSmokeSuite(evaluation.GeneratorSeed)
	path := filepath.Join(t.TempDir(), "smoke.jsonl")
	require.NoError(t, evaluation.WriteSuiteJSONL(path, records))
	loaded, err := evaluation.LoadSuite(path)
	require.NoError(t, err)
	require.Len(t, loaded.Records, len(records))
}

func TestGenerateSmokeSuite_WhenFixedSeed_ExpectDeterministic(t *testing.T) {
	t.Parallel()
	a := evaluation.GenerateSmokeSuite(evaluation.GeneratorSeed)
	b := evaluation.GenerateSmokeSuite(evaluation.GeneratorSeed)
	require.Len(t, a, len(b))
	for i := range a {
		assert.Equal(t, a[i].SourceRecordID, b[i].SourceRecordID)
		assert.Equal(t, a[i].InputSHA256, b[i].InputSHA256)
	}
}

func TestGenerateSmokeSuite_WhenMarshaled_ExpectStableJSONL(t *testing.T) {
	t.Parallel()
	a := evaluation.GenerateSmokeSuite(evaluation.GeneratorSeed)
	b := evaluation.GenerateSmokeSuite(evaluation.GeneratorSeed)
	lineA, err := json.Marshal(a[0])
	require.NoError(t, err)
	lineB, err := json.Marshal(b[0])
	require.NoError(t, err)
	assert.JSONEq(t, string(lineA), string(lineB))
}

func TestGenerateSmokeSuite_WhenValidINN_ExpectDetectable(t *testing.T) {
	t.Parallel()
	guard, err := evaluation.NewMVPGuard()
	require.NoError(t, err)
	for _, rec := range evaluation.GenerateSmokeSuite(evaluation.GeneratorSeed) {
		if rec.SourceRecordID != "inn-valid-10" && rec.SourceRecordID != "inn-valid-12" {
			continue
		}
		findings, err := guard.Detect(context.Background(), rec.Input)
		require.NoError(t, err)
		assert.NotEmpty(t, findings)
	}
}
