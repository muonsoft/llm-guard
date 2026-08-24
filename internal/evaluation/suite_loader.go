package evaluation

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/muonsoft/llm-guard"
)

var allowedSuiteRecordKeys = map[string]struct{}{
	"schema_version":    {},
	"suite_id":          {},
	"source_id":         {},
	"source_record_id":  {},
	"mapping_version":   {},
	"input":             {},
	"input_sha256":      {},
	"annotations":       {},
	"declared_entities": {},
	"lifecycle":         {},
}

var allowedAnnotationKeys = map[string]struct{}{
	"source_label":  {},
	"mapped_entity": {},
	"start":         {},
	"end":           {},
	"disposition":   {},
	"reason":        {},
}

var allowedLifecycleKeys = map[string]struct{}{
	"expected_action": {},
	"response_recipe": {},
}

var allowedDispositions = map[string]struct{}{
	DispositionSupported:   {},
	DispositionUnsupported: {},
	DispositionIgnored:     {},
}

var allowedLifecycleActions = map[string]struct{}{
	"mask":  {},
	"block": {},
	"allow": {},
}

var allowedResponseRecipes = map[string]struct{}{
	"identity":           {},
	"mutate_placeholder": {},
	"delete_placeholder": {},
	"collision":          {},
}

// LoadSuite reads and strictly validates a schema v2 JSONL suite.
func LoadSuite(path string) (Suite, error) {
	f, err := os.Open(path)
	if err != nil {
		return Suite{}, err
	}
	defer f.Close()

	var records []SuiteRecord
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	lineNo := 0
	seenRecordIDs := make(map[string]struct{})
	var suiteID, mappingVersion string
	sourceIDSet := make(map[string]struct{})
	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		rec, err := decodeStrictSuiteRecord(line, lineNo)
		if err != nil {
			return Suite{}, err
		}
		if err := validateSuiteRecord(rec, lineNo, seenRecordIDs); err != nil {
			return Suite{}, err
		}
		if suiteID == "" {
			suiteID = rec.SuiteID
			mappingVersion = rec.MappingVersion
		} else {
			if rec.SuiteID != suiteID {
				return Suite{}, fmt.Errorf("line %d: suite_id %q does not match file suite_id %q", lineNo, rec.SuiteID, suiteID)
			}
			if rec.MappingVersion != mappingVersion {
				return Suite{}, fmt.Errorf("line %d: mapping_version %q does not match file mapping_version %q", lineNo, rec.MappingVersion, mappingVersion)
			}
		}
		sourceIDSet[rec.SourceID] = struct{}{}
		records = append(records, rec)
	}
	if err := scanner.Err(); err != nil {
		return Suite{}, err
	}
	if len(records) == 0 {
		return Suite{}, fmt.Errorf("suite %q is empty", path)
	}
	sourceIDs := make([]string, 0, len(sourceIDSet))
	for id := range sourceIDSet {
		sourceIDs = append(sourceIDs, id)
	}
	sort.Strings(sourceIDs)
	scope, err := computeSuiteScope(records)
	if err != nil {
		return Suite{}, err
	}
	return Suite{
		Records:        records,
		SuiteID:        suiteID,
		MappingVersion: mappingVersion,
		SourceIDs:      sourceIDs,
		Scope:          scope,
	}, nil
}

func decodeStrictSuiteRecord(line string, lineNo int) (SuiteRecord, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal([]byte(line), &raw); err != nil {
		return SuiteRecord{}, fmt.Errorf("line %d: %w", lineNo, err)
	}
	for key := range raw {
		if _, ok := allowedSuiteRecordKeys[key]; !ok {
			return SuiteRecord{}, fmt.Errorf("line %d: unknown field %q", lineNo, key)
		}
	}

	dec := json.NewDecoder(strings.NewReader(line))
	dec.DisallowUnknownFields()

	var rec SuiteRecord
	if err := dec.Decode(&rec); err != nil {
		return SuiteRecord{}, fmt.Errorf("line %d: %w", lineNo, err)
	}
	var extra json.RawMessage
	if err := dec.Decode(&extra); err != io.EOF {
		if err == nil {
			return SuiteRecord{}, fmt.Errorf("line %d: trailing JSON after suite record", lineNo)
		}
		return SuiteRecord{}, fmt.Errorf("line %d: trailing JSON: %w", lineNo, err)
	}
	return rec, nil
}

