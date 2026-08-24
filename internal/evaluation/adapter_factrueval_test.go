package evaluation_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/muonsoft/llm-guard/internal/evaluation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFactRuEvalAdapter_WhenSyntheticFixture_ExpectPersonMapping(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	split := "devset"
	stem := "doc1"
	text := "Иван Петров"
	require.NoError(t, os.MkdirAll(filepath.Join(dir, split), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, split, stem+".txt"), []byte(text), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, split, stem+".tokens"), []byte("1 0 4 Иван\n2 5 6 Петров\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, split, stem+".spans"), []byte("1 name 0 4 1 1\n2 surname 5 6 2 1\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, split, stem+".objects"), []byte("1 Person 1 2\n"), 0o600))

	policyPath := filepath.Join(repoRoot(t), "testdata", "evaluation", "mappings", "factrueval-v1.json")
	policy, err := evaluation.LoadMappingPolicy(policyPath)
	require.NoError(t, err)
	adapter := evaluation.FactRuEvalAdapter{Policy: policy}
	rec, err := adapter.AdaptDocument(split, stem, text,
		filepath.Join(dir, split, stem+".tokens"),
		filepath.Join(dir, split, stem+".spans"),
		filepath.Join(dir, split, stem+".objects"),
	)
	require.NoError(t, err)
	assert.Equal(t, "devset/doc1", rec.SourceRecordID)
	var personSupported bool
	for _, ann := range rec.Annotations {
		if ann.MappedEntity == "PERSON" && ann.Disposition == evaluation.DispositionSupported {
			personSupported = true
		}
	}
	assert.True(t, personSupported)
}

func TestFactRuEvalAdapter_WhenPersonAndOrg_ExpectUnsupportedExposureSpans(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	split := "devset"
	stem := "doc2"
	text := "Иван Петров в Ростелеком"
	require.NoError(t, os.MkdirAll(filepath.Join(dir, split), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, split, stem+".txt"), []byte(text), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, split, stem+".tokens"), []byte("1 0 4 Иван\n2 5 6 Петров\n3 14 10 Ростелеком\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, split, stem+".spans"), []byte("1 name 0 4 1 1\n2 surname 5 6 2 1\n3 org 14 10 3 1\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, split, stem+".objects"), []byte("1 Person 1 2\n2 Org 3\n"), 0o600))

	policyPath := filepath.Join(repoRoot(t), "testdata", "evaluation", "mappings", "factrueval-v1.json")
	policy, err := evaluation.LoadMappingPolicy(policyPath)
	require.NoError(t, err)
	adapter := evaluation.FactRuEvalAdapter{Policy: policy}
	rec, err := adapter.AdaptDocument(split, stem, text,
		filepath.Join(dir, split, stem+".tokens"),
		filepath.Join(dir, split, stem+".spans"),
		filepath.Join(dir, split, stem+".objects"),
	)
	require.NoError(t, err)

	var personSupported, orgUnsupported bool
	var orgStart, orgEnd int
	for _, ann := range rec.Annotations {
		switch ann.SourceLabel {
		case "Org":
			if ann.Disposition == evaluation.DispositionUnsupported {
				orgUnsupported = true
				orgStart = ann.Start
				orgEnd = ann.End
			}
		default:
			if ann.MappedEntity == "PERSON" && ann.Disposition == evaluation.DispositionSupported {
				personSupported = true
			}
		}
	}
	assert.True(t, personSupported)
	assert.True(t, orgUnsupported)
	assert.Equal(t, "Ростелеком", text[orgStart:orgEnd])
}
