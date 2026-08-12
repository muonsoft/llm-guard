package llmguard_test

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/muonsoft/llm-guard"
	"github.com/stretchr/testify/require"
)

type addressCorpusCase struct {
	SchemaVersion     int           `json:"schema_version"`
	ID                string        `json:"id"`
	Category          string        `json:"category"`
	Input             string        `json:"input"`
	Expectation       string        `json:"expectation"`
	Spans             []addressSpan `json:"spans"`
	R0ReferenceID     string        `json:"r0_reference_id"`
	DifferentialClass string        `json:"differential_class"`
}

type addressSpan struct {
	Start int `json:"start"`
	End   int `json:"end"`
}

var r0AddressCaseIDs = []string{
	"address-abbreviated-city-street-house-013",
	"address-city-street-house-014",
	"address-street-house-apartment-015",
	"address-prospect-house-016",
	"address-compact-punctuation-017",
	"address-negative-settlement-018",
	"address-negative-region-019",
	"address-negative-street-020",
	"address-ambiguous-city-street-021",
}

func TestAddressCorpus_WhenVersionedCases_ExpectZeroMandatoryErrors(t *testing.T) {
	t.Parallel()

	cases := loadAddressCorpusCases(t)
	validateAddressCorpusSchema(t, cases)

	natashaCases := loadNatashaCases(t)
	natashaExpected := loadNatashaExpected(t)
	validateAddressR0Linkage(t, cases, natashaCases)

	detector := llmguard.NewAddressDetector()
	metrics := corpusMetrics{}

	for _, tc := range cases {
		findings, err := detector.Detect(context.Background(), tc.Input)
		require.NoError(t, err)

		got := addressFindingsToSpans(findings)
		want := append([]addressSpan(nil), tc.Spans...)
		sortAddressSpans(got)
		sortAddressSpans(want)

		tp, fp, fn := spanSetMetrics(toPersonSpans(got), toPersonSpans(want))
		metrics.TP += tp
		metrics.FP += fp
		metrics.FN += fn

		if tp != len(want) || fp != 0 || fn != 0 {
			t.Errorf("case %s: got %v want %v (tp=%d fp=%d fn=%d)", tc.ID, got, want, tp, fp, fn)
		}
	}

	if metrics.TP+metrics.FP > 0 {
		metrics.Precision = float64(metrics.TP) / float64(metrics.TP+metrics.FP)
	}
	if metrics.TP+metrics.FN > 0 {
		metrics.Recall = float64(metrics.TP) / float64(metrics.TP+metrics.FN)
	}
	if metrics.TP+metrics.FN+metrics.FP > 0 {
		metrics.ExactSpanRate = float64(metrics.TP) / float64(metrics.TP+metrics.FN+metrics.FP)
	}

	t.Logf("ADDRESS corpus metrics: TP=%d FP=%d FN=%d precision=%.3f recall=%.3f exact_span_rate=%.3f",
		metrics.TP, metrics.FP, metrics.FN, metrics.Precision, metrics.Recall, metrics.ExactSpanRate)

	require.Zero(t, metrics.FP, "false positives in mandatory corpus")
	require.Zero(t, metrics.FN, "false negatives in mandatory corpus")

	differentialIDs := verifyAddressNatashaDifferentials(t, cases, natashaCases, natashaExpected)
	t.Logf("R0 intentional Natasha differentials: %v", differentialIDs)
}

func validateAddressCorpusSchema(t *testing.T, cases []addressCorpusCase) {
	t.Helper()

	seen := make(map[string]struct{}, len(cases))
	for _, tc := range cases {
		require.Equal(t, 1, tc.SchemaVersion, "schema_version for %s", tc.ID)
		require.NotEmpty(t, tc.ID)
		_, dup := seen[tc.ID]
		require.False(t, dup, "duplicate corpus id %s", tc.ID)
		seen[tc.ID] = struct{}{}

		require.Contains(t, []string{"positive", "negative", "ambiguous"}, tc.Category, "category for %s", tc.ID)
		require.Contains(t, []string{"match", "no_match"}, tc.Expectation, "expectation for %s", tc.ID)

		switch tc.Expectation {
		case "match":
			require.NotEmpty(t, tc.Spans, "match case %s must define spans", tc.ID)
		case "no_match":
			require.Empty(t, tc.Spans, "no_match case %s must not define spans", tc.ID)
		}

		for _, span := range tc.Spans {
			require.GreaterOrEqual(t, span.Start, 0, "span start for %s", tc.ID)
			require.Greater(t, span.End, span.Start, "span end for %s", tc.ID)
			require.LessOrEqual(t, span.End, len(tc.Input), "span end bound for %s", tc.ID)
			require.True(t, utf8.ValidString(tc.Input[span.Start:span.End]), "span rune boundary for %s", tc.ID)
		}

		if tc.DifferentialClass != "" {
			require.Equal(t, "intentional_difference", tc.DifferentialClass, "differential class for %s", tc.ID)
			require.NotEmpty(t, tc.R0ReferenceID, "differential case %s must link R0 id", tc.ID)
		}
	}
}

