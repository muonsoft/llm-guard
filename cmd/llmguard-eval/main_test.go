package main_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	evalBinary     string
	evalBinaryOnce sync.Once
	evalBinaryErr  error
)

func evalCLI(t *testing.T, args ...string) *exec.Cmd {
	t.Helper()
	bin := buildEvalBinary(t)
	cmd := exec.Command(bin, args...)
	cmd.Dir = repoRoot(t)
	return cmd
}

func buildEvalBinary(t *testing.T) string {
	t.Helper()
	evalBinaryOnce.Do(func() {
		dir, err := os.MkdirTemp("", "llmguard-eval-*")
		if err != nil {
			evalBinaryErr = err
			return
		}
		evalBinary = filepath.Join(dir, "llmguard-eval")
		build := exec.Command("go", "build", "-o", evalBinary, "./cmd/llmguard-eval")
		build.Dir = repoRootForBuild()
		evalBinaryErr = build.Run()
	})
	require.NoError(t, evalBinaryErr)
	return evalBinary
}

func TestEvaluationCLI_WhenCommittedCorpus_ExpectMarkdownRegressionGate(t *testing.T) {
	t.Parallel()

	corpus := filepath.Join(repoRoot(t), "testdata", "evaluation", "cases.jsonl")
	cmd := evalCLI(t,
		"-corpus", corpus,
		"-format", "markdown",
		"-fail-on-regression",
	)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))
	assert.Contains(t, string(out), "Aggregate TP=")
}

func TestEvaluationCLI_WhenMissingCorpus_ExpectNonZero(t *testing.T) {
	t.Parallel()

	cmd := evalCLI(t, "-corpus", filepath.Join(repoRoot(t), "missing.jsonl"))
	err := cmd.Run()
	require.Error(t, err)
}

func TestEvaluationCLI_WhenJSONFormat_ExpectValidOutput(t *testing.T) {
	t.Parallel()

	corpus := filepath.Join(repoRoot(t), "testdata", "evaluation", "cases.jsonl")
	cmd := evalCLI(t, "-corpus", corpus, "-format", "json")
	out, err := cmd.Output()
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(string(out), "{\n"))
	assert.Contains(t, string(out), `"entities"`)
}

func TestEvaluationCLI_WhenBothCorpusAndSuite_ExpectExit2(t *testing.T) {
	t.Parallel()

	root := repoRoot(t)
	cmd := evalCLI(t,
		"-corpus", filepath.Join(root, "testdata", "evaluation", "cases.jsonl"),
		"-suite", filepath.Join(root, "testdata", "evaluation", "suites", "contract.jsonl"),
	)
	err := cmd.Run()
	requireExitCode(t, err, 2)
}

func TestEvaluationCLI_WhenSuiteWithoutProfile_ExpectExit2(t *testing.T) {
	t.Parallel()

	root := repoRoot(t)
	cmd := evalCLI(t,
		"-suite", filepath.Join(root, "testdata", "evaluation", "suites", "contract.jsonl"),
	)
	err := cmd.Run()
	requireExitCode(t, err, 2)
}

func TestEvaluationCLI_WhenCorpusWithProfile_ExpectExit2(t *testing.T) {
	t.Parallel()

	root := repoRoot(t)
	cmd := evalCLI(t,
		"-corpus", filepath.Join(root, "testdata", "evaluation", "cases.jsonl"),
		"-profile", "contract",
	)
	err := cmd.Run()
	requireExitCode(t, err, 2)
}

func TestEvaluationCLI_WhenCorpusWithThresholds_ExpectExit2(t *testing.T) {
	t.Parallel()

	root := repoRoot(t)
	cmd := evalCLI(t,
		"-corpus", filepath.Join(root, "testdata", "evaluation", "cases.jsonl"),
		"-thresholds", filepath.Join(root, "testdata", "evaluation", "thresholds", "diagnostic.json"),
	)
	err := cmd.Run()
	requireExitCode(t, err, 2)
}

