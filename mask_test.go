package llmguard_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"

	muerrors "github.com/muonsoft/errors"
	"github.com/muonsoft/llm-guard"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newMaskGuard(t *testing.T, opts ...llmguard.Option) *llmguard.Guard {
	t.Helper()
	base := []llmguard.Option{llmguard.WithDetector(llmguard.NewEmailDetector())}
	guard, err := llmguard.New(append(base, opts...)...)
	require.NoError(t, err)
	return guard
}

type sequenceReader struct {
	mu     sync.Mutex
	chunks [][]byte
	err    error
}

func (r *sequenceReader) Read(p []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.err != nil {
		return 0, r.err
	}
	if len(r.chunks) == 0 {
		return 0, io.EOF
	}
	n := copy(p, r.chunks[0])
	r.chunks[0] = r.chunks[0][n:]
	if len(r.chunks[0]) == 0 {
		r.chunks = r.chunks[1:]
	}
	return n, nil
}

func TestMask_WhenEmailPresent_ExpectPlaceholderAndRestore(t *testing.T) {
	t.Parallel()

	guard := newMaskGuard(t)
	text := "Contact a@b.co please"
	result, err := guard.Mask(context.Background(), text)
	require.NoError(t, err)
	assert.NotEqual(t, text, result.Text)
	assert.Contains(t, result.Text, "{{LLMG_")
	require.Len(t, result.Findings, 1)

	restored, err := guard.Restore(context.Background(), result.Text, result.Tokens)
	require.NoError(t, err)
	assert.Equal(t, text, restored)
}

func TestMask_WhenNoFindings_ExpectUnchangedTextAndEmptyMappings(t *testing.T) {
	t.Parallel()

	guard := newMaskGuard(t)
	text := "plain text without mailbox"
	result, err := guard.Mask(context.Background(), text)
	require.NoError(t, err)
	assert.Equal(t, text, result.Text)
	assert.Nil(t, result.Findings)
	require.NotNil(t, result.Tokens)

	restored, err := guard.Restore(context.Background(), result.Text, result.Tokens)
	require.NoError(t, err)
	assert.Equal(t, text, restored)
}

func TestMask_WhenRepeatedMailbox_ExpectDistinctTokens(t *testing.T) {
	t.Parallel()

	guard := newMaskGuard(t)
	text := "a@b.co and a@b.co"
	result, err := guard.Mask(context.Background(), text)
	require.NoError(t, err)
	require.Len(t, result.Findings, 2)

	first := strings.Index(result.Text, "{{LLMG_")
	second := strings.Index(result.Text[first+len("{{LLMG_"):], "{{LLMG_")
	require.NotEqual(t, -1, first)
	require.NotEqual(t, -1, second)
	firstToken := extractMaskedTokenAt(result.Text, first)
	secondToken := extractMaskedTokenAt(result.Text, first+len("{{LLMG_")+second)
	assert.NotEqual(t, firstToken, secondToken)

	restored, err := guard.Restore(context.Background(), result.Text, result.Tokens)
	require.NoError(t, err)
	assert.Equal(t, text, restored)
}

func TestMask_WhenMixedUnicode_ExpectSurroundingBytesPreserved(t *testing.T) {
	t.Parallel()

	guard := newMaskGuard(t)
	text := "Привет a@b.co мир"
	result, err := guard.Mask(context.Background(), text)
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(result.Text, "Привет "))
	assert.True(t, strings.HasSuffix(result.Text, " мир"))
	assert.NotContains(t, result.Text, "a@b.co")

	restored, err := guard.Restore(context.Background(), result.Text, result.Tokens)
	require.NoError(t, err)
	assert.Equal(t, text, restored)
}

func TestMask_WhenInputContainsPlaceholderLikeFragment_ExpectNewNamespace(t *testing.T) {
	t.Parallel()

	guard := newMaskGuard(t)
	existing := "{{LLMG_0123456789abcdef0123456789abcdef_0001}}"
	text := "mail " + existing + " and a@b.co"
	result, err := guard.Mask(context.Background(), text)
	require.NoError(t, err)
	assert.Contains(t, result.Text, existing)
	assert.NotContains(t, result.Text, "a@b.co")
	assert.Equal(t, 2, strings.Count(result.Text, "{{LLMG_"))
}

func extractMaskedToken(masked string) string {
	return extractMaskedTokenAt(masked, strings.Index(masked, "{{LLMG_"))
}

func extractMaskedTokenAt(masked string, start int) string {
	if start == -1 {
		return ""
	}
	end := strings.Index(masked[start:], "}}")
	if end == -1 {
		return ""
	}
	return masked[start : start+end+2]
}

func TestMask_WhenRandomSourceFails_ExpectSafeError(t *testing.T) {
	t.Parallel()

	reader := &sequenceReader{err: io.ErrClosedPipe}
	guard := newMaskGuard(t, llmguard.WithRandomSource(reader))
	_, err := guard.Mask(context.Background(), "a@b.co")
	require.Error(t, err)
	assert.ErrorIs(t, err, llmguard.ErrNamespaceSource)
	assert.NotContains(t, err.Error(), "a@b.co")
}

