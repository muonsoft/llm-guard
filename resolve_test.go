package llmguard_test

import (
	"testing"

	muerrors "github.com/muonsoft/errors"
	"github.com/muonsoft/llm-guard"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolve_WhenExactDuplicates_ExpectSingleFinding(t *testing.T) {
	t.Parallel()

	text := "contact a@b.co"
	finding := llmguard.Finding{
		Entity: llmguard.EntityEmail, Start: 8, End: 14, Confidence: 0.9, Detector: "email",
	}
	resolved, err := llmguard.Resolve(text, []llmguard.Finding{finding, finding, finding})
	require.NoError(t, err)
	require.Len(t, resolved, 1)
}

func TestResolve_WhenNestedEmailAndCustom_ExpectEmailWins(t *testing.T) {
	t.Parallel()

	text := "mail a@b.co"
	email := llmguard.Finding{
		Entity: llmguard.EntityEmail, Start: 5, End: 11, Confidence: 0.9, Detector: "email",
	}
	custom := llmguard.Finding{
		Entity: llmguard.EntityType("CUSTOM"), Start: 7, End: 10, Confidence: 1, Detector: "custom",
	}
	resolved, err := llmguard.Resolve(text, []llmguard.Finding{custom, email})
	require.NoError(t, err)
	require.Len(t, resolved, 1)
	assert.Equal(t, llmguard.EntityEmail, resolved[0].Entity)
}

func TestResolve_WhenEqualPriorityOverlap_ExpectLongerSpanWins(t *testing.T) {
	t.Parallel()

	text := "abcdefghij"
	a := llmguard.Finding{
		Entity: llmguard.EntityType("A"), Start: 0, End: 6, Confidence: 0.5, Detector: "d1",
	}
	b := llmguard.Finding{
		Entity: llmguard.EntityType("B"), Start: 2, End: 8, Confidence: 0.5, Detector: "d2",
	}
	resolved, err := llmguard.Resolve(text, []llmguard.Finding{a, b})
	require.NoError(t, err)
	require.Len(t, resolved, 1)
	assert.Equal(t, 0, resolved[0].Start)
	assert.Equal(t, 6, resolved[0].End)
}

func TestResolve_WhenAdjacentSpans_ExpectBothKept(t *testing.T) {
	t.Parallel()

	text := "abcdef"
	first := llmguard.Finding{
		Entity: llmguard.EntityEmail, Start: 0, End: 3, Confidence: 1, Detector: "email",
	}
	second := llmguard.Finding{
		Entity: llmguard.EntityPhone, Start: 3, End: 6, Confidence: 1, Detector: "phone",
	}
	resolved, err := llmguard.Resolve(text, []llmguard.Finding{first, second})
	require.NoError(t, err)
	require.Len(t, resolved, 2)
}

func TestResolve_WhenPermutedInput_ExpectStableOutput(t *testing.T) {
	t.Parallel()

	text := "x a@b.co y c@d.org z"
	f1 := llmguard.Finding{
		Entity: llmguard.EntityEmail, Start: 2, End: 8, Confidence: 0.9, Detector: "email",
	}
	f2 := llmguard.Finding{
		Entity: llmguard.EntityEmail, Start: 11, End: 18, Confidence: 0.9, Detector: "email",
	}
	perms := [][]llmguard.Finding{
		{f1, f2},
		{f2, f1},
	}
	var baseline []llmguard.Finding
	for i, input := range perms {
		resolved, err := llmguard.Resolve(text, input)
		require.NoError(t, err)
		if i == 0 {
			baseline = resolved
		} else {
			assert.Equal(t, baseline, resolved)
		}
	}
}

func TestResolve_WhenInvalidUTF8Text_ExpectInvalidText(t *testing.T) {
	t.Parallel()

	_, err := llmguard.Resolve(string([]byte{0xff}), nil)
	require.Error(t, err)
	assert.ErrorIs(t, err, llmguard.ErrInvalidText)
}

func TestResolve_WhenInvalidSpan_ExpectInvalidFinding(t *testing.T) {
	t.Parallel()

	text := "café"
	_, err := llmguard.Resolve(text, []llmguard.Finding{{
		Entity: llmguard.EntityEmail, Start: 2, End: 4, Confidence: 1, Detector: "email",
	}})
	require.Error(t, err)
	assert.ErrorIs(t, err, llmguard.ErrInvalidFinding)
}

func TestResolve_WhenEmptyDetector_ExpectInvalidFinding(t *testing.T) {
	t.Parallel()

	_, err := llmguard.Resolve("abc", []llmguard.Finding{{
		Entity: llmguard.EntityEmail, Start: 0, End: 1, Confidence: 1,
	}})
	require.Error(t, err)
	assert.ErrorIs(t, err, llmguard.ErrInvalidFinding)

	inv, ok := muerrors.As[*llmguard.InvalidFindingError](err)
	require.True(t, ok)
	assert.Equal(t, "detector", inv.Field)
}

func TestResolve_WhenEmptyInput_ExpectEmptyResult(t *testing.T) {
	t.Parallel()

	resolved, err := llmguard.Resolve("text", nil)
	require.NoError(t, err)
	assert.Nil(t, resolved)
}

func TestResolve_WhenTwoCustomEntitiesConflict_ExpectDeterministicWinner(t *testing.T) {
	t.Parallel()

	text := "0123456789"
	a := llmguard.Finding{
		Entity: llmguard.EntityType("ALPHA"), Start: 0, End: 5, Confidence: 0.7, Detector: "a",
	}
	b := llmguard.Finding{
		Entity: llmguard.EntityType("BETA"), Start: 0, End: 5, Confidence: 0.7, Detector: "b",
	}
	first, err := llmguard.Resolve(text, []llmguard.Finding{a, b})
	require.NoError(t, err)
	second, err := llmguard.Resolve(text, []llmguard.Finding{b, a})
	require.NoError(t, err)
	assert.Equal(t, first, second)
}

func TestResolve_WhenEqualPriorityTieBreak_ExpectHigherConfidenceWins(t *testing.T) {
	t.Parallel()

	text := "0123456789"
	low := llmguard.Finding{
		Entity: llmguard.EntityType("X"), Start: 0, End: 5, Confidence: 0.4, Detector: "d",
	}
	high := llmguard.Finding{
		Entity: llmguard.EntityType("Y"), Start: 1, End: 6, Confidence: 0.9, Detector: "d",
	}
	resolved, err := llmguard.Resolve(text, []llmguard.Finding{low, high})
	require.NoError(t, err)
	require.Len(t, resolved, 1)
	assert.InDelta(t, 0.9, resolved[0].Confidence, 0)
}

func TestResolve_WhenErrorMessage_ExpectNoInputSubstring(t *testing.T) {
	t.Parallel()

	sensitive := "secret@example.com"
	err := func() error {
		_, resolveErr := llmguard.Resolve(sensitive, []llmguard.Finding{{
			Entity: llmguard.EntityEmail, Start: 0, End: 5, Confidence: 2, Detector: "email",
		}})
		return resolveErr
	}()
	require.Error(t, err)
	assert.NotContains(t, err.Error(), sensitive)
}
