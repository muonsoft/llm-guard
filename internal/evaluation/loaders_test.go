package evaluation_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/muonsoft/llm-guard/internal/evaluation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadSourceManifest_WhenValidFixture_ExpectSuccess(t *testing.T) {
	t.Parallel()

	path := filepath.Join(repoRoot(t), "testdata", "evaluation", "manifests", "valid.json")
	manifest, err := evaluation.LoadSourceManifest(path)
	require.NoError(t, err)
	assert.Equal(t, "synthetic-source", manifest.ID)
	assert.Equal(t, "committed-derived", manifest.Distribution)
}

func TestLoadSourceManifest_WhenUnknownField_ExpectRejection(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "manifest.json")
	require.NoError(t, os.WriteFile(path, []byte(`{"schema_version":1,"id":"x","canonical_url":"u","revision":"r","digest_sha256":"d","license":"l","attribution":"a","distribution":"manifest-only","adapter_version":"1","extra":1}`), 0o600))
	_, err := evaluation.LoadSourceManifest(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown field")
}

func TestLoadMappingPolicy_WhenValidFixture_ExpectSuccess(t *testing.T) {
	t.Parallel()

	path := filepath.Join(repoRoot(t), "testdata", "evaluation", "mappings", "valid.json")
	policy, err := evaluation.LoadMappingPolicy(path)
	require.NoError(t, err)
	assert.Equal(t, "test-v1", policy.Version)
	assert.Equal(t, "EMAIL", policy.Direct["EMAIL"])
}

func TestLoadMappingPolicy_WhenUnknownEntity_ExpectRejection(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "mapping.json")
	require.NoError(t, os.WriteFile(path, []byte(`{"schema_version":1,"version":"v1","source_id":"s","direct":{"X":"NOT_AN_ENTITY"}}`), 0o600))
	_, err := evaluation.LoadMappingPolicy(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown product entity")
}

func TestLoadThresholdSet_WhenValidFixture_ExpectSuccess(t *testing.T) {
	t.Parallel()

	path := filepath.Join(repoRoot(t), "testdata", "evaluation", "thresholds", "diagnostic.json")
	set, err := evaluation.LoadThresholdSet(path)
	require.NoError(t, err)
	assert.Equal(t, "diagnostic-default", set.ID)
	assert.Equal(t, "diagnostic", set.ProfileStatus("exposure"))
}

func TestLoadThresholdSet_WhenPrefilterV1Fixture_ExpectSuccess(t *testing.T) {
	t.Parallel()

	path := filepath.Join(repoRoot(t), "testdata", "evaluation", "thresholds", "prefilter-v1.json")
	set, err := evaluation.LoadThresholdSet(path)
	require.NoError(t, err)
	assert.Equal(t, "prefilter-v1", set.ID)
	assert.Equal(t, "gate", set.ProfileStatus("lifecycle"))
}

func TestLoadThresholdSet_WhenUnknownProfile_ExpectRejection(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "thresholds.json")
	require.NoError(t, os.WriteFile(path, []byte(`{"schema_version":1,"id":"t","profiles":{"bad":{"status":"diagnostic"}}}`), 0o600))
	_, err := evaluation.LoadThresholdSet(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown profile")
}
