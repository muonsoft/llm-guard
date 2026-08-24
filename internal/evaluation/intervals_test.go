package evaluation_test

import (
	"testing"

	"github.com/muonsoft/llm-guard"
	"github.com/muonsoft/llm-guard/internal/evaluation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExposureByteMetrics_WhenTableDriven_ExpectExactCounts(t *testing.T) {
	t.Parallel()

	type interval = evaluation.ByteIntervalForTest
	cases := []struct {
		name            string
		gold            []interval
		pred            []interval
		wantSensitive   int
		wantCovered     int
		wantLeaked      int
		wantOvermatched int
	}{
		{
			name:            "disjoint",
			gold:            []interval{{0, 3}},
			pred:            []interval{{5, 8}},
			wantSensitive:   3,
			wantCovered:     0,
			wantLeaked:      3,
			wantOvermatched: 3,
		},
		{
			name:            "nested",
			gold:            []interval{{0, 10}, {2, 6}},
			pred:            []interval{{0, 10}},
			wantSensitive:   10,
			wantCovered:     10,
			wantLeaked:      0,
			wantOvermatched: 0,
		},
		{
			name:            "overlapping",
			gold:            []interval{{0, 5}, {3, 8}},
			pred:            []interval{{0, 8}},
			wantSensitive:   8,
			wantCovered:     8,
			wantLeaked:      0,
			wantOvermatched: 0,
		},
		{
			name:            "partial_coverage",
			gold:            []interval{{0, 10}},
			pred:            []interval{{0, 4}},
			wantSensitive:   10,
			wantCovered:     4,
			wantLeaked:      6,
			wantOvermatched: 0,
		},
		{
			name:            "overmatched_prediction",
			gold:            []interval{{0, 4}},
			pred:            []interval{{0, 8}},
			wantSensitive:   4,
			wantCovered:     4,
			wantLeaked:      0,
			wantOvermatched: 4,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			sensitive, covered, leaked, overmatched := evaluation.ExposureByteMetricsForTest(tc.gold, tc.pred)
			assert.Equal(t, tc.wantSensitive, sensitive, "sensitive")
			assert.Equal(t, tc.wantCovered, covered, "covered")
			assert.Equal(t, tc.wantLeaked, leaked, "leaked")
			assert.Equal(t, tc.wantOvermatched, overmatched, "overmatched")
		})
	}
}

func TestExposureCaseMetrics_WhenOverlappingGoldLabels_ExpectUnionSensitiveBytes(t *testing.T) {
	t.Parallel()

	rec := evaluation.SuiteRecord{
		Input: "0123456789",
		Annotations: []evaluation.SuiteAnnotation{
			{SourceLabel: "A", MappedEntity: "EMAIL", Start: 0, End: 5, Disposition: evaluation.DispositionSupported},
			{SourceLabel: "A", MappedEntity: "EMAIL", Start: 3, End: 8, Disposition: evaluation.DispositionSupported},
		},
	}
	sensitive, covered, leaked, overmatched, byLabel := evaluation.ExposureCaseMetricsForTest(rec, nil)
	assert.Equal(t, 8, sensitive)
	assert.Equal(t, 0, covered)
	assert.Equal(t, 8, leaked)
	assert.Equal(t, 0, overmatched)

	key := evaluation.LabelMetricsLookupKey("A", "EMAIL", evaluation.DispositionSupported)
	m, ok := byLabel[key]
	require.True(t, ok)
	assert.Equal(t, 8, m.SensitiveBytes)
	assert.Equal(t, 2, m.SpanCount)
}

func TestExposureCaseMetrics_WhenExtraPrediction_ExpectEntityOvermatched(t *testing.T) {
	t.Parallel()

	rec := evaluation.SuiteRecord{
		Input: "Contact a@b.co",
		Annotations: []evaluation.SuiteAnnotation{
			{SourceLabel: "EMAIL", MappedEntity: "EMAIL", Start: 8, End: 14, Disposition: evaluation.DispositionSupported},
		},
	}
	resolved := []llmguard.Finding{
		{Entity: llmguard.EntityEmail, Start: 8, End: 14},
		{Entity: llmguard.EntityPhone, Start: 0, End: 7},
	}
	sensitive, _, _, overmatched, byLabel := evaluation.ExposureCaseMetricsForTest(rec, resolved)
	assert.Equal(t, 6, sensitive)
	assert.Equal(t, 7, overmatched)

	phoneKey := evaluation.LabelMetricsLookupKey("", "PHONE", "")
	m, ok := byLabel[phoneKey]
	require.True(t, ok)
	assert.Equal(t, 7, m.Overmatched)
}

func TestMergeIntervals_WhenAdjacent_ExpectMergedCount(t *testing.T) {
	t.Parallel()

	count := evaluation.MergeIntervalsForTest([]evaluation.ByteIntervalForTest{
		{0, 3}, {3, 6},
	})
	assert.Equal(t, 6, count)
}
