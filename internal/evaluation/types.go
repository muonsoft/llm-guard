package evaluation

import "github.com/muonsoft/llm-guard"

// ExpectedSpan is an exact UTF-8 byte span annotation for one entity.
type ExpectedSpan struct {
	Entity llmguard.EntityType `json:"entity"`
	Start  int                 `json:"start"`
	End    int                 `json:"end"`
}

// Case is one annotated evaluation input.
type Case struct {
	SchemaVersion int            `json:"schema_version"`
	ID            string         `json:"id"`
	Category      string         `json:"category"`
	Input         string         `json:"input"`
	Expected      []ExpectedSpan `json:"expected"`
}

// EntityMetrics holds per-entity quality counters and derived rates.
type EntityMetrics struct {
	Entity             llmguard.EntityType
	TP                 int
	FP                 int
	FN                 int
	NegativeCases      int
	FalsePositiveCases int
	Precision          float64
	Recall             float64
	F1                 float64
	FPR                float64
	FNR                float64
}

// CoverageReport records whether each MVP entity has positive and negative opportunities.
type CoverageReport struct {
	Complete bool
	Entities []EntityCoverage
}

// EntityCoverage is one entity's corpus opportunity summary.
type EntityCoverage struct {
	Entity      llmguard.EntityType
	HasPositive bool
	HasNegative bool
}

// Report is the deterministic evaluation output for a corpus run.
type Report struct {
	CorpusPath string
	Cases      int
	Entities   []EntityMetrics
	Summary    SummaryMetrics
	Coverage   CoverageReport
}

// SummaryMetrics aggregates corpus-wide counters.
type SummaryMetrics struct {
	TP int
	FP int
	FN int
}

// HasRegression returns true when coverage is incomplete or any FP/FN occurred.
func (r Report) HasRegression() bool {
	return !r.Coverage.Complete || r.Summary.FP > 0 || r.Summary.FN > 0
}
