package evaluation

import "testing"

func TestApplyContractThresholds_WhenEmptyEntityBoundsInMemory_ExpectZeroRegression(t *testing.T) {
	t.Parallel()
	report := ContractReport{
		Summary: SummaryMetrics{TP: 1, FP: 1, FN: 0},
		Status:  StatusPass,
	}
	ApplyContractThresholds(&report, ThresholdSet{
		ID: "gate",
		Profiles: map[string]ProfileThreshold{
			"contract": {
				Status:   "gate",
				Entities: map[string]NumericBounds{"EMAIL": {}},
			},
		},
	})
	if report.Status != StatusFail {
		t.Fatalf("status = %q, want fail on FP with empty entity bounds", report.Status)
	}
}

func TestApplyContractThresholds_WhenMaxFPAllowsRegression_ExpectPass(t *testing.T) {
	t.Parallel()
	maxFP := 5.0
	report := ContractReport{
		Summary: SummaryMetrics{TP: 10, FP: 1, FN: 0},
		Status:  StatusFail,
	}
	ApplyContractThresholds(&report, ThresholdSet{
		ID: "gate",
		Profiles: map[string]ProfileThreshold{
			"contract": {Status: "gate", MaxFP: &maxFP},
		},
	})
	if report.Status != StatusPass {
		t.Fatalf("status = %q, want pass", report.Status)
	}
}

func TestApplyContractThresholds_WhenMaxFPExceeded_ExpectFail(t *testing.T) {
	t.Parallel()
	maxFP := 5.0
	report := ContractReport{
		Summary: SummaryMetrics{TP: 10, FP: 6, FN: 0},
		Status:  StatusPass,
	}
	ApplyContractThresholds(&report, ThresholdSet{
		ID: "gate",
		Profiles: map[string]ProfileThreshold{
			"contract": {Status: "gate", MaxFP: &maxFP},
		},
	})
	if report.Status != StatusFail {
		t.Fatalf("status = %q, want fail", report.Status)
	}
}

func TestApplyContractThresholds_WhenGateNoBounds_ExpectZeroRegression(t *testing.T) {
	t.Parallel()
	report := ContractReport{
		Summary: SummaryMetrics{TP: 1, FP: 1, FN: 0},
		Status:  StatusPass,
	}
	ApplyContractThresholds(&report, ThresholdSet{
		ID: "gate",
		Profiles: map[string]ProfileThreshold{
			"contract": {Status: "gate"},
		},
	})
	if report.Status != StatusFail {
		t.Fatalf("status = %q, want fail on any FP/FN", report.Status)
	}
}

func TestViolatesContractBounds_WhenMinPrecision_ExpectFailAndPass(t *testing.T) {
	t.Parallel()
	minP := 0.9
	reportFail := ContractReport{Summary: SummaryMetrics{TP: 1, FP: 1, FN: 0}}
	if !violatesContractBounds(&reportFail, ProfileThreshold{MinPrecision: &minP}) {
		t.Fatal("expected fail when precision below floor")
	}
	minP80 := 0.8
	reportPass := ContractReport{Summary: SummaryMetrics{TP: 9, FP: 1, FN: 0}}
	if violatesContractBounds(&reportPass, ProfileThreshold{MinPrecision: &minP80}) {
		t.Fatal("expected pass when precision meets floor")
	}
}

func TestViolatesContractBounds_WhenSourceMaxFP_ExpectPerSourceCheck(t *testing.T) {
	t.Parallel()
	maxFP := 0.0
	report := ContractReport{
		sourceSummaries: sourceContractSummaries{
			"src-a": {TP: 1, FP: 0, FN: 0},
			"src-b": {TP: 1, FP: 1, FN: 0},
		},
	}
	profile := ProfileThreshold{
		Sources: map[string]NumericBounds{
			"src-a": {MaxFP: &maxFP},
			"src-b": {MaxFP: &maxFP},
		},
	}
	if !violatesContractBounds(&report, profile) {
		t.Fatal("expected src-b FP to fail source bound")
	}
}

func TestViolatesExposureBounds_WhenSourceLeakedBytes_ExpectPerSourceCheck(t *testing.T) {
	t.Parallel()
	maxLeaked := 5.0
	report := ExposureReport{
		sourceSummaries: sourceExposureSummaries{
			"src-a": {LeakedSensitiveBytes: 3, ByteCoverage: 0.5},
			"src-b": {LeakedSensitiveBytes: 10, ByteCoverage: 0.5},
		},
	}
	profile := ProfileThreshold{
		Sources: map[string]NumericBounds{
			"src-b": {MaxLeakedSensitiveBytes: &maxLeaked},
		},
	}
	if !violatesExposureBounds(&report, profile) {
		t.Fatal("expected src-b leaked bytes to fail source bound")
	}
}

func TestLoadThresholdSet_WhenLifecycleSources_ExpectRejection(t *testing.T) {
	t.Parallel()
	set := ThresholdSet{
		SchemaVersion: 1,
		ID:            "bad",
		Profiles: map[string]ProfileThreshold{
			"lifecycle": {
				Status:  "gate",
				Sources: map[string]NumericBounds{"src": {}},
			},
		},
	}
	if err := validateThresholdSet(set); err == nil {
		t.Fatal("expected lifecycle sources bounds to be rejected")
	}
}

func TestApplyLifecycleThresholds_WhenMaxFNAllowsDiagnostics_ExpectPass(t *testing.T) {
	t.Parallel()
	maxFN := 2.0
	report := LifecycleReport{
		Diagnostics: []LifecycleFailureDiagnostic{{SourceRecordID: "r1"}},
		Status:      StatusFail,
	}
	ApplyLifecycleThresholds(&report, ThresholdSet{
		ID: "gate",
		Profiles: map[string]ProfileThreshold{
			"lifecycle": {Status: "gate", MaxFN: &maxFN},
		},
	})
	if report.Status != StatusPass {
		t.Fatalf("status = %q, want pass", report.Status)
	}
}
