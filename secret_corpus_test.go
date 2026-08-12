package llmguard_test

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"unicode/utf8"

	"github.com/muonsoft/llm-guard"
	"github.com/stretchr/testify/require"
)

type secretCorpusCase struct {
	SchemaVersion int          `json:"schema_version"`
	ID            string       `json:"id"`
	Category      string       `json:"category"`
	Family        string       `json:"family"`
	Shape         string       `json:"shape"`
	Input         string       `json:"input"`
	Expectation   string       `json:"expectation"`
	Spans         []secretSpan `json:"spans"`
}

type secretSpan struct {
	Start int `json:"start"`
	End   int `json:"end"`
}

var (
	allowedSecretCategories   = map[string]struct{}{"positive": {}, "negative": {}, "malformed": {}}
	allowedSecretExpectations = map[string]struct{}{"match": {}, "no_match": {}}
	allowedSecretFamilies     = map[string]struct{}{"jwt": {}, "pem": {}, "api_key": {}, "dsn": {}}
	requiredPositiveShapes    = map[string]struct{}{
		"compact_signed":            {},
		"rsa_private_key":           {},
		"private_key":               {},
		"ec_private_key":            {},
		"openssh_private_key":       {},
		"pgp_private_key_block":     {},
		"github_ghp":                {},
		"github_pat":                {},
		"gitlab_glpat":              {},
		"aws_akia":                  {},
		"aws_asia":                  {},
		"openai_sk":                 {},
		"openai_sk_proj":            {},
		"postgres":                  {},
		"mysql":                     {},
		"redis":                     {},
		"postgres_ipv6_authority":   {},
		"postgres_ipv6_punctuation": {},
	}
)

func loadSecretCorpus(t *testing.T) []secretCorpusCase {
	t.Helper()
	path := filepath.Join(repoRoot(t), "testdata", "secrets", "cases.jsonl")

	f, err := os.Open(path)
	require.NoError(t, err)
	defer f.Close()

	var cases []secretCorpusCase
	scanner := bufio.NewScanner(f)
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var raw map[string]json.RawMessage
		require.NoError(t, json.Unmarshal(line, &raw), "line %d", lineNo)
		validateJSONObjectKeys(t, raw, map[string]struct{}{
			"schema_version": {},
			"id":             {},
			"category":       {},
			"family":         {},
			"shape":          {},
			"input":          {},
			"expectation":    {},
			"spans":          {},
		}, fmt.Sprintf("line %d", lineNo))

		if spansRaw, ok := raw["spans"]; ok {
			validateSpanArrayKeys(t, spansRaw, fmt.Sprintf("line %d spans", lineNo))
		}

		var c secretCorpusCase
		require.NoError(t, json.Unmarshal(line, &c), "line %d", lineNo)
		cases = append(cases, c)
	}
	require.NoError(t, scanner.Err())
	require.NotEmpty(t, cases)
	return cases
}

func validateJSONObjectKeys(t *testing.T, obj map[string]json.RawMessage, allowed map[string]struct{}, context string) {
	t.Helper()
	for key := range obj {
		if _, ok := allowed[key]; !ok {
			t.Fatalf("unexpected field %q in %s", key, context)
		}
	}
}

func validateSpanArrayKeys(t *testing.T, raw json.RawMessage, context string) {
	t.Helper()
	var spans []map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(raw, &spans), context)
	for i, span := range spans {
		validateJSONObjectKeys(t, span, map[string]struct{}{
			"start": {},
			"end":   {},
		}, fmt.Sprintf("%s index %d", context, i))
	}
}

