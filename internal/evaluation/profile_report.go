package evaluation

import "github.com/muonsoft/llm-guard"

// ProfileName identifies an evaluation profile.
type ProfileName string

const (
	ProfileContract  ProfileName = "contract"
	ProfileExposure  ProfileName = "exposure"
	ProfileLifecycle ProfileName = "lifecycle"
)

// ProfileStatus values for v2 reports.
const (
	StatusPass       = "pass"
	StatusFail       = "fail"
	StatusDiagnostic = "diagnostic"
)

// ContractDiagnostics holds optional non-gating overlap counts and failure rows.
type ContractDiagnostics struct {
	OverlappingPairs int
	Failures         []ContractFailureDiagnostic
}

// ContractFailureDiagnostic identifies a contract FP/FN without raw input.
type ContractFailureDiagnostic struct {
	SourceRecordID string
	SourceLabel    string
	Entity         string
	Kind           string // "fp" or "fn"
	GoldStart      *int   `json:"gold_start,omitempty"`
	GoldEnd        *int   `json:"gold_end,omitempty"`
	PredictedStart *int   `json:"predicted_start,omitempty"`
	PredictedEnd   *int   `json:"predicted_end,omitempty"`
	InputSHA256    string
}

// ContractReport is the contract profile evaluation output for a suite run.
type ContractReport struct {
	Profile         string
	SuiteID         string
	MappingVersion  string
	ThresholdID     string
	SourceIDs       []string
	Cases           int
	Entities        []EntityMetrics
	Summary         SummaryMetrics
	Diagnostics     ContractDiagnostics
	Status          string
	sourceSummaries sourceContractSummaries
}

// SourceContractSummary holds per-source contract counters for threshold checks.
type SourceContractSummary struct {
	TP int
	FP int
	FN int
}

type sourceContractSummaries map[string]SourceContractSummary

// HasContractRegression returns true when any in-scope FP or FN occurred.
func (r ContractReport) HasContractRegression() bool {
	return r.Summary.FP > 0 || r.Summary.FN > 0
}

// IgnoredCount aggregates ignored annotations by label and reason.
type IgnoredCount struct {
	SourceLabel string
	Reason      string
	Count       int
}

// LabelExposureMetrics aggregates exposure metrics for one label/entity/disposition.
type LabelExposureMetrics struct {
	SourceLabel           string
	MappedEntity          string
	Disposition           string
	SpanCount             int
	FullyCoveredSpanCount int
	SensitiveBytes        int
	CoveredSensitiveBytes int
	LeakedSensitiveBytes  int
	OvermatchedBytes      int
}

// ExposureSummary aggregates suite-wide exposure byte metrics.
type ExposureSummary struct {
	SensitiveBytes        int
	CoveredSensitiveBytes int
	LeakedSensitiveBytes  int
	OvermatchedBytes      int
	ByteCoverage          float64
}

// ExposureReport is the exposure profile evaluation output for a suite run.
type ExposureReport struct {
	Profile         string
	SuiteID         string
	MappingVersion  string
	ThresholdID     string
	SourceIDs       []string
	Cases           int
	ByLabel         []LabelExposureMetrics
	Ignored         []IgnoredCount
	Summary         ExposureSummary
	Status          string
	sourceSummaries sourceExposureSummaries
}

// SourceExposureSummary holds per-source exposure byte metrics for threshold checks.
type SourceExposureSummary struct {
	SensitiveBytes        int
	CoveredSensitiveBytes int
	LeakedSensitiveBytes  int
	OvermatchedBytes      int
	ByteCoverage          float64
}

type sourceExposureSummaries map[string]SourceExposureSummary

func entityScopeOrder(scope []llmguard.EntityType) []llmguard.EntityType {
	out := make([]llmguard.EntityType, len(scope))
	copy(out, scope)
	return out
}
