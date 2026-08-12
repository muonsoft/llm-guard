package nlp_test

import (
	"context"
	"testing"

	"github.com/muonsoft/llm-guard/internal/nlp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTokenize_WhenRazdelInvariant_ExpectSliceEqualsText(t *testing.T) {
	t.Parallel()

	text := "Иван Петров"
	tokens := nlp.Tokenize(text)
	require.NotEmpty(t, tokens)
	for _, tok := range tokens {
		assert.Equal(t, text[tok.Start:tok.End], tok.Text)
	}
}

func TestTokenize_WhenClosedKinds_ExpectWordPunctuationIntegerOther(t *testing.T) {
	t.Parallel()

	text := "Иван, 42 ★"
	tokens := nlp.Tokenize(text)
	require.Len(t, tokens, 4)

	assert.Equal(t, nlp.KindWord, tokens[0].Kind)
	assert.Equal(t, "Иван", tokens[0].Text)
	assert.Equal(t, text[0:8], text[tokens[0].Start:tokens[0].End])

	assert.Equal(t, nlp.KindPunctuation, tokens[1].Kind)
	assert.Equal(t, ",", tokens[1].Text)
	assert.Equal(t, text[8:9], text[tokens[1].Start:tokens[1].End])

	assert.Equal(t, nlp.KindInteger, tokens[2].Kind)
	assert.Equal(t, "42", tokens[2].Text)
	assert.Equal(t, text[10:12], text[tokens[2].Start:tokens[2].End])

	assert.Equal(t, nlp.KindOther, tokens[3].Kind)
	assert.Equal(t, "★", tokens[3].Text)
	assert.Equal(t, text[13:16], text[tokens[3].Start:tokens[3].End])
}

func TestDetectPersonSpans_WhenUnicodeEmbedded_ExpectInnerSpanOnly(t *testing.T) {
	t.Parallel()

	text := "«Иван Петров» сообщил"
	spans, err := nlp.DetectPersonSpans(context.Background(), text)
	require.NoError(t, err)
	require.Len(t, spans, 1)
	assert.Equal(t, "Иван Петров", text[spans[0].Start:spans[0].End])
}

func TestDetectPersonSpans_WhenLowercase_ExpectNoMatch(t *testing.T) {
	t.Parallel()

	spans, err := nlp.DetectPersonSpans(context.Background(), "иван петров")
	require.NoError(t, err)
	assert.Empty(t, spans)
}

func TestDetectPersonSpans_WhenEmbeddedAlphanumeric_ExpectNoMatch(t *testing.T) {
	t.Parallel()

	spans, err := nlp.DetectPersonSpans(context.Background(), "xxxИван Петров")
	require.NoError(t, err)
	assert.Empty(t, spans)
}