func TestMask_WhenNamespaceCollisionExhausted_ExpectCollisionError(t *testing.T) {
	t.Parallel()

	const namespace = "deadbeefdeadbeefdeadbeefdeadbeef"
	token := fmt.Sprintf("{{LLMG_%s_0001}}", namespace)
	chunk := bytes.Repeat([]byte{0xde, 0xad, 0xbe, 0xef}, 4)
	chunks := make([][]byte, 33)
	for i := range chunks {
		chunks[i] = append([]byte(nil), chunk...)
	}
	reader := &sequenceReader{chunks: chunks}
	guard := newMaskGuard(t, llmguard.WithRandomSource(reader))

	text := token + " a@b.co"
	_, err := guard.Mask(context.Background(), text)
	require.Error(t, err)
	assert.ErrorIs(t, err, llmguard.ErrNamespaceCollision)
}

func TestMask_WhenContextCanceledBeforeCall_ExpectCanceled(t *testing.T) {
	t.Parallel()

	guard := newMaskGuard(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := guard.Mask(ctx, "a@b.co")
	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestRestore_WhenKnownTokenRepeated_ExpectAllRestored(t *testing.T) {
	t.Parallel()

	guard := newMaskGuard(t)
	text := "a@b.co"
	result, err := guard.Mask(context.Background(), text)
	require.NoError(t, err)

	token := extractMaskedToken(result.Text)
	doubled := token + " and " + token
	restored, err := guard.Restore(context.Background(), doubled, result.Tokens)
	require.NoError(t, err)
	assert.Equal(t, "a@b.co and a@b.co", restored)
}

func TestRestore_WhenUnknownToken_ExpectUnchanged(t *testing.T) {
	t.Parallel()

	guard := newMaskGuard(t)
	empty, err := guard.Mask(context.Background(), "plain text")
	require.NoError(t, err)

	unknown := "{{LLMG_ffffffffffffffffffffffffffffffff_9999}}"
	restored, err := guard.Restore(context.Background(), "keep "+unknown, empty.Tokens)
	require.NoError(t, err)
	assert.Equal(t, "keep "+unknown, restored)
}

func TestRestore_WhenMutatedToken_ExpectUnchanged(t *testing.T) {
	t.Parallel()

	guard := newMaskGuard(t)
	result, err := guard.Mask(context.Background(), "a@b.co")
	require.NoError(t, err)

	token := extractMaskedToken(result.Text)
	mutated := token[:len(token)-1] + "X}"
	restored, err := guard.Restore(context.Background(), mutated, result.Tokens)
	require.NoError(t, err)
	assert.Equal(t, mutated, restored)
}

func TestRestore_WhenCrossTokenSet_ExpectUnknownLeftAlone(t *testing.T) {
	t.Parallel()

	guard := newMaskGuard(t)
	first, err := guard.Mask(context.Background(), "a@b.co")
	require.NoError(t, err)
	second, err := guard.Mask(context.Background(), "c@d.org")
	require.NoError(t, err)

	mixed := first.Text + " " + second.Text
	restoredFirst, err := guard.Restore(context.Background(), mixed, first.Tokens)
	require.NoError(t, err)
	assert.Contains(t, restoredFirst, "a@b.co")
	assert.Contains(t, restoredFirst, extractMaskedToken(second.Text))
}

func TestRestore_WhenValueLooksLikeToken_ExpectNoRecursiveSubstitution(t *testing.T) {
	t.Parallel()

	tokenLike := "{{LLMG_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa_0001}}"
	custom := &tokenValueDetector{
		name: "token-like",
	}
	guard, err := llmguard.New(llmguard.WithDetector(custom), llmguard.WithRandomSource(&sequenceReader{
		chunks: [][]byte{bytes.Repeat([]byte{0xab}, 16)},
	}))
	require.NoError(t, err)

	result, err := guard.Mask(context.Background(), tokenLike)
	require.NoError(t, err)

	restored, err := guard.Restore(context.Background(), result.Text, result.Tokens)
	require.NoError(t, err)
	assert.Equal(t, tokenLike, restored)
}

type tokenValueDetector struct {
	name string
}

func (d *tokenValueDetector) Name() string { return d.name }

func (d *tokenValueDetector) Detect(_ context.Context, text string) ([]llmguard.Finding, error) {
	if text == "" {
		return nil, nil
	}
	return []llmguard.Finding{{
		Entity: llmguard.EntityEmail, Start: 0, End: len(text), Confidence: 1, Detector: d.name,
	}}, nil
}

func TestRestore_WhenNilTokenSet_ExpectInvalidTokenSet(t *testing.T) {
	t.Parallel()

	guard := newMaskGuard(t)
	_, err := guard.Restore(context.Background(), "text", nil)
	require.Error(t, err)
	assert.ErrorIs(t, err, llmguard.ErrInvalidTokenSet)

	inv, ok := muerrors.As[*llmguard.InvalidTokenSetError](err)
	require.True(t, ok)
	assert.Equal(t, "nil token set", inv.Reason)
}

func TestRestore_WhenInvalidUTF8_ExpectInvalidText(t *testing.T) {
	t.Parallel()

	guard := newMaskGuard(t)
	result, err := guard.Mask(context.Background(), "a@b.co")
	require.NoError(t, err)

	_, err = guard.Restore(context.Background(), string([]byte{0xff}), result.Tokens)
	require.Error(t, err)
	assert.ErrorIs(t, err, llmguard.ErrInvalidText)
}

func TestRestore_WhenContextCanceled_ExpectCanceled(t *testing.T) {
	t.Parallel()

	guard := newMaskGuard(t)
	result, err := guard.Mask(context.Background(), "a@b.co")
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = guard.Restore(ctx, result.Text, result.Tokens)
	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestToken_WhenFormatted_ExpectRedactedOutput(t *testing.T) {
	t.Parallel()

	guard := newMaskGuard(t)
	result, err := guard.Mask(context.Background(), "a@b.co")
	require.NoError(t, err)

	summary := fmt.Sprintf("%v %#v %+v", result.Tokens, result.Tokens, result.Tokens)
	assert.Equal(t, "llmguard.TokenSet llmguard.TokenSet llmguard.TokenSet", summary)
	assert.NotContains(t, summary, "a@b.co")
	assert.NotContains(t, summary, "LLMG_")
}

func TestToken_WhenJSONMarshaled_ExpectNoSensitiveFields(t *testing.T) {
	t.Parallel()

	guard := newMaskGuard(t)
	result, err := guard.Mask(context.Background(), "a@b.co")
	require.NoError(t, err)

	tokenJSON, err := json.Marshal(result.Tokens)
	require.NoError(t, err)
	assert.JSONEq(t, `"llmguard.TokenSet"`, string(tokenJSON))

	maskJSON, err := json.Marshal(result)
	require.NoError(t, err)
	assert.NotContains(t, string(maskJSON), "a@b.co")
	assert.NotContains(t, string(maskJSON), "LLMG_")
	assert.Contains(t, string(maskJSON), "llmguard.MaskResult.text")
}

func TestMask_WhenConcurrentCalls_ExpectIndependentTokenSets(t *testing.T) {
	t.Parallel()

	guard := newMaskGuard(t)
	const workers = 16
	type workerOutcome struct {
		maskedToken string
		restored    string
		maskErr     error
		restoreErr  error
	}

	results := make(chan workerOutcome, workers)
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			var outcome workerOutcome
			result, err := guard.Mask(context.Background(), "a@b.co")
			outcome.maskErr = err
			if err != nil {
				results <- outcome
				return
			}
			outcome.maskedToken = extractMaskedToken(result.Text)
			outcome.restored, outcome.restoreErr = guard.Restore(context.Background(), result.Text, result.Tokens)
			results <- outcome
		}()
	}
	wg.Wait()
	close(results)

	tokens := make([]string, 0, workers)
	for outcome := range results {
		require.NoError(t, outcome.maskErr)
		require.NoError(t, outcome.restoreErr)
		assert.Equal(t, "a@b.co", outcome.restored)
		require.NotEmpty(t, outcome.maskedToken)
		tokens = append(tokens, outcome.maskedToken)
	}
	require.Len(t, tokens, workers)
	for i := 0; i < len(tokens); i++ {
		for j := i + 1; j < len(tokens); j++ {
			assert.NotEqual(t, tokens[i], tokens[j])
		}
	}
}

func TestMask_WhenTypedNilRandomSource_ExpectInvalidConfig(t *testing.T) {
	t.Parallel()

	var reader *bytes.Reader
	_, err := llmguard.New(llmguard.WithRandomSource(reader))
	require.Error(t, err)
	assert.ErrorIs(t, err, llmguard.ErrInvalidConfig)
}

func TestMask_WhenNilRandomSource_ExpectInvalidConfig(t *testing.T) {
	t.Parallel()

	_, err := llmguard.New(llmguard.WithRandomSource(nil))
	require.Error(t, err)
	assert.ErrorIs(t, err, llmguard.ErrInvalidConfig)
}

func TestMask_WhenDetectionFails_ExpectNoPartialResult(t *testing.T) {
	t.Parallel()

	failing := &staticDetector{name: "fail", err: io.ErrUnexpectedEOF}
	guard, err := llmguard.New(
		llmguard.WithDetector(llmguard.NewEmailDetector()),
		llmguard.WithDetector(failing),
	)
	require.NoError(t, err)

	_, err = guard.Mask(context.Background(), "a@b.co")
	require.Error(t, err)
	assert.ErrorIs(t, err, llmguard.ErrDetector)
}

func TestMask_WhenIndependentInvocations_ExpectDifferentNamespaces(t *testing.T) {
	t.Parallel()

	guard := newMaskGuard(t)
	first, err := guard.Mask(context.Background(), "a@b.co")
	require.NoError(t, err)
	second, err := guard.Mask(context.Background(), "a@b.co")
	require.NoError(t, err)
	assert.NotEqual(t, extractMaskedToken(first.Text), extractMaskedToken(second.Text))
}
