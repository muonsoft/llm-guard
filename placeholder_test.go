package llmguard_test

import (
	"testing"

	"github.com/muonsoft/llm-guard"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlaceholder_WhenValidGrammar_ExpectParsed(t *testing.T) {
	t.Parallel()

	token := "{{LLMG_0123456789abcdef0123456789abcdef_0001}}"
	tokens := llmguard.ScanValidPlaceholdersForTest(token + " tail")
	require.Len(t, tokens, 1)
	assert.Equal(t, token, tokens[0])
}

func TestPlaceholder_WhenCounterAbove9999_ExpectParsed(t *testing.T) {
	t.Parallel()

	token := "{{LLMG_0123456789abcdef0123456789abcdef_10000}}"
	tokens := llmguard.ScanValidPlaceholdersForTest(token)
	require.Len(t, tokens, 1)
	assert.Equal(t, token, tokens[0])
}

func TestPlaceholder_WhenInvalidGrammar_ExpectIgnored(t *testing.T) {
	t.Parallel()

	tokens := llmguard.ScanValidPlaceholdersForTest("{{LLMG_bad}} {{not a token}}")
	assert.Empty(t, tokens)
}

func TestPlaceholder_WhenRepeatedUnknownTokens_ExpectSeparateMissCount(t *testing.T) {
	t.Parallel()

	token := "{{LLMG_0123456789abcdef0123456789abcdef_0001}}"
	response := token + " " + token
	misses := llmguard.CountRestoreMissesForTest(response, &llmguard.TokenSet{})
	assert.Equal(t, 2, misses)
}

func TestPlaceholder_WhenKnownTokenInResponse_ExpectZeroMissesBeforeRestore(t *testing.T) {
	t.Parallel()

	token := "{{LLMG_0123456789abcdef0123456789abcdef_0001}}"
	tokens := llmguard.NewTokenSetForTest(token, "secret")
	misses := llmguard.CountRestoreMissesForTest("prefix "+token, tokens)
	assert.Zero(t, misses)
}

func TestPlaceholder_WhenMutatedToken_ExpectMiss(t *testing.T) {
	t.Parallel()

	valid := "{{LLMG_0123456789abcdef0123456789abcdef_0001}}"
	mutated := "{{LLMG_0123456789abcdef0123456789abcdef_0002}}"
	tokens := llmguard.NewTokenSetForTest(valid, "secret")
	misses := llmguard.CountRestoreMissesForTest(mutated, tokens)
	assert.Equal(t, 1, misses)
}
