package evaluation

import "testing"

func TestApplyLifecycleThresholds_WhenGateAndRegression_ExpectFail(t *testing.T) {
	t.Parallel()
	report := LifecycleReport{
		Diagnostics: []LifecycleFailureDiagnostic{{SourceRecordID: "r1"}},
		Status:      StatusPass,
	}
	ApplyLifecycleThresholds(&report, ThresholdSet{
		ID: "gate",
		Profiles: map[string]ProfileThreshold{
			"lifecycle": {Status: "gate", MaxFN: floatPtr(0)},
		},
	})
	if report.Status != StatusFail {
		t.Fatalf("status = %q, want fail", report.Status)
	}
}

func TestApplyLifecycleThresholds_WhenDiagnostic_ExpectDiagnosticStatus(t *testing.T) {
	t.Parallel()
	report := LifecycleReport{
		Diagnostics: []LifecycleFailureDiagnostic{{SourceRecordID: "r1"}},
		Status:      StatusPass,
	}
	ApplyLifecycleThresholds(&report, ThresholdSet{
		ID: "diag",
		Profiles: map[string]ProfileThreshold{
			"lifecycle": {Status: "diagnostic"},
		},
	})
	if report.Status != StatusDiagnostic {
		t.Fatalf("status = %q, want diagnostic", report.Status)
	}
}

func TestLifecycleFailsGate_WhenDiagnosticProfile_ExpectFalse(t *testing.T) {
	t.Parallel()
	report := LifecycleReport{Diagnostics: []LifecycleFailureDiagnostic{{}}}
	set := ThresholdSet{Profiles: map[string]ProfileThreshold{
		"lifecycle": {Status: "diagnostic"},
	}}
	if LifecycleFailsGate(report, set) {
		t.Fatal("diagnostic profile must not gate")
	}
}

func floatPtr(v float64) *float64 { return &v }
