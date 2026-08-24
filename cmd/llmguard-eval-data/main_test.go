package main_test

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/muonsoft/llm-guard/internal/evaluation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEvalDataCLI_WhenFetchRedMadRobot_ExpectVerifiedCache(t *testing.T) {
	t.Parallel()
	const body = "text,tokens,ner_tags\n"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(body))
	}))
	defer server.Close()

	dir := t.TempDir()
	manifest := evaluation.SourceManifest{
		SchemaVersion:  1,
		ID:             "redmadrobot-pii-benchmark",
		CanonicalURL:   server.URL,
		Revision:       "rev",
		DigestSHA256:   sha256Hex([]byte(body)),
		License:        "MIT",
		Attribution:    "test",
		Distribution:   "cache-only",
		AdapterVersion: "1.0.0",
	}
	err := evaluation.FetchSource(manifest, dir, server.Client())
	require.NoError(t, err)
	_, err = os.Stat(filepath.Join(dir, "sources", manifest.ID, "test.csv"))
	require.NoError(t, err)
}

func TestEvalDataCLI_WhenDigestMismatch_ExpectError(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("bad"))
	}))
	defer server.Close()
	manifest := evaluation.SourceManifest{
		SchemaVersion:  1,
		ID:             "redmadrobot-pii-benchmark",
		CanonicalURL:   server.URL,
		Revision:       "rev",
		DigestSHA256:   "00",
		License:        "MIT",
		Attribution:    "test",
		Distribution:   "cache-only",
		AdapterVersion: "1.0.0",
	}
	err := evaluation.FetchSource(manifest, t.TempDir(), server.Client())
	require.Error(t, err)
}

func TestNormalizeGitleaks_WhenFixturesPresent_ExpectRecords(t *testing.T) {
	t.Chdir(repoRoot(t))
	policy := evaluation.MappingPolicy{
		Version:  "gitleaks-v1",
		SourceID: "gitleaks-reviewed-fixtures",
		Direct:   map[string]string{},
		Person:   nil,
		Address:  nil,
	}
	var count int
	err := evaluation.NormalizeGitleaksFixtures(policy, func(rec evaluation.SuiteRecord) error {
		count++
		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, 6, count)
}

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok)
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
