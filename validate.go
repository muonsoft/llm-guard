package llmguard

import (
	"math"
	"unicode"
	"unicode/utf8"

	"github.com/muonsoft/errors"
)

func validateDetectorName(name string) error {
	if name == "" {
		return newInvalidConfigError("detector name is empty")
	}
	for _, r := range name {
		if unicode.IsControl(r) {
			return newInvalidConfigError("detector name contains control characters")
		}
	}
	return nil
}

func validateInputText(text string) error {
	if !utf8.ValidString(text) {
		return errors.Wrap(ErrInvalidText, errors.SkipCaller())
	}
	return nil
}

func validateFinding(text string, registeredName string, finding Finding, localIndex int) (Finding, error) {
	if finding.Entity == "" {
		return Finding{}, newInvalidFindingError(registeredName, finding.Entity, "entity")
	}

	if !isValidConfidence(finding.Confidence) {
		return Finding{}, newInvalidFindingError(registeredName, finding.Entity, "confidence")
	}

	textLen := len(text)
	if finding.Start < 0 || finding.End <= finding.Start || finding.End > textLen {
		return Finding{}, newInvalidFindingError(registeredName, finding.Entity, "span")
	}

	if !utf8.RuneStart(text[finding.Start]) {
		return Finding{}, newInvalidFindingError(registeredName, finding.Entity, "utf8_boundary")
	}
	if finding.End < textLen && !utf8.RuneStart(text[finding.End]) {
		return Finding{}, newInvalidFindingError(registeredName, finding.Entity, "utf8_boundary")
	}

	detectorName := finding.Detector
	switch {
	case detectorName == "":
		detectorName = registeredName
	case detectorName != registeredName:
		return Finding{}, newInvalidFindingError(registeredName, finding.Entity, "detector")
	}

	return Finding{
		Entity:     finding.Entity,
		Start:      finding.Start,
		End:        finding.End,
		Confidence: finding.Confidence,
		Detector:   detectorName,
	}, nil
}

func isValidConfidence(confidence float64) bool {
	if math.IsNaN(confidence) || math.IsInf(confidence, 0) {
		return false
	}
	return confidence >= 0 && confidence <= 1
}

func validateFindings(text, registeredName string, findings []Finding) ([]indexedFinding, error) {
	validated := make([]indexedFinding, 0, len(findings))
	for i, finding := range findings {
		normalized, err := validateFinding(text, registeredName, finding, i)
		if err != nil {
			return nil, err
		}
		validated = append(validated, indexedFinding{
			Finding:    normalized,
			localIndex: i,
		})
	}
	return validated, nil
}