func validateSuiteRecord(rec SuiteRecord, lineNo int, seenRecordIDs map[string]struct{}) error {
	if rec.SchemaVersion != SuiteSchemaVersion {
		return fmt.Errorf("line %d: unsupported schema_version %d", lineNo, rec.SchemaVersion)
	}
	if strings.TrimSpace(rec.SuiteID) == "" {
		return fmt.Errorf("line %d: suite_id is empty", lineNo)
	}
	if strings.TrimSpace(rec.SourceID) == "" {
		return fmt.Errorf("line %d: source_id is empty", lineNo)
	}
	if strings.TrimSpace(rec.SourceRecordID) == "" {
		return fmt.Errorf("line %d: source_record_id is empty", lineNo)
	}
	if strings.TrimSpace(rec.MappingVersion) == "" {
		return fmt.Errorf("line %d: mapping_version is empty", lineNo)
	}
	if !utf8.ValidString(rec.Input) {
		return fmt.Errorf("line %d: input is not valid UTF-8", lineNo)
	}
	if err := validateInputSHA256(rec, lineNo); err != nil {
		return err
	}
	if _, exists := seenRecordIDs[rec.SourceRecordID]; exists {
		return fmt.Errorf("line %d: duplicate source_record_id %q", lineNo, rec.SourceRecordID)
	}
	seenRecordIDs[rec.SourceRecordID] = struct{}{}

	seenSpans := make(map[annotationSpanKey]struct{}, len(rec.Annotations))
	for i, ann := range rec.Annotations {
		if err := validateAnnotation(ann, rec, lineNo, i, seenSpans); err != nil {
			return err
		}
	}
	if rec.Lifecycle != nil {
		if err := validateLifecycle(*rec.Lifecycle, lineNo); err != nil {
			return err
		}
	}
	for _, entity := range rec.DeclaredEntities {
		if entity == "" {
			return fmt.Errorf("line %d: declared_entities contains empty entity", lineNo)
		}
		if _, ok := mvpEntitySet[llmguard.EntityType(entity)]; !ok {
			return fmt.Errorf("line %d: declared_entities unknown entity %q", lineNo, entity)
		}
	}
	return nil
}

func validateInputSHA256(rec SuiteRecord, lineNo int) error {
	if strings.TrimSpace(rec.InputSHA256) == "" {
		return fmt.Errorf("line %d: input_sha256 is empty", lineNo)
	}
	sum := sha256.Sum256([]byte(rec.Input))
	expected := hex.EncodeToString(sum[:])
	if rec.InputSHA256 != expected {
		return fmt.Errorf("line %d record %q: input_sha256 mismatch", lineNo, rec.SourceRecordID)
	}
	for _, ch := range rec.InputSHA256 {
		if (ch < '0' || ch > '9') && (ch < 'a' || ch > 'f') {
			return fmt.Errorf("line %d record %q: input_sha256 must be lowercase hex", lineNo, rec.SourceRecordID)
		}
	}
	return nil
}

type annotationSpanKey struct {
	label string
	start int
	end   int
}

