package evaluation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

// FormatContractMarkdown renders a safe Markdown contract profile report.
func FormatContractMarkdown(report ContractReport) string {
	var b strings.Builder
	b.WriteString("# llm-guard contract evaluation report\n\n")
	b.WriteString(fmt.Sprintf("Profile: %s\n", report.Profile))
	b.WriteString(fmt.Sprintf("Suite: %s\n", report.SuiteID))
	b.WriteString(fmt.Sprintf("Mapping version: %s\n", report.MappingVersion))
	if report.ThresholdID != "" {
		b.WriteString(fmt.Sprintf("Thresholds: %s\n", report.ThresholdID))
	}
	b.WriteString(fmt.Sprintf("Sources: %s\n", strings.Join(report.SourceIDs, ", ")))
	b.WriteString(fmt.Sprintf("Cases: %d\n", report.Cases))
	b.WriteString(fmt.Sprintf("Status: %s\n\n", report.Status))
	b.WriteString("Matching uses Detect → Resolve with exact `(entity, start, end)` UTF-8 byte spans.\n\n")
	b.WriteString("| Entity | TP | FP | FN | Neg cases | FP cases | Precision | Recall | F1 | FPR | FNR |\n")
	b.WriteString("| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |\n")
	for _, row := range report.Entities {
		b.WriteString(fmt.Sprintf(
			"| %s | %d | %d | %d | %d | %d | %.4f | %.4f | %.4f | %.4f | %.4f |\n",
			row.Entity, row.TP, row.FP, row.FN, row.NegativeCases, row.FalsePositiveCases,
			row.Precision, row.Recall, row.F1, row.FPR, row.FNR,
		))
	}
	b.WriteString(fmt.Sprintf("\nAggregate TP=%d FP=%d FN=%d\n", report.Summary.TP, report.Summary.FP, report.Summary.FN))
	if report.Diagnostics.OverlappingPairs > 0 {
		b.WriteString(fmt.Sprintf("Overlap diagnostics (non-gating): %d overlapping-but-not-exact pairs\n", report.Diagnostics.OverlappingPairs))
	}
	if len(report.Diagnostics.Failures) > 0 {
		b.WriteString("\nFailure diagnostics:\n\n")
		b.WriteString("| Source record | Label | Entity | Kind | Gold span | Predicted span | Input SHA-256 |\n")
		b.WriteString("| --- | --- | --- | --- | --- | --- | --- |\n")
		for _, row := range report.Diagnostics.Failures {
			b.WriteString(fmt.Sprintf(
				"| %s | %s | %s | %s | %s | %s | %s |\n",
				row.SourceRecordID,
				row.SourceLabel,
				row.Entity,
				row.Kind,
				formatSpanOffsets(row.GoldStart, row.GoldEnd),
				formatSpanOffsets(row.PredictedStart, row.PredictedEnd),
				row.InputSHA256,
			))
		}
	}
	return b.String()
}

func formatSpanOffsets(start, end *int) string {
	if start == nil || end == nil {
		return ""
	}
	return fmt.Sprintf("[%d,%d)", *start, *end)
}

