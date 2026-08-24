package evaluation

import (
	"sort"
	"strings"

	"github.com/muonsoft/llm-guard"
)

// SourceLabelSpan is one source annotation interval before mapping.
type SourceLabelSpan struct {
	Label string
	Start int
	End   int
}

var personComponentLabels = map[string]struct{}{
	"FIRST_NAME":  {},
	"LAST_NAME":   {},
	"MIDDLE_NAME": {},
}

var addressComponentLabels = map[string]struct{}{
	"STREET":   {},
	"HOUSE":    {},
	"CITY":     {},
	"REGION":   {},
	"DISTRICT": {},
	"COUNTRY":  {},
}

// ApplyMapping converts source label spans into suite annotations using policy rules.
func ApplyMapping(policy MappingPolicy, spans []SourceLabelSpan, text string) []SuiteAnnotation {
	if len(spans) == 0 {
		return nil
	}
	sorted := append([]SourceLabelSpan(nil), spans...)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Start != sorted[j].Start {
			return sorted[i].Start < sorted[j].Start
		}
		return sorted[i].End < sorted[j].End
	})

	unmappedSet := make(map[string]struct{}, len(policy.UnmappedSensitive))
	for _, label := range policy.UnmappedSensitive {
		unmappedSet[stripBIOPrefix(label)] = struct{}{}
	}

	var out []SuiteAnnotation
	used := make([]bool, len(sorted))

	for i := 0; i < len(sorted); i++ {
		if used[i] {
			continue
		}
		base := stripBIOPrefix(sorted[i].Label)
		if _, ok := personComponentLabels[base]; ok {
			run := collectContiguousRun(sorted, used, i, personComponentLabels, text)
			out = append(out, mapPersonRun(policy, run)...)
			continue
		}
		if _, ok := addressComponentLabels[base]; ok {
			run := collectContiguousRun(sorted, used, i, addressComponentLabels, text)
			out = append(out, mapAddressRun(policy, run)...)
			continue
		}
		used[i] = true
		out = append(out, mapDirectSpan(policy, unmappedSet, sorted[i]))
	}
	return out
}

func compositionGapOK(gap string) bool {
	if gap == "" {
		return true
	}
	for _, r := range gap {
		if r == ' ' || r == '\t' || r == '\n' || r == '\r' || r == ',' || r == '.' || r == ';' {
			continue
		}
		return false
	}
	return true
}

func stripBIOPrefix(label string) string {
	if strings.HasPrefix(label, "B-") || strings.HasPrefix(label, "I-") {
		return label[2:]
	}
	return label
}

func collectContiguousRun(spans []SourceLabelSpan, used []bool, start int, allowed map[string]struct{}, text string) []SourceLabelSpan {
	run := []SourceLabelSpan{spans[start]}
	used[start] = true
	end := spans[start].End
	for j := start + 1; j < len(spans); j++ {
		if used[j] {
			continue
		}
		base := stripBIOPrefix(spans[j].Label)
		if _, ok := allowed[base]; !ok {
			break
		}
		if spans[j].Start > end {
			gap := text[end:spans[j].Start]
			if !compositionGapOK(gap) {
				break
			}
		}
		run = append(run, spans[j])
		used[j] = true
		if spans[j].End > end {
			end = spans[j].End
		}
	}
	return run
}

func mapDirectSpan(policy MappingPolicy, unmappedSet map[string]struct{}, span SourceLabelSpan) SuiteAnnotation {
	base := stripBIOPrefix(span.Label)
	if _, ok := unmappedSet[base]; ok {
		return SuiteAnnotation{
			SourceLabel: span.Label,
			Start:       span.Start,
			End:         span.End,
			Disposition: DispositionUnsupported,
		}
	}
	entity, ok := policy.Direct[base]
	if !ok {
		return SuiteAnnotation{
			SourceLabel: span.Label,
			Start:       span.Start,
			End:         span.End,
			Disposition: DispositionUnsupported,
		}
	}
	if entity == "" {
		return SuiteAnnotation{
			SourceLabel: span.Label,
			Start:       span.Start,
			End:         span.End,
			Disposition: DispositionUnsupported,
		}
	}
	return SuiteAnnotation{
		SourceLabel:  span.Label,
		MappedEntity: entity,
		Start:        span.Start,
		End:          span.End,
		Disposition:  DispositionSupported,
	}
}

func mapPersonRun(policy MappingPolicy, run []SourceLabelSpan) []SuiteAnnotation {
	if len(run) == 0 {
		return nil
	}
	components := make([]string, len(run))
	for i, span := range run {
		components[i] = stripBIOPrefix(span.Label)
	}
	supported := personRunSupported(policy, components)
	var out []SuiteAnnotation
	for _, span := range run {
		out = append(out, SuiteAnnotation{
			SourceLabel: span.Label,
			Start:       span.Start,
			End:         span.End,
			Disposition: DispositionUnsupported,
		})
	}
	if supported {
		start := run[0].Start
		end := run[0].End
		for _, span := range run[1:] {
			if span.Start < start {
				start = span.Start
			}
			if span.End > end {
				end = span.End
			}
		}
		out = append(out, SuiteAnnotation{
			SourceLabel:  "PERSON",
			MappedEntity: string(llmguard.EntityPerson),
			Start:        start,
			End:          end,
			Disposition:  DispositionSupported,
		})
	}
	return out
}

func personRunSupported(policy MappingPolicy, components []string) bool {
	if policy.Person == nil {
		return false
	}
	if len(components) < policy.Person.MinComponents {
		return false
	}
	allowed := make(map[string]struct{}, len(policy.Person.Components))
	for _, c := range policy.Person.Components {
		allowed[c] = struct{}{}
	}
	for _, c := range components {
		if _, ok := allowed[c]; !ok {
			return false
		}
	}
	switch len(components) {
	case 2:
		a, b := components[0], components[1]
		return (a == "FIRST_NAME" && b == "LAST_NAME") || (a == "LAST_NAME" && b == "FIRST_NAME")
	case 3:
		a, b, c := components[0], components[1], components[2]
		return (a == "FIRST_NAME" && b == "MIDDLE_NAME" && c == "LAST_NAME") ||
			(a == "LAST_NAME" && b == "FIRST_NAME" && c == "MIDDLE_NAME")
	default:
		return false
	}
}

func mapAddressRun(policy MappingPolicy, run []SourceLabelSpan) []SuiteAnnotation {
	if len(run) == 0 {
		return nil
	}
	hasStreet := false
	hasHouse := false
	for _, span := range run {
		switch stripBIOPrefix(span.Label) {
		case "STREET":
			hasStreet = true
		case "HOUSE":
			hasHouse = true
		}
	}
	var out []SuiteAnnotation
	for _, span := range run {
		out = append(out, SuiteAnnotation{
			SourceLabel: span.Label,
			Start:       span.Start,
			End:         span.End,
			Disposition: DispositionUnsupported,
		})
	}
	if policy.Address != nil && hasStreet && hasHouse {
		start := run[0].Start
		end := run[0].End
		for _, span := range run[1:] {
			if span.Start < start {
				start = span.Start
			}
			if span.End > end {
				end = span.End
			}
		}
		out = append(out, SuiteAnnotation{
			SourceLabel:  "ADDRESS",
			MappedEntity: string(llmguard.EntityAddress),
			Start:        start,
			End:          end,
			Disposition:  DispositionSupported,
		})
	}
	return out
}
