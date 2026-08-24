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

func TestLoadSuite_WhenMidRuneSpan_ExpectRejection(t *testing.T) {
	t.Parallel()

	input := "абв"
	sum := sha256.Sum256([]byte(input))
	// Span starts inside first Cyrillic rune (2 bytes), not on boundary.
	line := `{"schema_version":2,"suite_id":"s","source_id":"x","source_record_id":"r1","mapping_version":"v1","input":"` + input + `","input_sha256":"` + hex.EncodeToString(sum[:]) + `","annotations":[{"source_label":"EMAIL","mapped_entity":"EMAIL","start":1,"end":3,"disposition":"supported"}]}`
	path := writeSuiteLine(t, line)
	_, err := evaluation.LoadSuite(path)
	require.Error(t, err)
	assert.NotContains(t, err.Error(), input)
	assert.Contains(t, err.Error(), "UTF-8 boundary")
}

func TestLoadSuite_WhenDuplicateSourceRecordID_ExpectRejection(t *testing.T) {
	t.Parallel()

	sum := sha256.Sum256([]byte("x"))
	h := hex.EncodeToString(sum[:])
	line1 := `{"schema_version":2,"suite_id":"s","source_id":"x","source_record_id":"dup","mapping_version":"v1","input":"x","input_sha256":"` + h + `","annotations":[]}`
	line2 := line1
	dir := t.TempDir()
	path := filepath.Join(dir, "suite.jsonl")
	require.NoError(t, os.WriteFile(path, []byte(line1+"\n"+line2+"\n"), 0o600))
	_, err := evaluation.LoadSuite(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate source_record_id")
}

func TestLoadSuite_WhenDuplicateAnnotationSpan_ExpectRejection(t *testing.T) {
	t.Parallel()

	input := "aa"
	sum := sha256.Sum256([]byte(input))
	line := `{"schema_version":2,"suite_id":"s","source_id":"x","source_record_id":"r1","mapping_version":"v1","input":"` + input + `","input_sha256":"` + hex.EncodeToString(sum[:]) + `","annotations":[{"source_label":"L","mapped_entity":"EMAIL","start":0,"end":1,"disposition":"supported"},{"source_label":"L","mapped_entity":"EMAIL","start":0,"end":1,"disposition":"supported"}]}`
	path := writeSuiteLine(t, line)
	_, err := evaluation.LoadSuite(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate span")
}

func TestLoadSourceManifest_WhenEmptyDigest_ExpectRejection(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "manifest.json")
	require.NoError(t, os.WriteFile(path, []byte(`{"schema_version":1,"id":"x","canonical_url":"u","revision":"r","digest_sha256":"","license":"l","attribution":"a","distribution":"cache-only","adapter_version":"1"}`), 0o600))
	_, err := evaluation.LoadSourceManifest(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "digest_sha256 is empty")
}

func TestLoadSourceManifest_WhenMissingAttribution_ExpectRejection(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "manifest.json")
	require.NoError(t, os.WriteFile(path, []byte(`{"schema_version":1,"id":"x","canonical_url":"u","revision":"r","digest_sha256":"abc","license":"l","attribution":"","distribution":"cache-only","adapter_version":"1"}`), 0o600))
	_, err := evaluation.LoadSourceManifest(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "attribution is empty")
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
