package evaluation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

// FormatLifecycleMarkdown renders a safe Markdown lifecycle profile report.
func FormatLifecycleMarkdown(report LifecycleReport) string {
	var b strings.Builder
	b.WriteString("# llm-guard lifecycle evaluation report\n\n")
	b.WriteString(fmt.Sprintf("Profile: %s\n", report.Profile))
	b.WriteString(fmt.Sprintf("Suite: %s\n", report.SuiteID))
	b.WriteString(fmt.Sprintf("Mapping version: %s\n", report.MappingVersion))
	if report.ThresholdID != "" {
		b.WriteString(fmt.Sprintf("Thresholds: %s\n", report.ThresholdID))
	}
	b.WriteString(fmt.Sprintf("Sources: %s\n", strings.Join(report.SourceIDs, ", ")))
	b.WriteString(fmt.Sprintf("Cases: %d\n", report.Cases))
	b.WriteString(fmt.Sprintf("Status: %s\n\n", report.Status))
	b.WriteString(fmt.Sprintf("OK: %d\n", report.OK))
	b.WriteString(fmt.Sprintf("Blocked: %d\n", report.Blocked))
	b.WriteString(fmt.Sprintf("Mutation/miss expected: %d\n", report.MutationMiss))
	b.WriteString(fmt.Sprintf("Errors: %d\n", report.Errors))
	if len(report.Diagnostics) > 0 {
		b.WriteString("\nFailure diagnostics:\n\n")
		b.WriteString("| Source record | Expected action | Recipe | Outcome | Detail | Input SHA-256 |\n")
		b.WriteString("| --- | --- | --- | --- | --- | --- |\n")
		for _, row := range report.Diagnostics {
			b.WriteString(fmt.Sprintf(
				"| %s | %s | %s | %s | %s | %s |\n",
				row.SourceRecordID,
				row.ExpectedAction,
				row.ResponseRecipe,
				row.Outcome,
				row.Detail,
				row.InputSHA256,
			))
		}
	}
	return b.String()
}

// FormatLifecycleJSON renders a safe JSON lifecycle profile report.
func FormatLifecycleJSON(report LifecycleReport) ([]byte, error) {
	type failureRow struct {
		SourceRecordID string `json:"source_record_id"`
		ExpectedAction string `json:"expected_action"`
		ResponseRecipe string `json:"response_recipe,omitempty"`
		Outcome        string `json:"outcome"`
		Detail         string `json:"detail"`
		InputSHA256    string `json:"input_sha256"`
	}
	payload := struct {
		Profile        string       `json:"profile"`
		SuiteID        string       `json:"suite_id"`
		MappingVersion string       `json:"mapping_version"`
		ThresholdID    string       `json:"threshold_id,omitempty"`
		SourceIDs      []string     `json:"source_ids"`
		Cases          int          `json:"cases"`
		Status         string       `json:"status"`
		OK             int          `json:"ok"`
		Blocked        int          `json:"blocked"`
		MutationMiss   int          `json:"mutation_miss"`
		RestoreMiss    int          `json:"restore_miss"`
		Errors         int          `json:"errors"`
		Diagnostics    []failureRow `json:"diagnostics,omitempty"`
	}{
		Profile:        report.Profile,
		SuiteID:        report.SuiteID,
		MappingVersion: report.MappingVersion,
		ThresholdID:    report.ThresholdID,
		SourceIDs:      report.SourceIDs,
		Cases:          report.Cases,
		Status:         report.Status,
		OK:             report.OK,
		Blocked:        report.Blocked,
		MutationMiss:   report.MutationMiss,
		RestoreMiss:    report.RestoreMiss,
		Errors:         report.Errors,
	}
	for _, row := range report.Diagnostics {
		payload.Diagnostics = append(payload.Diagnostics, failureRow{
			SourceRecordID: row.SourceRecordID,
			ExpectedAction: row.ExpectedAction,
			ResponseRecipe: row.ResponseRecipe,
			Outcome:        string(row.Outcome),
			Detail:         row.Detail,
			InputSHA256:    row.InputSHA256,
		})
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	if err := enc.Encode(payload); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