func validateAnnotation(ann SuiteAnnotation, rec SuiteRecord, lineNo, idx int, seenSpans map[annotationSpanKey]struct{}) error {
	if strings.TrimSpace(ann.SourceLabel) == "" {
		return fmt.Errorf("line %d annotation[%d] record %q: source_label is empty", lineNo, idx, rec.SourceRecordID)
	}
	if _, ok := allowedDispositions[ann.Disposition]; !ok {
		return fmt.Errorf("line %d annotation[%d] record %q: unknown disposition %q", lineNo, idx, rec.SourceRecordID, ann.Disposition)
	}
	switch ann.Disposition {
	case DispositionIgnored:
		if strings.TrimSpace(ann.Reason) == "" {
			return fmt.Errorf("line %d annotation[%d] record %q: ignored disposition requires reason", lineNo, idx, rec.SourceRecordID)
		}
	case DispositionSupported:
		if ann.MappedEntity == "" {
			return fmt.Errorf("line %d annotation[%d] record %q: supported disposition requires mapped_entity", lineNo, idx, rec.SourceRecordID)
		}
		if _, ok := mvpEntitySet[llmguard.EntityType(ann.MappedEntity)]; !ok {
			return fmt.Errorf("line %d annotation[%d] record %q: unknown mapped_entity %q", lineNo, idx, rec.SourceRecordID, ann.MappedEntity)
		}
		if ann.Reason != "" {
			return fmt.Errorf("line %d annotation[%d] record %q: reason must be empty for non-ignored disposition", lineNo, idx, rec.SourceRecordID)
		}
	case DispositionUnsupported:
		if ann.MappedEntity != "" {
			if _, ok := mvpEntitySet[llmguard.EntityType(ann.MappedEntity)]; !ok {
				return fmt.Errorf("line %d annotation[%d] record %q: unknown mapped_entity %q", lineNo, idx, rec.SourceRecordID, ann.MappedEntity)
			}
		}
		if ann.Reason != "" {
			return fmt.Errorf("line %d annotation[%d] record %q: reason must be empty for non-ignored disposition", lineNo, idx, rec.SourceRecordID)
		}
	}
	if ann.Start < 0 || ann.End <= ann.Start || ann.End > len(rec.Input) {
		return fmt.Errorf("line %d annotation[%d] record %q label %q: invalid span [%d,%d)", lineNo, idx, rec.SourceRecordID, ann.SourceLabel, ann.Start, ann.End)
	}
	if !utf8.RuneStart(rec.Input[ann.Start]) {
		return fmt.Errorf("line %d annotation[%d] record %q label %q: start not on UTF-8 boundary", lineNo, idx, rec.SourceRecordID, ann.SourceLabel)
	}
	if ann.End < len(rec.Input) && !utf8.RuneStart(rec.Input[ann.End]) {
		return fmt.Errorf("line %d annotation[%d] record %q label %q: end not on UTF-8 boundary", lineNo, idx, rec.SourceRecordID, ann.SourceLabel)
	}
	key := annotationSpanKey{label: ann.SourceLabel, start: ann.Start, end: ann.End}
	if _, exists := seenSpans[key]; exists {
		return fmt.Errorf("line %d annotation[%d] record %q: duplicate span for label %q [%d,%d)", lineNo, idx, rec.SourceRecordID, ann.SourceLabel, ann.Start, ann.End)
	}
	seenSpans[key] = struct{}{}
	return nil
}

func validateLifecycle(lc SuiteLifecycle, lineNo int) error {
	if strings.TrimSpace(lc.ExpectedAction) == "" {
		return fmt.Errorf("line %d lifecycle: expected_action is empty", lineNo)
	}
	if _, ok := allowedLifecycleActions[lc.ExpectedAction]; !ok {
		return fmt.Errorf("line %d lifecycle: unknown expected_action %q", lineNo, lc.ExpectedAction)
	}
	if lc.ResponseRecipe != "" {
		if _, ok := allowedResponseRecipes[lc.ResponseRecipe]; !ok {
			return fmt.Errorf("line %d lifecycle: unknown response_recipe %q", lineNo, lc.ResponseRecipe)
		}
	}
	return nil
}

func computeSuiteScope(records []SuiteRecord) ([]llmguard.EntityType, error) {
	scopeSet := make(map[llmguard.EntityType]struct{})
	var declared []string
	for _, rec := range records {
		if len(rec.DeclaredEntities) > 0 {
			if declared == nil {
				declared = append([]string(nil), rec.DeclaredEntities...)
			} else if !stringSlicesEqual(declared, rec.DeclaredEntities) {
				return nil, fmt.Errorf("suite declared_entities mismatch across records")
			}
		}
	}
	if declared != nil {
		for _, entity := range declared {
			scopeSet[llmguard.EntityType(entity)] = struct{}{}
		}
	} else {
		for _, rec := range records {
			for _, ann := range rec.Annotations {
				if ann.Disposition == DispositionSupported && ann.MappedEntity != "" {
					scopeSet[llmguard.EntityType(ann.MappedEntity)] = struct{}{}
				}
			}
		}
	}
	scope := make([]llmguard.EntityType, 0, len(scopeSet))
	for entity := range scopeSet {
		scope = append(scope, entity)
	}
	sort.Slice(scope, func(i, j int) bool { return scope[i] < scope[j] })
	return scope, nil
}

func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func recordScope(rec SuiteRecord, suiteScope []llmguard.EntityType) map[llmguard.EntityType]struct{} {
	if len(rec.DeclaredEntities) > 0 {
		out := make(map[llmguard.EntityType]struct{}, len(rec.DeclaredEntities))
		for _, entity := range rec.DeclaredEntities {
			out[llmguard.EntityType(entity)] = struct{}{}
		}
		return out
	}
	out := make(map[llmguard.EntityType]struct{}, len(suiteScope))
	for _, entity := range suiteScope {
		out[entity] = struct{}{}
	}
	return out
}

func entityInScope(entity llmguard.EntityType, scope map[llmguard.EntityType]struct{}) bool {
	_, ok := scope[entity]
	return ok
}