func validateSecretCorpusSchema(t *testing.T, cases []secretCorpusCase) {
	t.Helper()

	seen := make(map[string]struct{}, len(cases))
	positiveShapes := make(map[string]struct{})

	for _, c := range cases {
		require.Equal(t, 1, c.SchemaVersion, "schema version mismatch for case %s", c.ID)
		require.NotEmpty(t, c.ID)
		if _, exists := seen[c.ID]; exists {
			t.Fatalf("duplicate corpus id %s", c.ID)
		}
		seen[c.ID] = struct{}{}

		if _, ok := allowedSecretCategories[c.Category]; !ok {
			t.Fatalf("invalid category for case %s", c.ID)
		}
		if _, ok := allowedSecretExpectations[c.Expectation]; !ok {
			t.Fatalf("invalid expectation for case %s", c.ID)
		}
		if _, ok := allowedSecretFamilies[c.Family]; !ok {
			t.Fatalf("invalid family for case %s", c.ID)
		}

		switch c.Category {
		case "positive":
			require.Equal(t, "match", c.Expectation, "category/expectation mismatch for case %s", c.ID)
			require.NotEmpty(t, c.Shape, "missing shape for positive case %s", c.ID)
			require.NotEmpty(t, c.Spans, "missing spans for positive case %s", c.ID)
			positiveShapes[c.Shape] = struct{}{}
		case "negative", "malformed":
			require.Equal(t, "no_match", c.Expectation, "category/expectation mismatch for case %s", c.ID)
			require.Empty(t, c.Shape, "unexpected shape for non-positive case %s", c.ID)
			require.Empty(t, c.Spans, "unexpected spans for no_match case %s", c.ID)
		}

		if c.Expectation == "no_match" {
			require.Empty(t, c.Spans, "unexpected spans for no_match case %s", c.ID)
		}

		require.True(t, utf8.ValidString(c.Input), "invalid utf-8 for case %s", c.ID)
		for i, span := range c.Spans {
			require.GreaterOrEqual(t, span.Start, 0, "span start for case %s index %d", c.ID, i)
			require.Greater(t, span.End, span.Start, "span end for case %s index %d", c.ID, i)
			require.LessOrEqual(t, span.End, len(c.Input), "span end for case %s index %d", c.ID, i)
			if span.Start < len(c.Input) {
				require.True(t, utf8.RuneStart(c.Input[span.Start]), "span start boundary for case %s", c.ID)
			}
			if span.End < len(c.Input) {
				require.True(t, utf8.RuneStart(c.Input[span.End]), "span end boundary for case %s", c.ID)
			}
		}
	}

	for shape := range requiredPositiveShapes {
		if _, ok := positiveShapes[shape]; !ok {
			t.Fatalf("missing required positive shape %q", shape)
		}
	}
}

func secretDetectorsByFamily() map[string]llmguard.Detector {
	return map[string]llmguard.Detector{
		"jwt":     llmguard.NewJWTDetector(),
		"pem":     llmguard.NewPEMPrivateKeyDetector(),
		"api_key": llmguard.NewAPIKeyDetector(),
		"dsn":     llmguard.NewDSNDetector(),
	}
}

func TestSecretCorpus_WhenLoaded_ExpectValidSchema(t *testing.T) {
	cases := loadSecretCorpus(t)
	validateSecretCorpusSchema(t, cases)
}

func TestSecretCorpus_WhenVersionedCases_ExpectZeroMandatoryErrors(t *testing.T) {
	cases := loadSecretCorpus(t)
	validateSecretCorpusSchema(t, cases)
	detectors := secretDetectorsByFamily()

	for _, c := range cases {
		t.Run(c.ID, func(t *testing.T) {
			detector, ok := detectors[c.Family]
			require.True(t, ok, "unknown family for case %s", c.ID)

			findings, err := detector.Detect(context.Background(), c.Input)
			require.NoError(t, err)

			switch c.Expectation {
			case "match":
				require.Len(t, findings, len(c.Spans), "finding count mismatch for case %s", c.ID)
				for i, span := range c.Spans {
					require.Equal(t, span.Start, findings[i].Start, "start mismatch for case %s span %d", c.ID, i)
					require.Equal(t, span.End, findings[i].End, "end mismatch for case %s span %d", c.ID, i)
				}
			case "no_match":
				require.Empty(t, findings, "unexpected findings for case %s", c.ID)
			default:
				t.Fatalf("unknown expectation for case %s", c.ID)
			}
		})
	}
}

func TestSecretCorpus_WhenDetectResolve_ExpectDeterministic(t *testing.T) {
	cases := loadSecretCorpus(t)
	matchCases := make([]secretCorpusCase, 0)
	for _, c := range cases {
		if c.Expectation == "match" {
			matchCases = append(matchCases, c)
		}
	}
	require.NotEmpty(t, matchCases)

	guard, err := llmguard.New(
		llmguard.WithDetector(llmguard.NewJWTDetector()),
		llmguard.WithDetector(llmguard.NewPEMPrivateKeyDetector()),
		llmguard.WithDetector(llmguard.NewAPIKeyDetector()),
		llmguard.WithDetector(llmguard.NewDSNDetector()),
	)
	require.NoError(t, err)

	for _, c := range matchCases {
		t.Run(c.ID, func(t *testing.T) {
			findings, err := guard.Detect(context.Background(), c.Input)
			require.NoError(t, err)
			resolved, err := llmguard.Resolve(c.Input, findings)
			require.NoError(t, err)
			require.NotEmpty(t, resolved)
		})
	}
}
