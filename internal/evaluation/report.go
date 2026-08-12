package evaluation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

// FormatMarkdown renders a deterministic Markdown quality report.
func FormatMarkdown(report Report) string {
	var b strings.Builder
	b.WriteString("# llm-guard MVP evaluation report\n\n")
	b.WriteString(fmt.Sprintf("Cases: %d\n\n", report.Cases))
	b.WriteString("Matching uses Detect → Resolve with exact `(entity, start, end)` UTF-8 byte spans.\n\n")
	b.WriteString("Formulas: precision = TP/(TP+FP), recall = TP/(TP+FN), F1 = 2PR/(P+R), ")
	b.WriteString("FPR = false_positive_cases/negative_cases, FNR = FN/(TP+FN). Zero denominators yield 0.\n\n")
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
	b.WriteString(fmt.Sprintf("Coverage complete: %t\n", report.Coverage.Complete))
	if !report.Coverage.Complete {
		b.WriteString("\n| Entity | Positive | Negative |\n")
		b.WriteString("| --- | :---: | :---: |\n")
		for _, row := range report.Coverage.Entities {
			b.WriteString(fmt.Sprintf("| %s | %t | %t |\n", row.Entity, row.HasPositive, row.HasNegative))
		}
	}
	return b.String()
}

// FormatJSON renders a deterministic JSON quality report.
func FormatJSON(report Report) ([]byte, error) {
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
	type coverageEntityRow struct {
		Entity      string `json:"entity"`
		HasPositive bool   `json:"has_positive"`
		HasNegative bool   `json:"has_negative"`
	}
	type coveragePayload struct {
		Complete bool                `json:"complete"`
		Entities []coverageEntityRow `json:"entities"`
	}
	payload := struct {
		Cases    int         `json:"cases"`
		Entities []entityRow `json:"entities"`
		Summary  struct {
			TP int `json:"tp"`
			FP int `json:"fp"`
			FN int `json:"fn"`
		} `json:"summary"`
		Coverage coveragePayload `json:"coverage"`
	}{
		Cases: report.Cases,
		Coverage: coveragePayload{
			Complete: report.Coverage.Complete,
		},
	}
	payload.Summary.TP = report.Summary.TP
	payload.Summary.FP = report.Summary.FP
	payload.Summary.FN = report.Summary.FN
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
	for _, row := range report.Coverage.Entities {
		payload.Coverage.Entities = append(payload.Coverage.Entities, coverageEntityRow{
			Entity:      string(row.Entity),
			HasPositive: row.HasPositive,
			HasNegative: row.HasNegative,
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
