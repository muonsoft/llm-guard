package evaluation

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

const redMadRobotSourceID = "redmadrobot-pii-benchmark"

// RedMadRobotAdapter converts pinned CSV rows into normalized suite records.
type RedMadRobotAdapter struct {
	Policy MappingPolicy
}

// AdaptCSVRecord parses one CSV row into a suite record or an ignored stub.
func (a RedMadRobotAdapter) AdaptCSVRecord(recordID string, text, tokensJSON, tagsJSON string) (SuiteRecord, error) {
	var tokens []string
	if err := json.Unmarshal([]byte(tokensJSON), &tokens); err != nil {
		return SuiteRecord{}, fmt.Errorf("record %s: tokens json: %w", recordID, err)
	}
	var tags []string
	if err := json.Unmarshal([]byte(tagsJSON), &tags); err != nil {
		return SuiteRecord{}, fmt.Errorf("record %s: ner_tags json: %w", recordID, err)
	}
	if len(tokens) != len(tags) {
		return ignoredRecord(redMadRobotSourceID, recordID, a.Policy.Version, text, "adapter_token_tag_length_mismatch"), nil
	}

	spans, reason := alignRedMadRobotTokens(text, tokens, tags)
	if reason != "" {
		return ignoredRecord(redMadRobotSourceID, recordID, a.Policy.Version, text, reason), nil
	}
	annotations := ApplyMapping(a.Policy, spans, text)
	return buildSuiteRecord(redMadRobotSourceID, recordID, a.Policy.Version, text, annotations), nil
}

func alignRedMadRobotTokens(text string, tokens, tags []string) ([]SourceLabelSpan, string) {
	pos := 0
	var spans []SourceLabelSpan

	for i := 0; i < len(tokens); {
		localStart, localEnd, _, ok := LocateTokenInText(text[pos:], tokens[i])
		if !ok {
			return nil, "adapter_token_not_found_in_text"
		}
		absStart := pos + localStart
		absEnd := pos + localEnd

		tag := tags[i]
		if tag != "O" {
			label := stripBIOPrefix(tag)
			runStart := absStart
			runEnd := absEnd
			nextPos := absEnd
			j := i + 1
			for j < len(tokens) && strings.HasPrefix(tags[j], "I-") && stripBIOPrefix(tags[j]) == label {
				_, le, _, found := LocateTokenInText(text[nextPos:], tokens[j])
				if !found {
					return nil, "adapter_token_not_found_in_text"
				}
				runEnd = nextPos + le
				nextPos = runEnd
				j++
			}
			spans = append(spans, SourceLabelSpan{
				Label: label,
				Start: runStart,
				End:   runEnd,
			})
			pos = nextPos
			i = j
			continue
		}

		pos = absEnd
		i++
	}
	return spans, ""
}

// NormalizeRedMadRobotCSV reads cache CSV and writes normalized records.
func NormalizeRedMadRobotCSV(r io.Reader, policy MappingPolicy, emit func(SuiteRecord) error) error {
	adapter := RedMadRobotAdapter{Policy: policy}
	reader := csv.NewReader(r)
	reader.FieldsPerRecord = -1
	header, err := reader.Read()
	if err != nil {
		return err
	}
	colText, colTokens, colTags, err := redMadRobotCSVColumns(header)
	if err != nil {
		return err
	}
	rowIdx := 0
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if len(row) <= colTags {
			return fmt.Errorf("row %d: insufficient columns", rowIdx)
		}
		recordID := fmt.Sprintf("%d", rowIdx)
		rec, err := adapter.AdaptCSVRecord(recordID, row[colText], row[colTokens], row[colTags])
		if err != nil {
			return err
		}
		if err := emit(rec); err != nil {
			return err
		}
		rowIdx++
	}
	return nil
}

func redMadRobotCSVColumns(header []string) (text, tokens, tags int, err error) {
	text, tokens, tags = -1, -1, -1
	for i, col := range header {
		switch strings.TrimSpace(col) {
		case "text":
			text = i
		case "tokens":
			tokens = i
		case "ner_tags":
			tags = i
		}
	}
	if text < 0 || tokens < 0 || tags < 0 {
		return 0, 0, 0, fmt.Errorf("csv header missing required columns")
	}
	return text, tokens, tags, nil
}
