package evaluation

import (
	"encoding/json"
	"testing"
)

func TestFormatContractJSON_WhenMultipleSources_ExpectBySourceMetrics(t *testing.T) {
	t.Parallel()
	report := ContractReport{
		Profile: "contract",
		SuiteID: "s",
		Status:  StatusPass,
		sourceSummaries: sourceContractSummaries{
			"src-b": {TP: 2, FP: 1, FN: 0},
			"src-a": {TP: 1, FP: 0, FN: 1},
		},
	}
	out, err := FormatContractJSON(report)
	if err != nil {
		t.Fatalf("FormatContractJSON: %v", err)
	}
	var payload struct {
		BySource []struct {
			SourceID string `json:"source_id"`
			TP       int    `json:"tp"`
			FP       int    `json:"fp"`
			FN       int    `json:"fn"`
		} `json:"by_source"`
	}
	if err := json.Unmarshal(out, &payload); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(payload.BySource) != 2 {
		t.Fatalf("by_source len = %d, want 2", len(payload.BySource))
	}
	if payload.BySource[0].SourceID != "src-a" || payload.BySource[1].SourceID != "src-b" {
		t.Fatalf("by_source order = %+v, want src-a then src-b", payload.BySource)
	}
	if payload.BySource[0].TP != 1 || payload.BySource[0].FN != 1 {
		t.Fatalf("src-a metrics = %+v", payload.BySource[0])
	}
	if payload.BySource[1].TP != 2 || payload.BySource[1].FP != 1 {
		t.Fatalf("src-b metrics = %+v", payload.BySource[1])
	}
}

func TestFormatExposureJSON_WhenMultipleSources_ExpectBySourceMetrics(t *testing.T) {
	t.Parallel()
	report := ExposureReport{
		Profile: "exposure",
		SuiteID: "s",
		Status:  StatusDiagnostic,
		sourceSummaries: sourceExposureSummaries{
			"b": {SensitiveBytes: 10, CoveredSensitiveBytes: 8, LeakedSensitiveBytes: 2, OvermatchedBytes: 1, ByteCoverage: 0.8},
			"a": {SensitiveBytes: 5, CoveredSensitiveBytes: 5, LeakedSensitiveBytes: 0, OvermatchedBytes: 0, ByteCoverage: 1.0},
		},
	}
	out, err := FormatExposureJSON(report)
	if err != nil {
		t.Fatalf("FormatExposureJSON: %v", err)
	}
	var payload struct {
		BySource []struct {
			SourceID       string  `json:"source_id"`
			SensitiveBytes int     `json:"sensitive_bytes"`
			ByteCoverage   float64 `json:"byte_coverage"`
		} `json:"by_source"`
	}
	if err := json.Unmarshal(out, &payload); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(payload.BySource) != 2 {
		t.Fatalf("by_source len = %d, want 2", len(payload.BySource))
	}
	if payload.BySource[0].SourceID != "a" || payload.BySource[1].SourceID != "b" {
		t.Fatalf("by_source order = %+v", payload.BySource)
	}
	if payload.BySource[0].SensitiveBytes != 5 || payload.BySource[1].SensitiveBytes != 10 {
		t.Fatalf("by_source bytes = %+v", payload.BySource)
	}
}
