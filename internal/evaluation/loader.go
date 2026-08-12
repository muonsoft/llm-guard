package evaluation

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/muonsoft/llm-guard"
)

var allowedCategories = map[string]struct{}{
	"positive": {},
	"negative": {},
	"mixed":    {},
}

var allowedCaseKeys = map[string]struct{}{
	"schema_version": {},
	"id":             {},
	"category":       {},
	"input":          {},
	"expected":       {},
}

var mvpEntitySet = func() map[llmguard.EntityType]struct{} {
	set := make(map[llmguard.EntityType]struct{}, len(mvpEntityOrder))
	for _, entity := range mvpEntityOrder {
		set[entity] = struct{}{}
	}
	return set
}()

// LoadCases reads and strictly validates a versioned JSONL corpus.
func LoadCases(path string) ([]Case, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var cases []Case
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	lineNo := 0
	seenIDs := make(map[string]struct{})
	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		c, err := decodeStrictCase(line, lineNo)
		if err != nil {
			return nil, err
		}
		if err := validateCase(c, lineNo); err != nil {
			return nil, err
		}
		if _, exists := seenIDs[c.ID]; exists {
			return nil, fmt.Errorf("line %d: duplicate id %q", lineNo, c.ID)
		}
		seenIDs[c.ID] = struct{}{}
		cases = append(cases, c)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(cases) == 0 {
		return nil, fmt.Errorf("corpus %q is empty", path)
	}
	if err := validateEntityCoverage(cases); err != nil {
		return nil, err
	}
	return cases, nil
}

func decodeStrictCase(line string, lineNo int) (Case, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal([]byte(line), &raw); err != nil {
		return Case{}, fmt.Errorf("line %d: %w", lineNo, err)
	}
	for key := range raw {
		if _, ok := allowedCaseKeys[key]; !ok {
			return Case{}, fmt.Errorf("line %d: unknown field %q", lineNo, key)
		}
	}

	dec := json.NewDecoder(strings.NewReader(line))
	dec.DisallowUnknownFields()

	var c Case
	if err := dec.Decode(&c); err != nil {
		return Case{}, fmt.Errorf("line %d: %w", lineNo, err)
	}
	var extra json.RawMessage
	if err := dec.Decode(&extra); err != io.EOF {
		if err == nil {
			return Case{}, fmt.Errorf("line %d: trailing JSON after case object", lineNo)
		}
		return Case{}, fmt.Errorf("line %d: trailing JSON: %w", lineNo, err)
	}
	return c, nil
}

func validateCase(c Case, lineNo int) error {
	if c.SchemaVersion != SchemaVersion {
		return fmt.Errorf("line %d: unsupported schema_version %d", lineNo, c.SchemaVersion)
	}
	if strings.TrimSpace(c.ID) == "" {
		return fmt.Errorf("line %d: id is empty", lineNo)
	}
	if _, ok := allowedCategories[c.Category]; !ok {
		return fmt.Errorf("line %d: unknown category %q", lineNo, c.Category)
	}
	if !utf8.ValidString(c.Input) {
		return fmt.Errorf("line %d: input is not valid UTF-8", lineNo)
	}

	seenSpans := make(map[spanKey]struct{}, len(c.Expected))
	entityKinds := make(map[llmguard.EntityType]struct{})
	for i, span := range c.Expected {
		if span.Entity == "" {
			return fmt.Errorf("line %d expected[%d]: entity is empty", lineNo, i)
		}
		if _, ok := mvpEntitySet[span.Entity]; !ok {
			return fmt.Errorf("line %d expected[%d]: unknown entity %q", lineNo, i, span.Entity)
		}
		if span.Start < 0 || span.End <= span.Start || span.End > len(c.Input) {
			return fmt.Errorf("line %d expected[%d]: invalid span [%d,%d)", lineNo, i, span.Start, span.End)
		}
		if !utf8.RuneStart(c.Input[span.Start]) {
			return fmt.Errorf("line %d expected[%d]: start not on UTF-8 boundary", lineNo, i)
		}
		if span.End < len(c.Input) && !utf8.RuneStart(c.Input[span.End]) {
			return fmt.Errorf("line %d expected[%d]: end not on UTF-8 boundary", lineNo, i)
		}
		key := spanKeyFromExpected(span)
		if _, exists := seenSpans[key]; exists {
			return fmt.Errorf("line %d expected[%d]: duplicate span", lineNo, i)
		}
		seenSpans[key] = struct{}{}
		entityKinds[span.Entity] = struct{}{}
	}

	switch c.Category {
	case "positive":
		if len(c.Expected) == 0 {
			return fmt.Errorf("line %d: positive case requires expected spans", lineNo)
		}
	case "negative":
		if len(c.Expected) != 0 {
			return fmt.Errorf("line %d: negative case must not include expected spans", lineNo)
		}
	case "mixed":
		if len(entityKinds) < 2 {
			return fmt.Errorf("line %d: mixed case requires at least two distinct expected entities", lineNo)
		}
	}
	return nil
}

