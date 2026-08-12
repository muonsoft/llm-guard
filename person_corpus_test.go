package llmguard_test

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/muonsoft/llm-guard"
	"github.com/stretchr/testify/require"
)

type personCorpusCase struct {
	SchemaVersion     int          `json:"schema_version"`
	ID                string       `json:"id"`
	Category          string       `json:"category"`
	Input             string       `json:"input"`
	Expectation       string       `json:"expectation"`
	Spans             []personSpan `json:"spans"`
	R0ReferenceID     string       `json:"r0_reference_id"`
	DifferentialClass string       `json:"differential_class"`
}

type personSpan struct {
	Start int `json:"start"`
	End   int `json:"end"`
}

type natashaCase struct {
	ID                    string `json:"id"`
	Input                 string `json:"input"`
	ProductExpectation    string `json:"product_expectation"`
	IntentionalDifference string `json:"intentional_difference_class"`
}

type natashaExpected struct {
	ID      string `json:"id"`
	Input   string `json:"input"`
	Matches []struct {
		Entity      string `json:"entity"`
		MatchedText string `json:"matched_text"`
		SpanBytes   struct {
			Start int `json:"start"`
			End   int `json:"end"`
		} `json:"span_bytes"`
	} `json:"matches"`
}

type corpusMetrics struct {
	TP            int
	FP            int
	FN            int
	Precision     float64
	Recall        float64
	ExactSpanRate float64
}

var r0PersonCaseIDs = []string{
	"person-first-last-001",
	"person-last-first-002",
	"person-first-patronymic-last-003",
	"person-last-first-patronymic-004",
	"person-last-initials-005",
	"person-initials-last-006",
	"person-dative-007",
	"person-instrumental-008",
	"person-negative-isolated-first-009",
	"person-negative-isolated-surname-010",
	"person-negative-role-context-011",
	"person-negative-street-like-012",
}

