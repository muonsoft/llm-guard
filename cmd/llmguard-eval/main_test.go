package main_test

import (
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEvaluationCLI_WhenCommittedCorpus_ExpectMarkdownRegressionGate(t *testing.T) {
	t.Parallel()

	root := repoRoot(t)
	corpus := filepath.Join(root, "testdata", "evaluation", "cases.jsonl")
	cmd := exec.Command("go", "run", "./cmd/llmguard-eval",
		"-corpus", corpus,
		"-format", "markdown",
		"-fail-on-regression",
	)
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))
	assert.Contains(t, string(out), "Aggregate TP=")
}

func TestEvaluationCLI_WhenMissingCorpus_ExpectNonZero(t *testing.T) {
	t.Parallel()

	root := repoRoot(t)
	cmd := exec.Command("go", "run", "./cmd/llmguard-eval", "-corpus", filepath.Join(root, "missing.jsonl"))
	cmd.Dir = root
	err := cmd.Run()
	require.Error(t, err)
}

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok)
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

func TestEvaluationCLI_WhenJSONFormat_ExpectValidOutput(t *testing.T) {
	t.Parallel()

	root := repoRoot(t)
	corpus := filepath.Join(root, "testdata", "evaluation", "cases.jsonl")
	cmd := exec.Command("go", "run", "./cmd/llmguard-eval", "-corpus", corpus, "-format", "json")
	cmd.Dir = root
	out, err := cmd.Output()
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(string(out), "{\n"))
	assert.Contains(t, string(out), `"entities"`)
}
