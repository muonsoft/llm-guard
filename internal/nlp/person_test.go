package nlp_test

import (
	"context"
	"testing"

	"github.com/muonsoft/llm-guard/internal/nlp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectPersonSpans_MandatoryR0Positives_ExpectExactSpans(t *testing.T) {
	t.Parallel()

	cases := []struct {
		input string
		want  string
	}{
		{"Иван Петров", "Иван Петров"},
		{"Петров Иван", "Петров Иван"},
		{"Иван Сергеевич Петров", "Иван Сергеевич Петров"},
		{"Петров Иван Сергеевич", "Петров Иван Сергеевич"},
		{"Петров И. С.", "Петров И. С."},
		{"И. С. Петров", "И. С. Петров"},
		{"Ивану Сергеевичу Петрову", "Ивану Сергеевичу Петрову"},
		{"Иваном Сергеевичем Петровым", "Иваном Сергеевичем Петровым"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.want, func(t *testing.T) {
			t.Parallel()
			spans, err := nlp.DetectPersonSpans(context.Background(), tc.input)
			require.NoError(t, err)
			require.Len(t, spans, 1)
			assert.Equal(t, tc.want, tc.input[spans[0].Start:spans[0].End])
		})
	}
}

func TestDetectPersonSpans_MandatoryR0Negatives_ExpectNoMatch(t *testing.T) {
	t.Parallel()

	for _, input := range []string{"Иван", "Петров", "директор Иван", "улица Гагарина"} {
		input := input
		t.Run(input, func(t *testing.T) {
			t.Parallel()
			spans, err := nlp.DetectPersonSpans(context.Background(), input)
			require.NoError(t, err)
			assert.Empty(t, spans)
		})
	}
}

func TestDetectPersonSpans_WhenAdjectiveEndings_ExpectNoMatch(t *testing.T) {
	t.Parallel()

	for _, input := range []string{"Иван Новый", "Сергей Синий", "Мария Новая"} {
		input := input
		t.Run(input, func(t *testing.T) {
			t.Parallel()
			spans, err := nlp.DetectPersonSpans(context.Background(), input)
			require.NoError(t, err)
			assert.Empty(t, spans)
		})
	}
}

func TestTokenize_WhenInitials_ExpectWordKindWithInitialFlag(t *testing.T) {
	t.Parallel()

	text := "Петров И. С."
	tokens := nlp.Tokenize(text)
	require.GreaterOrEqual(t, len(tokens), 3)
	assert.Equal(t, nlp.KindWord, tokens[0].Kind)
	assert.False(t, tokens[0].Initial)
	assert.True(t, tokens[0].Capitalized)
	assert.Equal(t, "Петров", tokens[0].Text)
	assert.Equal(t, nlp.KindWord, tokens[1].Kind)
	assert.True(t, tokens[1].Initial)
	assert.Equal(t, "И.", tokens[1].Text)
	assert.Equal(t, nlp.KindWord, tokens[2].Kind)
	assert.True(t, tokens[2].Initial)
	assert.Equal(t, "С.", tokens[2].Text)
}

func TestTokenize_WhenSpacedInitialDots_ExpectSeparateTokens(t *testing.T) {
	t.Parallel()

	text := "Петров И . С ."
	tokens := nlp.Tokenize(text)
	for _, tok := range tokens {
		if tok.Text == "И" || tok.Text == "С" {
			assert.False(t, tok.Initial, "spaced letter must not merge into initial token")
		}
	}

	spans, err := nlp.DetectPersonSpans(context.Background(), text)
	require.NoError(t, err)
	assert.Empty(t, spans)
}