func TestEvaluationCLI_WhenMissingCorpusAndSuite_ExpectExit2(t *testing.T) {
	t.Parallel()

	cmd := evalCLI(t)
	err := cmd.Run()
	requireExitCode(t, err, 2)
}

func TestEvaluationCLI_WhenContractSuite_ExpectExitZero(t *testing.T) {
	t.Parallel()

	root := repoRoot(t)
	cmd := evalCLI(t,
		"-suite", filepath.Join(root, "testdata", "evaluation", "suites", "contract.jsonl"),
		"-profile", "contract",
		"-format", "json",
	)
	out, err := cmd.Output()
	require.NoError(t, err, string(out))
}

func TestEvaluationCLI_WhenLifecycleProfile_ExpectExit1(t *testing.T) {
	t.Parallel()

	root := repoRoot(t)
	cmd := evalCLI(t,
		"-suite", filepath.Join(root, "testdata", "evaluation", "suites", "contract.jsonl"),
		"-profile", "lifecycle",
	)
	err := cmd.Run()
	requireExitCode(t, err, 1)
}

func TestEvaluationCLI_WhenUnsupportedFormatCorpus_ExpectExit2(t *testing.T) {
	t.Parallel()

	root := repoRoot(t)
	cmd := evalCLI(t,
		"-corpus", filepath.Join(root, "testdata", "evaluation", "cases.jsonl"),
		"-format", "xml",
	)
	out, err := cmd.CombinedOutput()
	requireExitCode(t, err, 2)
	assert.Contains(t, string(out), "unsupported format")
}

func TestEvaluationCLI_WhenUnsupportedFormatSuite_ExpectExit2(t *testing.T) {
	t.Parallel()

	root := repoRoot(t)
	cmd := evalCLI(t,
		"-suite", filepath.Join(root, "testdata", "evaluation", "suites", "contract.jsonl"),
		"-profile", "contract",
		"-format", "xml",
	)
	out, err := cmd.CombinedOutput()
	requireExitCode(t, err, 2)
	assert.Contains(t, string(out), "unsupported format")
}

func TestEvaluationCLI_WhenExposureNoThresholds_ExpectDiagnosticExitZero(t *testing.T) {
	t.Parallel()

	root := repoRoot(t)
	cmd := evalCLI(t,
		"-suite", filepath.Join(root, "testdata", "evaluation", "suites", "exposure.jsonl"),
		"-profile", "exposure",
		"-format", "json",
		"-fail-on-regression",
	)
	out, err := cmd.Output()
	require.NoError(t, err, string(out))
	assert.Contains(t, string(out), `"status": "diagnostic"`)
}

func TestEvaluationCLI_WhenExposureGateThresholdViolated_ExpectExit1(t *testing.T) {
	t.Parallel()

	root := repoRoot(t)
	cmd := evalCLI(t,
		"-suite", filepath.Join(root, "testdata", "evaluation", "suites", "exposure.jsonl"),
		"-profile", "exposure",
		"-thresholds", filepath.Join(root, "testdata", "evaluation", "thresholds", "gate_exposure.json"),
		"-format", "json",
	)
	err := cmd.Run()
	requireExitCode(t, err, 1)
}

func TestEvaluationCLI_WhenExposureMarker_ExpectNoLeakInOutput(t *testing.T) {
	t.Parallel()

	root := repoRoot(t)
	cmd := evalCLI(t,
		"-suite", filepath.Join(root, "testdata", "evaluation", "suites", "exposure.jsonl"),
		"-profile", "exposure",
		"-format", "markdown",
	)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))
	body := string(out)
	assert.NotContains(t, body, "PEB_SAFE_MARKER_XYZ")
	assert.NotContains(t, body, "a@b.co")
}

func repoRootForBuild() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok)
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

func requireExitCode(t *testing.T, err error, code int) {
	t.Helper()
	require.Error(t, err)
	exit, ok := err.(*exec.ExitError)
	require.True(t, ok, "expected exit error, got %v", err)
	require.Equal(t, code, exit.ExitCode())
}