func TestPersonCorpus_WhenVersionedCases_ExpectZeroMandatoryErrors(t *testing.T) {
	t.Parallel()

	cases := loadPersonCorpusCases(t)
	validatePersonCorpusSchema(t, cases)

	natashaCases := loadNatashaCases(t)
	natashaExpected := loadNatashaExpected(t)
	validateR0Linkage(t, cases, natashaCases)

	detector := llmguard.NewPersonDetector()
	metrics := corpusMetrics{}

	for _, tc := range cases {
		findings, err := detector.Detect(context.Background(), tc.Input)
		require.NoError(t, err)

		got := findingsToSpans(findings)
		want := append([]personSpan(nil), tc.Spans...)
		sortSpans(got)
		sortSpans(want)

		tp, fp, fn := spanSetMetrics(got, want)
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

	t.Logf("PERSON corpus metrics: TP=%d FP=%d FN=%d precision=%.3f recall=%.3f exact_span_rate=%.3f",
		metrics.TP, metrics.FP, metrics.FN, metrics.Precision, metrics.Recall, metrics.ExactSpanRate)

	require.Zero(t, metrics.FP, "false positives in mandatory corpus")
	require.Zero(t, metrics.FN, "false negatives in mandatory corpus")

	differentialIDs := verifyNatashaDifferentials(t, cases, natashaCases, natashaExpected)
	t.Logf("R0 intentional Natasha differentials: %v", differentialIDs)
	require.Len(t, differentialIDs, 4)
}

func validatePersonCorpusSchema(t *testing.T, cases []personCorpusCase) {
	t.Helper()

	seen := make(map[string]struct{}, len(cases))
	for _, tc := range cases {
		require.Equal(t, 1, tc.SchemaVersion, "schema_version for %s", tc.ID)
		require.NotEmpty(t, tc.ID)
		_, dup := seen[tc.ID]
		require.False(t, dup, "duplicate corpus id %s", tc.ID)
		seen[tc.ID] = struct{}{}

		require.Contains(t, []string{"positive", "negative"}, tc.Category, "category for %s", tc.ID)
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

func validateR0Linkage(t *testing.T, cases []personCorpusCase, natashaCases []natashaCase) {
	t.Helper()

	byR0 := make(map[string]personCorpusCase)
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
		if strings.HasPrefix(tc.ID, "person-") {
			natashaByID[tc.ID] = tc
		}
	}

	for _, id := range r0PersonCaseIDs {
		productCase, ok := byR0[id]
		require.True(t, ok, "missing linked R0 PERSON case %s", id)
		refCase, ok := natashaByID[id]
		require.True(t, ok, "missing natasha R0 PERSON case %s", id)
		require.Equal(t, refCase.Input, productCase.Input, "input mismatch for %s", id)
	}
}

func verifyNatashaDifferentials(t *testing.T, productCases []personCorpusCase, natashaCases []natashaCase, expected map[string]natashaExpected) []string {
	t.Helper()

	byID := make(map[string]personCorpusCase, len(productCases))
	for _, tc := range productCases {
		if tc.R0ReferenceID != "" {
			byID[tc.R0ReferenceID] = tc
		}
	}

	var differentialIDs []string
	for _, refCase := range natashaCases {
		if !strings.HasPrefix(refCase.ID, "person-") {
			continue
		}
		productCase, ok := byID[refCase.ID]
		if !ok {
			continue
		}
		exp, ok := expected[refCase.ID]
		require.True(t, ok, "missing expected reference for %s", refCase.ID)
		require.Equal(t, refCase.Input, exp.Input)

		refMatches := filterPersonMatches(exp)
		productMatches := productCase.Spans

		if spansEqual(personSpansFromNatasha(refMatches), productMatches) {
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

func spanSetMetrics(got, want []personSpan) (tp, fp, fn int) {
	wantSet := spanSet(want)
	gotSet := spanSet(got)
	for key := range wantSet {
		if _, ok := gotSet[key]; ok {
			tp++
		} else {
			fn++
		}
	}
	for key := range gotSet {
		if _, ok := wantSet[key]; !ok {
			fp++
		}
	}
	return tp, fp, fn
}

func spanSet(spans []personSpan) map[string]struct{} {
	out := make(map[string]struct{}, len(spans))
	for _, span := range spans {
		out[spanKey(span)] = struct{}{}
	}
	return out
}

func spanKey(span personSpan) string {
	return fmt.Sprintf("%d:%d", span.Start, span.End)
}

func findingsToSpans(findings []llmguard.Finding) []personSpan {
	out := make([]personSpan, 0, len(findings))
	for _, finding := range findings {
		out = append(out, personSpan{Start: finding.Start, End: finding.End})
	}
	return out
}

func loadPersonCorpusCases(t *testing.T) []personCorpusCase {
	t.Helper()
	path := filepath.Join(repoRoot(t), "testdata", "person", "cases.jsonl")
	file, err := os.Open(path)
	require.NoError(t, err)
	defer file.Close()

	var cases []personCorpusCase
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var tc personCorpusCase
		require.NoError(t, json.Unmarshal([]byte(line), &tc))
		cases = append(cases, tc)
	}
	require.NoError(t, scanner.Err())
	return cases
}

func loadNatashaCases(t *testing.T) []natashaCase {
	t.Helper()
	path := filepath.Join(repoRoot(t), "testdata", "natasha", "cases.jsonl")
	file, err := os.Open(path)
	require.NoError(t, err)
	defer file.Close()

	var cases []natashaCase
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var tc natashaCase
		require.NoError(t, json.Unmarshal([]byte(line), &tc))
		cases = append(cases, tc)
	}
	require.NoError(t, scanner.Err())
	return cases
}

func loadNatashaExpected(t *testing.T) map[string]natashaExpected {
	t.Helper()
	path := filepath.Join(repoRoot(t), "testdata", "natasha", "expected-python.jsonl")
	file, err := os.Open(path)
	require.NoError(t, err)
	defer file.Close()

	out := make(map[string]natashaExpected)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var exp natashaExpected
		require.NoError(t, json.Unmarshal([]byte(line), &exp))
		out[exp.ID] = exp
	}
	require.NoError(t, scanner.Err())
	return out
}

func filterPersonMatches(exp natashaExpected) []personSpan {
	var spans []personSpan
	for _, match := range exp.Matches {
		if match.Entity != "PERSON" {
			continue
		}
		spans = append(spans, personSpan{Start: match.SpanBytes.Start, End: match.SpanBytes.End})
	}
	sortSpans(spans)
	return spans
}

func personSpansFromNatasha(spans []personSpan) []personSpan {
	out := append([]personSpan(nil), spans...)
	sortSpans(out)
	return out
}

func sortSpans(spans []personSpan) {
	sort.Slice(spans, func(i, j int) bool {
		if spans[i].Start != spans[j].Start {
			return spans[i].Start < spans[j].Start
		}
		return spans[i].End < spans[j].End
	})
}

func spansEqual(a, b []personSpan) bool {
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

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok)
	return filepath.Dir(file)
}
