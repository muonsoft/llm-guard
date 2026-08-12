package nlp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectPersonSpans_WhenContextCancelledAfterTokenize_ExpectError(t *testing.T) {
	t.Parallel()

	ctx := withCancelAfter(context.Background(), func(step string) error {
		if step == "post_tokenize" {
			return context.Canceled
		}
		return nil
	})

	spans, err := DetectPersonSpans(ctx, "Проект Альфа")
	require.ErrorIs(t, err, context.Canceled)
	assert.Nil(t, spans)
}

func TestDetectPersonSpans_WhenContextCancelledDuringScan_ExpectError(t *testing.T) {
	t.Parallel()

	ctx := withCancelAfter(context.Background(), func(step string) error {
		if step == "scan" {
			return context.Canceled
		}
		return nil
	})

	spans, err := DetectPersonSpans(ctx, "Иван Петров")
	require.ErrorIs(t, err, context.Canceled)
	assert.Nil(t, spans)
}

func TestDetectPersonSpans_WhenContextCancelledBeforeTokenize_ExpectError(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	spans, err := DetectPersonSpans(ctx, "Иван Петров")
	require.Error(t, err)
	assert.Nil(t, spans)
}
