package llmguard

import (
	"sort"
)

type resolveCandidate struct {
	Finding
	index int
}

// Resolve validates findings against text, removes duplicates and overlaps using
// a deterministic priority policy, and returns a stable textual order. The input
// findings slice is not mutated.
func Resolve(text string, findings []Finding) ([]Finding, error) {
	if err := validateInputText(text); err != nil {
		return nil, err
	}
	if len(findings) == 0 {
		return nil, nil
	}

	candidates := make([]resolveCandidate, 0, len(findings))
	for i, finding := range findings {
		if finding.Detector == "" {
			return nil, newInvalidFindingError("resolve", finding.Entity, "detector")
		}
		validated, err := validateFinding(text, finding.Detector, finding, i)
		if err != nil {
			return nil, err
		}
		candidates = append(candidates, resolveCandidate{
			Finding: validated,
			index:   i,
		})
	}

	sort.Slice(candidates, func(i, j int) bool {
		return compareResolveSelection(candidates[i], candidates[j]) > 0
	})

	selected := make([]resolveCandidate, 0, len(candidates))
	for _, candidate := range candidates {
		if overlapsAny(candidate, selected) {
			continue
		}
		selected = append(selected, candidate)
	}

	sort.Slice(selected, func(i, j int) bool {
		return compareResolveOutput(selected[i], selected[j]) < 0
	})

	out := make([]Finding, len(selected))
	for i, candidate := range selected {
		out[i] = candidate.Finding
	}
	return out, nil
}

func compareResolveSelection(a, b resolveCandidate) int {
	if c := compareEntityPriority(a.Entity, b.Entity); c != 0 {
		return c
	}
	la, lb := a.End-a.Start, b.End-b.Start
	if la != lb {
		if la > lb {
			return 1
		}
		return -1
	}
	if a.Confidence != b.Confidence {
		if a.Confidence > b.Confidence {
			return 1
		}
		return -1
	}
	if a.Start != b.Start {
		if a.Start < b.Start {
			return 1
		}
		return -1
	}
	if a.End != b.End {
		if a.End < b.End {
			return 1
		}
		return -1
	}
	if a.Entity != b.Entity {
		if a.Entity < b.Entity {
			return 1
		}
		return -1
	}
	if a.Detector != b.Detector {
		if a.Detector < b.Detector {
			return 1
		}
		return -1
	}
	if a.index != b.index {
		if a.index < b.index {
			return 1
		}
		return -1
	}
	return 0
}

func compareResolveOutput(a, b resolveCandidate) int {
	if a.Start != b.Start {
		if a.Start < b.Start {
			return -1
		}
		return 1
	}
	if a.End != b.End {
		if a.End < b.End {
			return -1
		}
		return 1
	}
	if a.Entity != b.Entity {
		if a.Entity < b.Entity {
			return -1
		}
		return 1
	}
	if a.Detector != b.Detector {
		if a.Detector < b.Detector {
			return -1
		}
		return 1
	}
	if a.Confidence != b.Confidence {
		if a.Confidence > b.Confidence {
			return -1
		}
		return 1
	}
	return 0
}

func overlapsAny(candidate resolveCandidate, selected []resolveCandidate) bool {
	for _, existing := range selected {
		if intervalsOverlap(candidate.Start, candidate.End, existing.Start, existing.End) {
			return true
		}
	}
	return false
}

func intervalsOverlap(aStart, aEnd, bStart, bEnd int) bool {
	return aStart < bEnd && bStart < aEnd
}

func entityPriority(entity EntityType) int {
	switch entity {
	case EntityConnectionString:
		return 150
	case EntitySecretJWT, EntitySecretPrivateKey, EntitySecretAPIKey:
		return 140
	case EntityURL:
		return 110
	case EntityEmail:
		return 100
	case EntityAddress:
		return 80
	case EntityBankCard:
		return 70
	case EntitySNILS, EntityINN, EntityPassport:
		return 60
	case EntityPerson:
		return 50
	case EntityPhone:
		return 40
	case EntityIPAddress:
		return 30
	case EntityBankAccount, EntityDateOfBirth:
		return 20
	default:
		return 0
	}
}

func compareEntityPriority(a, b EntityType) int {
	pa, pb := entityPriority(a), entityPriority(b)
	if pa == pb {
		return 0
	}
	if pa > pb {
		return 1
	}
	return -1
}