// FormatContractJSON renders a safe JSON contract profile report.
func FormatContractJSON(report ContractReport) ([]byte, error) {
	type entityRow struct {
		Entity             string  `json:"entity"`
		TP                 int     `json:"tp"`
		FP                 int     `json:"fp"`
		FN                 int     `json:"fn"`
		NegativeCases      int     `json:"negative_cases"`
		FalsePositiveCases int     `json:"false_positive_cases"`
		Precision          float64 `json:"precision"`
		Recall             float64 `json:"recall"`
		F1                 float64 `json:"f1"`
		FPR                float64 `json:"fpr"`
		FNR                float64 `json:"fnr"`
	}
	type failureRow struct {
		SourceRecordID string `json:"source_record_id"`
		SourceLabel    string `json:"source_label,omitempty"`
		Entity         string `json:"entity"`
		Kind           string `json:"kind"`
		GoldStart      *int   `json:"gold_start,omitempty"`
		GoldEnd        *int   `json:"gold_end,omitempty"`
		PredictedStart *int   `json:"predicted_start,omitempty"`
		PredictedEnd   *int   `json:"predicted_end,omitempty"`
		InputSHA256    string `json:"input_sha256"`
	}
	payload := struct {
		Profile        string      `json:"profile"`
		SuiteID        string      `json:"suite_id"`
		MappingVersion string      `json:"mapping_version"`
		ThresholdID    string      `json:"threshold_id,omitempty"`
		SourceIDs      []string    `json:"source_ids"`
		Cases          int         `json:"cases"`
		Status         string      `json:"status"`
		Entities       []entityRow `json:"entities"`
		Summary        struct {
			TP int `json:"tp"`
			FP int `json:"fp"`
			FN int `json:"fn"`
		} `json:"summary"`
		Diagnostics struct {
			OverlappingPairs int          `json:"overlapping_pairs"`
			Failures         []failureRow `json:"failures,omitempty"`
		} `json:"diagnostics"`
	}{
		Profile:        report.Profile,
		SuiteID:        report.SuiteID,
		MappingVersion: report.MappingVersion,
		ThresholdID:    report.ThresholdID,
		SourceIDs:      report.SourceIDs,
		Cases:          report.Cases,
		Status:         report.Status,
	}
	payload.Summary.TP = report.Summary.TP
	payload.Summary.FP = report.Summary.FP
	payload.Summary.FN = report.Summary.FN
	payload.Diagnostics.OverlappingPairs = report.Diagnostics.OverlappingPairs
	for _, row := range report.Diagnostics.Failures {
		payload.Diagnostics.Failures = append(payload.Diagnostics.Failures, failureRow{
			SourceRecordID: row.SourceRecordID,
			SourceLabel:    row.SourceLabel,
			Entity:         row.Entity,
			Kind:           row.Kind,
			GoldStart:      row.GoldStart,
			GoldEnd:        row.GoldEnd,
			PredictedStart: row.PredictedStart,
			PredictedEnd:   row.PredictedEnd,
			InputSHA256:    row.InputSHA256,
		})
	}
	for _, row := range report.Entities {
		payload.Entities = append(payload.Entities, entityRow{
			Entity:             string(row.Entity),
			TP:                 row.TP,
			FP:                 row.FP,
			FN:                 row.FN,
			NegativeCases:      row.NegativeCases,
			FalsePositiveCases: row.FalsePositiveCases,
			Precision:          row.Precision,
			Recall:             row.Recall,
			F1:                 row.F1,
			FPR:                row.FPR,
			FNR:                row.FNR,
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

// FormatExposureMarkdown renders a safe Markdown exposure profile report.
func FormatExposureMarkdown(report ExposureReport) string {
	var b strings.Builder
	b.WriteString("# llm-guard exposure evaluation report\n\n")
	b.WriteString(fmt.Sprintf("Profile: %s\n", report.Profile))
	b.WriteString(fmt.Sprintf("Suite: %s\n", report.SuiteID))
	b.WriteString(fmt.Sprintf("Mapping version: %s\n", report.MappingVersion))
	if report.ThresholdID != "" {
		b.WriteString(fmt.Sprintf("Thresholds: %s\n", report.ThresholdID))
	}
	b.WriteString(fmt.Sprintf("Sources: %s\n", strings.Join(report.SourceIDs, ", ")))
	b.WriteString(fmt.Sprintf("Cases: %d\n", report.Cases))
	b.WriteString(fmt.Sprintf("Status: %s\n\n", report.Status))
	b.WriteString(fmt.Sprintf("Sensitive bytes: %d\n", report.Summary.SensitiveBytes))
	b.WriteString(fmt.Sprintf("Covered sensitive bytes: %d\n", report.Summary.CoveredSensitiveBytes))
	b.WriteString(fmt.Sprintf("Leaked sensitive bytes: %d\n", report.Summary.LeakedSensitiveBytes))
	b.WriteString(fmt.Sprintf("Overmatched bytes: %d\n", report.Summary.OvermatchedBytes))
	b.WriteString(fmt.Sprintf("Byte coverage: %.4f\n\n", report.Summary.ByteCoverage))
	if len(report.ByLabel) > 0 {
		b.WriteString("| Source label | Mapped entity | Disposition | Spans | Fully covered | Sensitive bytes | Covered | Leaked | Overmatched |\n")
		b.WriteString("| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | ---: |\n")
		for _, row := range report.ByLabel {
			b.WriteString(fmt.Sprintf(
				"| %s | %s | %s | %d | %d | %d | %d | %d | %d |\n",
				row.SourceLabel, row.MappedEntity, row.Disposition,
				row.SpanCount, row.FullyCoveredSpanCount,
				row.SensitiveBytes, row.CoveredSensitiveBytes,
				row.LeakedSensitiveBytes, row.OvermatchedBytes,
			))
		}
		b.WriteString("\n")
	}
	if len(report.Ignored) > 0 {
		b.WriteString("Ignored annotations:\n\n")
		b.WriteString("| Source label | Reason | Count |\n")
		b.WriteString("| --- | --- | ---: |\n")
		for _, row := range report.Ignored {
			b.WriteString(fmt.Sprintf("| %s | %s | %d |\n", row.SourceLabel, row.Reason, row.Count))
		}
	}
	return b.String()
}

// FormatExposureJSON renders a safe JSON exposure profile report.
func FormatExposureJSON(report ExposureReport) ([]byte, error) {
	type labelRow struct {
		SourceLabel           string `json:"source_label"`
		MappedEntity          string `json:"mapped_entity"`
		Disposition           string `json:"disposition"`
		SpanCount             int    `json:"span_count"`
		FullyCoveredSpanCount int    `json:"fully_covered_span_count"`
		SensitiveBytes        int    `json:"sensitive_bytes"`
		CoveredSensitiveBytes int    `json:"covered_sensitive_bytes"`
		LeakedSensitiveBytes  int    `json:"leaked_sensitive_bytes"`
		OvermatchedBytes      int    `json:"overmatched_bytes"`
	}
	type ignoredRow struct {
		SourceLabel string `json:"source_label"`
		Reason      string `json:"reason"`
		Count       int    `json:"count"`
	}
	payload := struct {
		Profile        string   `json:"profile"`
		SuiteID        string   `json:"suite_id"`
		MappingVersion string   `json:"mapping_version"`
		ThresholdID    string   `json:"threshold_id,omitempty"`
		SourceIDs      []string `json:"source_ids"`
		Cases          int      `json:"cases"`
		Status         string   `json:"status"`
		Summary        struct {
			SensitiveBytes        int     `json:"sensitive_bytes"`
			CoveredSensitiveBytes int     `json:"covered_sensitive_bytes"`
			LeakedSensitiveBytes  int     `json:"leaked_sensitive_bytes"`
			OvermatchedBytes      int     `json:"overmatched_bytes"`
			ByteCoverage          float64 `json:"byte_coverage"`
		} `json:"summary"`
		ByLabel []labelRow   `json:"by_label"`
		Ignored []ignoredRow `json:"ignored"`
	}{
		Profile:        report.Profile,
		SuiteID:        report.SuiteID,
		MappingVersion: report.MappingVersion,
		ThresholdID:    report.ThresholdID,
		SourceIDs:      report.SourceIDs,
		Cases:          report.Cases,
		Status:         report.Status,
	}
	payload.Summary.SensitiveBytes = report.Summary.SensitiveBytes
	payload.Summary.CoveredSensitiveBytes = report.Summary.CoveredSensitiveBytes
	payload.Summary.LeakedSensitiveBytes = report.Summary.LeakedSensitiveBytes
	payload.Summary.OvermatchedBytes = report.Summary.OvermatchedBytes
	payload.Summary.ByteCoverage = report.Summary.ByteCoverage
	for _, row := range report.ByLabel {
		payload.ByLabel = append(payload.ByLabel, labelRow{
			SourceLabel:           row.SourceLabel,
			MappedEntity:          row.MappedEntity,
			Disposition:           row.Disposition,
			SpanCount:             row.SpanCount,
			FullyCoveredSpanCount: row.FullyCoveredSpanCount,
			SensitiveBytes:        row.SensitiveBytes,
			CoveredSensitiveBytes: row.CoveredSensitiveBytes,
			LeakedSensitiveBytes:  row.LeakedSensitiveBytes,
			OvermatchedBytes:      row.OvermatchedBytes,
		})
	}
	for _, row := range report.Ignored {
		payload.Ignored = append(payload.Ignored, ignoredRow{
			SourceLabel: row.SourceLabel,
			Reason:      row.Reason,
			Count:       row.Count,
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
