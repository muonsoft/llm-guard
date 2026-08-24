package evaluation

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/muonsoft/llm-guard"
)

const gitleaksFixtureRoot = "testdata/evaluation/generated/gitleaks"

// NormalizeGitleaksFixtures emits normalized records from committed MIT fixtures.
func NormalizeGitleaksFixtures(policy MappingPolicy, emit func(SuiteRecord) error) error {
	fixtures := []struct {
		id          string
		file        string
		annotations func(string, string) []SuiteAnnotation
	}{
		{
			id:   "ignoreCommit.go",
			file: "ignoreCommit.go",
			annotations: func(text, _ string) []SuiteAnnotation {
				token := "AKIALALEMEL33243OLIA"
				start := strings.Index(text, token)
				if start < 0 {
					return nil
				}
				return []SuiteAnnotation{{
					SourceLabel:  "AWS_ACCESS_KEY",
					MappedEntity: string(llmguard.EntitySecretAPIKey),
					Start:        start,
					End:          start + len(token),
					Disposition:  DispositionSupported,
				}}
			},
		},
		{
			id:   "ignoreGlobal.go",
			file: "ignoreGlobal.go",
			annotations: func(text, _ string) []SuiteAnnotation {
				token := "AKIALALEMEL33243OLIA"
				start := strings.Index(text, token)
				if start < 0 {
					return nil
				}
				return []SuiteAnnotation{{
					SourceLabel:  "AWS_ACCESS_KEY",
					MappedEntity: string(llmguard.EntitySecretAPIKey),
					Start:        start,
					End:          start + len(token),
					Disposition:  DispositionSupported,
				}}
			},
		},
		{
			id:          "main.go",
			file:        "main.go",
			annotations: func(string, string) []SuiteAnnotation { return nil },
		},
		{
			id:          "api.go",
			file:        "api.go",
			annotations: func(string, string) []SuiteAnnotation { return nil },
		},
		{
			id:          "README.md",
			file:        "README.md",
			annotations: func(string, string) []SuiteAnnotation { return nil },
		},
		{
			id:   "privatekey.go",
			file: "privatekey.go",
			annotations: func(text, _ string) []SuiteAnnotation {
				begin := "-----BEGIN PRIVATE KEY-----"
				start := strings.Index(text, begin)
				if start < 0 {
					return []SuiteAnnotation{{
						SourceLabel: "_record",
						Start:       0,
						End:         len(text),
						Disposition: DispositionIgnored,
						Reason:      "fixture_truncated_key",
					}}
				}
				end := strings.Index(text[start:], "-----END PRIVATE KEY-----")
				if end < 0 {
					return []SuiteAnnotation{{
						SourceLabel: "_record",
						Start:       0,
						End:         len(text),
						Disposition: DispositionIgnored,
						Reason:      "fixture_truncated_key",
					}}
				}
				end = start + end + len("-----END PRIVATE KEY-----")
				return []SuiteAnnotation{{
					SourceLabel:  "SECRET_PRIVATE_KEY",
					MappedEntity: string(llmguard.EntitySecretPrivateKey),
					Start:        start,
					End:          end,
					Disposition:  DispositionUnsupported,
					Reason:       "",
				}}
			},
		},
	}
	for _, fx := range fixtures {
		path := filepath.Join(gitleaksFixtureRoot, fx.file)
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		text := string(data)
		rec := buildSuiteRecord("gitleaks-reviewed-fixtures", fx.id, policy.Version, text, fx.annotations(text, fx.id))
		if err := emit(rec); err != nil {
			return err
		}
	}
	return nil
}
