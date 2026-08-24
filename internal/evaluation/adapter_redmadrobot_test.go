package evaluation_test

import (
	"strings"
	"testing"

	"github.com/muonsoft/llm-guard/internal/evaluation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedMadRobotAdapter_WhenCyrillicOffsets_ExpectByteSpans(t *testing.T) {
	t.Parallel()
	policy := redMadRobotPolicy(t)
	adapter := evaluation.RedMadRobotAdapter{Policy: policy}
	text := "Email иван@example.com"
	tokens := `["Email","иван@example.com"]`
	tags := `["O","B-EMAIL"]`
	rec, err := adapter.AdaptCSVRecord("0", text, tokens, tags)
	require.NoError(t, err)
	require.Len(t, rec.Annotations, 1)
	assert.Equal(t, evaluation.DispositionSupported, rec.Annotations[0].Disposition)
	assert.Equal(t, 6, rec.Annotations[0].Start)
	assert.Equal(t, 26, rec.Annotations[0].End)
}

func TestRedMadRobotAdapter_WhenBIORun_ExpectCollapsedSpan(t *testing.T) {
	t.Parallel()
	policy := redMadRobotPolicy(t)
	adapter := evaluation.RedMadRobotAdapter{Policy: policy}
	text := "Иван Петров"
	tokens := `["Иван","Петров"]`
	tags := `["B-FIRST_NAME","B-LAST_NAME"]`
	rec, err := adapter.AdaptCSVRecord("1", text, tokens, tags)
	require.NoError(t, err)
	var personSupported bool
	for _, ann := range rec.Annotations {
		if ann.MappedEntity == "PERSON" && ann.Disposition == evaluation.DispositionSupported {
			personSupported = true
			assert.Equal(t, 0, ann.Start)
			assert.Equal(t, 21, ann.End)
		}
	}
	assert.True(t, personSupported)
}

func TestRedMadRobotAdapter_WhenTokenMissing_ExpectIgnored(t *testing.T) {
	t.Parallel()
	policy := redMadRobotPolicy(t)
	adapter := evaluation.RedMadRobotAdapter{Policy: policy}
	text := "site example.com"
	tokens := `["WWW"]`
	tags := `["B-URL"]`
	rec, err := adapter.AdaptCSVRecord("2", text, tokens, tags)
	require.NoError(t, err)
	require.Len(t, rec.Annotations, 1)
	assert.Equal(t, evaluation.DispositionIgnored, rec.Annotations[0].Disposition)
	assert.Equal(t, "adapter_token_not_found_in_text", rec.Annotations[0].Reason)
}

func TestRedMadRobotAdapter_WhenOMS_ExpectUnsupported(t *testing.T) {
	t.Parallel()
	policy := redMadRobotPolicy(t)
	adapter := evaluation.RedMadRobotAdapter{Policy: policy}
	text := "полис 1234567890123456"
	tokens := `["полис","1234567890123456"]`
	tags := `["O","B-OMS"]`
	rec, err := adapter.AdaptCSVRecord("3", text, tokens, tags)
	require.NoError(t, err)
	var oms bool
	for _, ann := range rec.Annotations {
		if ann.SourceLabel == "OMS" {
			oms = true
			assert.Equal(t, evaluation.DispositionUnsupported, ann.Disposition)
		}
	}
	assert.True(t, oms)
}

func TestRedMadRobotAdapter_WhenPunctuationGlued_ExpectAligned(t *testing.T) {
	t.Parallel()
	policy := redMadRobotPolicy(t)
	adapter := evaluation.RedMadRobotAdapter{Policy: policy}
	text := "tel:+79991234567"
	tokens := `["tel:+79991234567"]`
	tags := `["B-PHONE"]`
	rec, err := adapter.AdaptCSVRecord("4", text, tokens, tags)
	require.NoError(t, err)
	require.NotEmpty(t, rec.Annotations)
	assert.Equal(t, 0, rec.Annotations[0].Start)
}

func TestNormalizeRedMadRobotCSV_WhenHeaderMissingNERTags_ExpectError(t *testing.T) {
	t.Parallel()
	csvData := "text,tokens\nrow1,\"[]\"\n"
	err := evaluation.NormalizeRedMadRobotCSV(strings.NewReader(csvData), redMadRobotPolicy(t), func(evaluation.SuiteRecord) error {
		return nil
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing required columns")
	assert.NotContains(t, err.Error(), "row1")
}
