package nlp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectAddressSpans_WhenContextCancelledAfterTokenize_ExpectError(t *testing.T) {
	t.Parallel()

	ctx := withCancelAfter(context.Background(), func(step string) error {
		if step == "post_tokenize" {
			return context.Canceled
		}
		return nil
	})

	spans, err := DetectAddressSpans(ctx, "ул. Ленина, 10")
	require.ErrorIs(t, err, context.Canceled)
	assert.Nil(t, spans)
}

func TestDetectAddressSpans_WhenContextCancelledDuringScan_ExpectError(t *testing.T) {
	t.Parallel()

	ctx := withCancelAfter(context.Background(), func(step string) error {
		if step == "scan" {
			return context.Canceled
		}
		return nil
	})

	spans, err := DetectAddressSpans(ctx, "ул. Ленина, 10")
	require.ErrorIs(t, err, context.Canceled)
	assert.Nil(t, spans)
}