func validateEntityCoverage(cases []Case) error {
	hasPositive := make(map[llmguard.EntityType]bool, len(mvpEntityOrder))
	hasNegative := make(map[llmguard.EntityType]bool, len(mvpEntityOrder))
	for _, entity := range mvpEntityOrder {
		hasPositive[entity] = false
		hasNegative[entity] = false
	}

	for _, tc := range cases {
		expectedEntities := make(map[llmguard.EntityType]struct{})
		for _, span := range tc.Expected {
			expectedEntities[span.Entity] = struct{}{}
			hasPositive[span.Entity] = true
		}
		for _, entity := range mvpEntityOrder[:] {
			if _, ok := expectedEntities[entity]; !ok {
				hasNegative[entity] = true
			}
		}
	}

	var missing []string
	for _, entity := range mvpEntityOrder[:] {
		if !hasPositive[entity] {
			missing = append(missing, fmt.Sprintf("%s missing positive opportunity", entity))
		}
		if !hasNegative[entity] {
			missing = append(missing, fmt.Sprintf("%s missing negative opportunity", entity))
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("corpus coverage incomplete: %s", strings.Join(missing, "; "))
	}
	return nil
}

// BuildCoverage summarizes positive/negative opportunities per MVP entity.
func BuildCoverage(cases []Case) CoverageReport {
	hasPositive := make(map[llmguard.EntityType]bool, len(mvpEntityOrder))
	hasNegative := make(map[llmguard.EntityType]bool, len(mvpEntityOrder))
	for _, entity := range mvpEntityOrder {
		hasPositive[entity] = false
		hasNegative[entity] = false
	}

	for _, tc := range cases {
		expectedEntities := make(map[llmguard.EntityType]struct{})
		for _, span := range tc.Expected {
			expectedEntities[span.Entity] = struct{}{}
			hasPositive[span.Entity] = true
		}
		for _, entity := range mvpEntityOrder[:] {
			if _, ok := expectedEntities[entity]; !ok {
				hasNegative[entity] = true
			}
		}
	}

	report := CoverageReport{Complete: true}
	for _, entity := range mvpEntityOrder[:] {
		entry := EntityCoverage{
			Entity:      entity,
			HasPositive: hasPositive[entity],
			HasNegative: hasNegative[entity],
		}
		if !entry.HasPositive || !entry.HasNegative {
			report.Complete = false
		}
		report.Entities = append(report.Entities, entry)
	}
	return report
}

type spanKey struct {
	entity llmguard.EntityType
	start  int
	end    int
}

func spanKeyFromFinding(f llmguard.Finding) spanKey {
	return spanKey{entity: f.Entity, start: f.Start, end: f.End}
}

func spanKeyFromExpected(s ExpectedSpan) spanKey {
	return spanKey{entity: s.Entity, start: s.Start, end: s.End}
}
