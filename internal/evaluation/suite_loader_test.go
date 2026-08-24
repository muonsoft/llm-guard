package evaluation_test

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/muonsoft/llm-guard/internal/evaluation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadSuite_WhenValidFixture_ExpectSuccess(t *testing.T) {
	t.Parallel()

	path := filepath.Join(repoRoot(t), "testdata", "evaluation", "suites", "contract.jsonl")
	suite, err := evaluation.LoadSuite(path)
	require.NoError(t, err)
	assert.Equal(t, "synthetic-contract", suite.SuiteID)
	assert.Equal(t, "test-v1", suite.MappingVersion)
	require.Len(t, suite.Records, 2)
	require.Len(t, suite.Scope, 1)
	assert.Equal(t, "EMAIL", string(suite.Scope[0]))
}

func TestLoadSuite_WhenSHA256Mismatch_ExpectRejection(t *testing.T) {
	t.Parallel()

	input := "bad input"
	sum := sha256.Sum256([]byte(input))
	line := `{"schema_version":2,"suite_id":"s","source_id":"x","source_record_id":"r1","mapping_version":"v1","input":"` + input + `","input_sha256":"` + hex.EncodeToString(sum[:]) + `1","annotations":[]}`
	path := writeSuiteLine(t, line)
	_, err := evaluation.LoadSuite(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "input_sha256 mismatch")
}

func TestLoadSuite_WhenSupportedWithoutEntity_ExpectRejection(t *testing.T) {
	t.Parallel()

	sum := sha256.Sum256([]byte("test"))
	line := `{"schema_version":2,"suite_id":"s","source_id":"x","source_record_id":"r1","mapping_version":"v1","input":"test","input_sha256":"` + hex.EncodeToString(sum[:]) + `","annotations":[{"source_label":"L","mapped_entity":"","start":0,"end":1,"disposition":"supported"}]}`
	path := writeSuiteLine(t, line)
	_, err := evaluation.LoadSuite(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "supported disposition requires mapped_entity")
}

func TestLoadSuite_WhenIgnoredWithoutReason_ExpectRejection(t *testing.T) {
	t.Parallel()

	sum := sha256.Sum256([]byte("test"))
	line := `{"schema_version":2,"suite_id":"s","source_id":"x","source_record_id":"r1","mapping_version":"v1","input":"test","input_sha256":"` + hex.EncodeToString(sum[:]) + `","annotations":[{"source_label":"L","mapped_entity":"","start":0,"end":1,"disposition":"ignored"}]}`
	path := writeSuiteLine(t, line)
	_, err := evaluation.LoadSuite(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ignored disposition requires reason")
}

func TestLoadSuite_WhenUnknownField_ExpectRejection(t *testing.T) {
	t.Parallel()

	sum := sha256.Sum256([]byte("test"))
	line := `{"schema_version":2,"suite_id":"s","source_id":"x","source_record_id":"r1","mapping_version":"v1","input":"test","input_sha256":"` + hex.EncodeToString(sum[:]) + `","annotations":[],"extra":1}`
	path := writeSuiteLine(t, line)
	_, err := evaluation.LoadSuite(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown field")
}

func TestLoadSuite_WhenSpanError_ExpectNoInputSubstring(t *testing.T) {
	t.Parallel()

	input := "SECRET_VALUE_XYZ"
	sum := sha256.Sum256([]byte(input))
	line := `{"schema_version":2,"suite_id":"s","source_id":"x","source_record_id":"rec-secret","mapping_version":"v1","input":"` + input + `","input_sha256":"` + hex.EncodeToString(sum[:]) + `","annotations":[{"source_label":"EMAIL","mapped_entity":"EMAIL","start":0,"end":99,"disposition":"supported"}]}`
	path := writeSuiteLine(t, line)
	_, err := evaluation.LoadSuite(path)
	require.Error(t, err)
	assert.NotContains(t, err.Error(), "SECRET_VALUE_XYZ")
	assert.Contains(t, err.Error(), "rec-secret")
}

func writeSuiteLine(t *testing.T, line string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "suite.jsonl")
	require.NoError(t, os.WriteFile(path, []byte(strings.TrimSpace(line)+"\n"), 0o600))
	return path
}