func validateAddressR0Linkage(t *testing.T, cases []addressCorpusCase, natashaCases []natashaCase) {
	t.Helper()

	byR0 := make(map[string]addressCorpusCase)
	for _, tc := range cases {
		if tc.R0ReferenceID == "" {
			continue
		}
		require.Equal(t, tc.ID, tc.R0ReferenceID, "r0_reference_id must match case id for %s", tc.ID)
		_, dup := byR0[tc.R0ReferenceID]
		require.False(t, dup, "duplicate r0 linkage for %s", tc.R0ReferenceID)
		byR0[tc.R0ReferenceID] = tc
	}

	natashaByID := make(map[string]natashaCase, len(natashaCases))
	for _, tc := range natashaCases {
		if strings.HasPrefix(tc.ID, "address-") {
			natashaByID[tc.ID] = tc
		}
	}

	for _, id := range r0AddressCaseIDs {
		productCase, ok := byR0[id]
		require.True(t, ok, "missing linked R0 ADDRESS case %s", id)
		refCase, ok := natashaByID[id]
		require.True(t, ok, "missing natasha R0 ADDRESS case %s", id)
		require.Equal(t, refCase.Input, productCase.Input, "input mismatch for %s", id)
	}
}

func verifyAddressNatashaDifferentials(t *testing.T, productCases []addressCorpusCase, natashaCases []natashaCase, expected map[string]natashaExpected) []string {
	t.Helper()

	byID := make(map[string]addressCorpusCase, len(productCases))
	for _, tc := range productCases {
		if tc.R0ReferenceID != "" {
			byID[tc.R0ReferenceID] = tc
		}
	}

	var differentialIDs []string
	for _, refCase := range natashaCases {
		if !strings.HasPrefix(refCase.ID, "address-") {
			continue
		}
		productCase, ok := byID[refCase.ID]
		if !ok {
			continue
		}
		exp, ok := expected[refCase.ID]
		require.True(t, ok, "missing expected reference for %s", refCase.ID)
		require.Equal(t, refCase.Input, exp.Input)

		refMatches := filterAddressMatches(exp)
		productMatches := productCase.Spans

		if addressSpansEqual(addressSpansFromNatasha(refMatches), productMatches) {
			continue
		}

		class := productCase.DifferentialClass
		if class == "" {
			class = refCase.IntentionalDifference
		}
		require.NotEmptyf(t, class,
			"unexplained differential for %s: product=%v reference=%v", refCase.ID, productMatches, refMatches)
		require.Equal(t, "intentional_difference", class,
			"differential %s must be intentional_difference, got %s", refCase.ID, class)
		differentialIDs = append(differentialIDs, refCase.ID)
	}

	sort.Strings(differentialIDs)
	return differentialIDs
}

func addressFindingsToSpans(findings []llmguard.Finding) []addressSpan {
	out := make([]addressSpan, 0, len(findings))
	for _, finding := range findings {
		out = append(out, addressSpan{Start: finding.Start, End: finding.End})
	}
	return out
}

func filterAddressMatches(exp natashaExpected) []addressSpan {
	var spans []addressSpan
	for _, match := range exp.Matches {
		if match.Entity != "ADDRESS" {
			continue
		}
		spans = append(spans, addressSpan{Start: match.SpanBytes.Start, End: match.SpanBytes.End})
	}
	sortAddressSpans(spans)
	return spans
}

func addressSpansFromNatasha(spans []addressSpan) []addressSpan {
	out := append([]addressSpan(nil), spans...)
	sortAddressSpans(out)
	return out
}

func sortAddressSpans(spans []addressSpan) {
	sort.Slice(spans, func(i, j int) bool {
		if spans[i].Start != spans[j].Start {
			return spans[i].Start < spans[j].Start
		}
		return spans[i].End < spans[j].End
	})
}

func addressSpansEqual(a, b []addressSpan) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func toPersonSpans(spans []addressSpan) []personSpan {
	out := make([]personSpan, len(spans))
	for i, span := range spans {
		out[i] = personSpan{Start: span.Start, End: span.End}
	}
	return out
}

func loadAddressCorpusCases(t *testing.T) []addressCorpusCase {
	t.Helper()
	path := filepath.Join(repoRoot(t), "testdata", "address", "cases.jsonl")
	file, err := os.Open(path)
	require.NoError(t, err)
	defer file.Close()

	var cases []addressCorpusCase
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var tc addressCorpusCase
		require.NoError(t, json.Unmarshal([]byte(line), &tc))
		cases = append(cases, tc)
	}
	require.NoError(t, scanner.Err())
	return cases
}
