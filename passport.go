package llmguard

import (
	"context"
	"regexp"
)

const passportDetectorName = "passport"

var (
	passportFourSixPattern   = regexp.MustCompile(`\d{4} \d{6}`)
	passportTwoTwoSixPattern = regexp.MustCompile(`\d{2} \d{2} \d{6}`)
)

var passportContextMarkers = []string{
	"паспортные данные",
	"паспорт рф",
	"паспорт",
	"серия",
}

type passportDetector struct{}

// NewPassportDetector returns an immutable built-in PASSPORT detector.
func NewPassportDetector() Detector {
	return passportDetector{}
}

func (passportDetector) Name() string {
	return passportDetectorName
}

func (passportDetector) Detect(ctx context.Context, text string) ([]Finding, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var findings []Finding
	for _, pattern := range []*regexp.Regexp{passportFourSixPattern, passportTwoTwoSixPattern} {
		matches := pattern.FindAllStringIndex(text, -1)
		for _, loc := range matches {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			start, end := loc[0], loc[1]
			if overlapsExisting(findings, start, end) {
				continue
			}
			if !numericIdentifierBoundaryOK(text, start, end) {
				continue
			}
			if !validatePassportSegment(text[start:end]) {
				continue
			}
			if !hasBoundedRUContext(text, start, passportContextMarkers) {
				continue
			}
			findings = append(findings, Finding{
				Entity:     EntityPassport,
				Start:      start,
				End:        end,
				Confidence: 0.86,
				Detector:   passportDetectorName,
			})
		}
	}

	if len(findings) == 0 {
		return nil, nil
	}
	return findings, nil
}

func validatePassportSegment(segment string) bool {
	digits := collectASCIIDigits(segment)
	if len(digits) != 10 {
		return false
	}
	if segment == digits {
		return true
	}
	if passportFourSixPattern.MatchString(segment) {
		return true
	}
	return passportTwoTwoSixPattern.MatchString(segment)
}
