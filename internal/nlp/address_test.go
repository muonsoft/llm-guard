package nlp_test

import (
	"context"
	"testing"

	"github.com/muonsoft/llm-guard/internal/nlp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectAddressSpans_MandatoryR0Positives_ExpectExactSpans(t *testing.T) {
	t.Parallel()

	cases := []struct {
		input string
		want  string
	}{
		{"г. Москва, ул. Тверская, д. 18", "г. Москва, ул. Тверская, д. 18"},
		{"Москва, Тверская улица, 18", "Москва, Тверская улица, 18"},
		{"ул. Ленина, дом 15, кв. 27", "ул. Ленина, дом 15, кв. 27"},
		{"проспект Мира, 101", "проспект Мира, 101"},
		{"ул.Тверская,д.1", "ул.Тверская,д.1"},
		{"ул. Ленина, 10", "ул. Ленина, 10"},
		{"ул. Ленина, д. 15, корп. 2, стр. 1, кв. 27", "ул. Ленина, д. 15, корп. 2, стр. 1, кв. 27"},
		{"ул. Академика Сахарова, 10", "ул. Академика Сахарова, 10"},
		{"проспект Мира, д. 101", "проспект Мира, д. 101"},
		{"ул. Ленина, д. 10/2", "ул. Ленина, д. 10/2"},
		{"ул. Ленина, д. 10-2", "ул. Ленина, д. 10-2"},
		{"ул. Ленина, д. 10А", "ул. Ленина, д. 10А"},
		{"Тверская ул., 18", "Тверская ул., 18"},
		{"Мира пр-т, 18", "Мира пр-т, 18"},
		{"Садовый пер., 18", "Садовый пер., 18"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.want, func(t *testing.T) {
			t.Parallel()
			spans, err := nlp.DetectAddressSpans(context.Background(), tc.input)
			require.NoError(t, err)
			require.Len(t, spans, 1, "tokens: %+v", nlp.Tokenize(tc.input))
			assert.Equal(t, tc.want, tc.input[spans[0].Start:spans[0].End])
		})
	}
}

func TestDetectAddressSpans_WhenContextCancelledBeforeTokenize_ExpectError(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	spans, err := nlp.DetectAddressSpans(ctx, "ул. Ленина, 10")
	require.Error(t, err)
	assert.Nil(t, spans)
}

func TestDetectAddressSpans_MandatoryR0Negatives_ExpectNoMatch(t *testing.T) {
	t.Parallel()

	for _, input := range []string{
		"Москва",
		"Ленинградская область",
		"улица Ленина",
		"Москва, Тверская улица",
		"Тверская",
		"ул. Ленина,,,,10",
		"ул. Ленина, д. 15,,,,корп. 2",
		"ул. Ленина, д. 10АБ",
		"ул. Ленина, д. 10/2/3",
		"ул. Ленина, д. 10-2-3",
	} {
		input := input
		t.Run(input, func(t *testing.T) {
			t.Parallel()
			spans, err := nlp.DetectAddressSpans(context.Background(), input)
			require.NoError(t, err)
			assert.Empty(t, spans)
		})
	}
}

func TestValidHouseIdentifier_WhenBoundedGrammar_ExpectPolicy(t *testing.T) {
	t.Parallel()

	cases := []struct {
		segment string
		want    bool
	}{
		{segment: "10", want: true},
		{segment: "10А", want: true},
		{segment: "10/2", want: true},
		{segment: "10-2", want: true},
		{segment: "10/2/3", want: false},
		{segment: "10-2-3", want: false},
		{segment: "10/2-3", want: false},
		{segment: "10АБ", want: false},
		{segment: "/2", want: false},
		{segment: "10/", want: false},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.segment, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, nlp.ValidHouseIdentifierForTest(tc.segment))
		})
	}
}
