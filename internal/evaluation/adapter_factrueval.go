package evaluation

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const factRuEvalSourceID = "factrueval-2016"

// FactRuEvalAdapter converts pinned FactRuEval document layers into suite records.
type FactRuEvalAdapter struct {
	Policy MappingPolicy
}

type factRuEvalSpan struct {
	id         int
	spanType   string
	charStart  int
	charLength int
}

type factRuEvalObject struct {
	id      int
	objType string
	spanIDs []int
}

// AdaptDocument parses one aligned document into a suite record.
func (a FactRuEvalAdapter) AdaptDocument(split, stem, text string, tokensPath, spansPath, objectsPath string) (SuiteRecord, error) {
	recordID := split + "/" + stem
	spansByID, err := parseFactRuEvalSpans(spansPath)
	if err != nil {
		return SuiteRecord{}, fmt.Errorf("record %s: %w", recordID, err)
	}
	objects, err := parseFactRuEvalObjects(objectsPath)
	if err != nil {
		return SuiteRecord{}, fmt.Errorf("record %s: %w", recordID, err)
	}
	_ = tokensPath

	var annotations []SuiteAnnotation
	for _, obj := range objects {
		switch obj.objType {
		case "Person":
			annotations = append(annotations, mapFactRuEvalPersonObject(a.Policy, obj, spansByID, text)...)
		case "Org", "LocOrg", "Location":
			annotations = append(annotations, mapFactRuEvalExposureObject(a.Policy, obj, spansByID, text)...)
		}
	}
	return buildSuiteRecord(factRuEvalSourceID, recordID, a.Policy.Version, text, annotations), nil
}

func mapFactRuEvalPersonObject(policy MappingPolicy, obj factRuEvalObject, spansByID map[int]factRuEvalSpan, text string) []SuiteAnnotation {
	var sourceSpans []SourceLabelSpan
	for _, id := range obj.spanIDs {
		span, ok := spansByID[id]
		if !ok {
			continue
		}
		label := factRuEvalSpanLabel(span.spanType)
		if label == "" {
			if span.spanType == "nickname" {
				start, end, ok := RuneIntervalToUTF8(text, span.charStart, span.charStart+span.charLength)
				if ok {
					sourceSpans = append(sourceSpans, SourceLabelSpan{Label: "nickname", Start: start, End: end})
				}
			}
			continue
		}
		start, end, ok := RuneIntervalToUTF8(text, span.charStart, span.charStart+span.charLength)
		if !ok {
			continue
		}
		sourceSpans = append(sourceSpans, SourceLabelSpan{Label: label, Start: start, End: end})
	}
	if len(sourceSpans) == 0 {
		return nil
	}
	sortSourceLabelSpans(sourceSpans)
	if !personMentionContiguous(text, sourceSpans) {
		return ApplyMapping(policy, sourceSpans, text)
	}
	return ApplyMapping(policy, sourceSpans, text)
}

func mapFactRuEvalExposureObject(policy MappingPolicy, obj factRuEvalObject, spansByID map[int]factRuEvalSpan, text string) []SuiteAnnotation {
	var sourceSpans []SourceLabelSpan
	for _, id := range obj.spanIDs {
		span, ok := spansByID[id]
		if !ok {
			continue
		}
		start, end, ok := RuneIntervalToUTF8(text, span.charStart, span.charStart+span.charLength)
		if !ok {
			continue
		}
		sourceSpans = append(sourceSpans, SourceLabelSpan{Label: obj.objType, Start: start, End: end})
	}
	if len(sourceSpans) == 0 {
		return nil
	}
	sortSourceLabelSpans(sourceSpans)
	return ApplyMapping(policy, sourceSpans, text)
}

func sortSourceLabelSpans(spans []SourceLabelSpan) {
	sort.Slice(spans, func(i, j int) bool {
		if spans[i].Start != spans[j].Start {
			return spans[i].Start < spans[j].Start
		}
		return spans[i].End < spans[j].End
	})
}

func personMentionContiguous(text string, spans []SourceLabelSpan) bool {
	for i := 1; i < len(spans); i++ {
		gap := text[spans[i-1].End:spans[i].Start]
		if strings.TrimSpace(gap) != "" && gap != "" {
			return false
		}
	}
	return true
}

func factRuEvalSpanLabel(spanType string) string {
	switch spanType {
	case "name":
		return "FIRST_NAME"
	case "surname":
		return "LAST_NAME"
	case "patronymic":
		return "MIDDLE_NAME"
	case "nickname":
		return "nickname"
	default:
		return ""
	}
}

func parseFactRuEvalSpans(path string) (map[int]factRuEvalSpan, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	out := make(map[int]factRuEvalSpan)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}
		id, err := strconv.Atoi(fields[0])
		if err != nil {
			return nil, fmt.Errorf("spans: invalid id in %s", filepath.Base(path))
		}
		charStart, err := strconv.Atoi(fields[2])
		if err != nil {
			return nil, err
		}
		charLength, err := strconv.Atoi(fields[3])
		if err != nil {
			return nil, err
		}
		out[id] = factRuEvalSpan{
			id:         id,
			spanType:   fields[1],
			charStart:  charStart,
			charLength: charLength,
		}
	}
	return out, scanner.Err()
}

func parseFactRuEvalObjects(path string) ([]factRuEvalObject, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var out []factRuEvalObject
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		id, err := strconv.Atoi(fields[0])
		if err != nil {
			return nil, err
		}
		obj := factRuEvalObject{
			id:      id,
			objType: fields[1],
		}
		for _, raw := range fields[2:] {
			spanID, err := strconv.Atoi(raw)
			if err != nil {
				return nil, err
			}
			obj.spanIDs = append(obj.spanIDs, spanID)
		}
		out = append(out, obj)
	}
	return out, scanner.Err()
}

// NormalizeFactRuEvalTree walks devset/testset trees and emits normalized records.
func NormalizeFactRuEvalTree(root string, policy MappingPolicy, emit func(SuiteRecord) error) error {
	adapter := FactRuEvalAdapter{Policy: policy}
	for _, split := range []string{"devset", "testset"} {
		splitRoot := filepath.Join(root, split)
		entries, err := os.ReadDir(splitRoot)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".txt") {
				continue
			}
			stem := strings.TrimSuffix(entry.Name(), ".txt")
			if split == "testset" && stem == "list" {
				continue
			}
			textBytes, err := os.ReadFile(filepath.Join(splitRoot, entry.Name()))
			if err != nil {
				return err
			}
			text := string(textBytes)
			rec, err := adapter.AdaptDocument(
				split,
				stem,
				text,
				filepath.Join(splitRoot, stem+".tokens"),
				filepath.Join(splitRoot, stem+".spans"),
				filepath.Join(splitRoot, stem+".objects"),
			)
			if err != nil {
				return err
			}
			if err := emit(rec); err != nil {
				return err
			}
		}
	}
	return nil
}

// ReadFactRuEvalText loads document text from split/stem paths.
func ReadFactRuEvalText(root, split, stem string) (string, error) {
	data, err := os.ReadFile(filepath.Join(root, split, stem+".txt"))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// CanonicalDigestFiles computes the adapter-input digest for FactRuEval cache verification.
func CanonicalDigestFiles(paths []string, open func(string) (io.ReadCloser, error)) (string, error) {
	return canonicalPathDigest(paths, open)
}
