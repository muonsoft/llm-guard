package evaluation_test

import (
	"path/filepath"
	"testing"

	"github.com/muonsoft/llm-guard"
	"github.com/muonsoft/llm-guard/internal/evaluation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyMapping_WhenDirectEmail_ExpectSupported(t *testing.T) {
	t.Parallel()
	policy := evaluation.MappingPolicy{
		Direct: map[string]string{"EMAIL": "EMAIL"},
	}
	anns := evaluation.ApplyMapping(policy, []evaluation.SourceLabelSpan{{
		Label: "EMAIL", Start: 8, End: 14,
	}}, "Contact a@b.co")
	require.Len(t, anns, 1)
	assert.Equal(t, evaluation.DispositionSupported, anns[0].Disposition)
	assert.Equal(t, "EMAIL", anns[0].MappedEntity)
}

func TestApplyMapping_WhenCreditCard_ExpectBankCard(t *testing.T) {
	t.Parallel()
	policy := evaluation.MappingPolicy{
		Direct: map[string]string{"CREDIT_CARD": "BANK_CARD"},
	}
	anns := evaluation.ApplyMapping(policy, []evaluation.SourceLabelSpan{{
		Label: "CREDIT_CARD", Start: 0, End: 16,
	}}, "4111111111111111")
	require.Len(t, anns, 1)
	assert.Equal(t, "BANK_CARD", anns[0].MappedEntity)
}

func TestApplyMapping_WhenPersonFirstLast_ExpectSupportedComposition(t *testing.T) {
	t.Parallel()
	policy := redMadRobotPolicy(t)
	text := "Иван Петров"
	anns := evaluation.ApplyMapping(policy, []evaluation.SourceLabelSpan{
		{Label: "FIRST_NAME", Start: 0, End: 8},
		{Label: "LAST_NAME", Start: 9, End: 21},
	}, text)
	var supportedPerson int
	for _, ann := range anns {
		if ann.MappedEntity == string(llmguard.EntityPerson) && ann.Disposition == evaluation.DispositionSupported {
			supportedPerson++
		}
	}
	assert.Equal(t, 1, supportedPerson)
}

func TestApplyMapping_WhenPersonSingle_ExpectUnsupportedOnly(t *testing.T) {
	t.Parallel()
	policy := redMadRobotPolicy(t)
	anns := evaluation.ApplyMapping(policy, []evaluation.SourceLabelSpan{{
		Label: "LAST_NAME", Start: 0, End: 12,
	}}, "Петров")
	for _, ann := range anns {
		assert.Equal(t, evaluation.DispositionUnsupported, ann.Disposition)
		assert.Empty(t, ann.MappedEntity)
	}
}

func TestApplyMapping_WhenAddressStreetHouse_ExpectSupported(t *testing.T) {
	t.Parallel()
	policy := redMadRobotPolicy(t)
	text := "ул. Ленина, 10"
	anns := evaluation.ApplyMapping(policy, []evaluation.SourceLabelSpan{
		{Label: "STREET", Start: 0, End: 18},
		{Label: "HOUSE", Start: 20, End: 22},
	}, text)
	var supportedAddress int
	for _, ann := range anns {
		if ann.MappedEntity == string(llmguard.EntityAddress) {
			supportedAddress++
		}
	}
	assert.Equal(t, 1, supportedAddress)
}

func TestApplyMapping_WhenOMS_ExpectUnsupportedExposure(t *testing.T) {
	t.Parallel()
	policy := redMadRobotPolicy(t)
	anns := evaluation.ApplyMapping(policy, []evaluation.SourceLabelSpan{{
		Label: "OMS", Start: 0, End: 16,
	}}, "полис ОМС 123456")
	require.Len(t, anns, 1)
	assert.Equal(t, evaluation.DispositionUnsupported, anns[0].Disposition)
	assert.Empty(t, anns[0].MappedEntity)
	assert.Equal(t, "OMS", anns[0].SourceLabel)
}

func TestApplyMapping_WhenFirstMiddleWithoutLast_ExpectNoSupportedPerson(t *testing.T) {
	t.Parallel()
	policy := redMadRobotPolicy(t)
	text := "Иван Сергеевич"
	anns := evaluation.ApplyMapping(policy, []evaluation.SourceLabelSpan{
		{Label: "FIRST_NAME", Start: 0, End: 8},
		{Label: "MIDDLE_NAME", Start: 9, End: 27},
	}, text)
	for _, ann := range anns {
		assert.NotEqual(t, evaluation.DispositionSupported, ann.Disposition, "unexpected supported: %+v", ann)
		if ann.MappedEntity == string(llmguard.EntityPerson) {
			t.Fatalf("unexpected supported PERSON: %+v", ann)
		}
	}
}

func TestApplyMapping_WhenCityWithoutStreetHouse_ExpectNoSupportedAddress(t *testing.T) {
	t.Parallel()
	policy := redMadRobotPolicy(t)
	text := "Москва"
	anns := evaluation.ApplyMapping(policy, []evaluation.SourceLabelSpan{{
		Label: "CITY", Start: 0, End: len(text),
	}}, text)
	for _, ann := range anns {
		if ann.MappedEntity == string(llmguard.EntityAddress) {
			t.Fatalf("unexpected supported ADDRESS: %+v", ann)
		}
	}
}

func TestApplyMapping_WhenDisjointPassportParts_ExpectSeparateNotMerged(t *testing.T) {
	t.Parallel()
	policy := redMadRobotPolicy(t)
	text := "серия 45 номер 08"
	anns := evaluation.ApplyMapping(policy, []evaluation.SourceLabelSpan{
		{Label: "PASSPORT", Start: 6, End: 8},
		{Label: "PASSPORT", Start: 15, End: 17},
	}, text)
	var supported []evaluation.SuiteAnnotation
	for _, ann := range anns {
		if ann.MappedEntity == string(llmguard.EntityPassport) && ann.Disposition == evaluation.DispositionSupported {
			supported = append(supported, ann)
		}
	}
	require.Len(t, supported, 2, "expected two separate supported PASSPORT spans, not one merged span")
	for _, ann := range supported {
		assert.Less(t, ann.End-ann.Start, 5, "supported span must not merge disjoint parts")
	}
	assert.NotEqual(t, 6, supported[0].Start+supported[0].End) // sanity: not one span covering both
}

func redMadRobotPolicy(t *testing.T) evaluation.MappingPolicy {
	t.Helper()
	path := filepath.Join(repoRoot(t), "testdata", "evaluation", "mappings", "redmadrobot-v1.json")
	policy, err := evaluation.LoadMappingPolicy(path)
	require.NoError(t, err)
	return policy
}
