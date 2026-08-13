package detect

import (
	"context"
	"regexp"
)

const innDetectorName = "inn"

var innCandidatePattern = regexp.MustCompile(`\d{12}|\d{10}`)

func INN(ctx context.Context, text string) ([]Span, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	matches := innCandidatePattern.FindAllStringIndex(text, -1)
	if len(matches) == 0 {
		return nil, nil
	}

	findings := make([]Span, 0, len(matches))
	for _, loc := range matches {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		start, end := loc[0], loc[1]
		if !numericIdentifierBoundaryOK(text, start, end) {
			continue
		}

		digits := text[start:end]
		if !validateINN(digits) {
			continue
		}

		findings = append(findings, Span{Start: start, End: end})
	}

	if len(findings) == 0 {
		return nil, nil
	}
	return findings, nil
}

func validateINN(digits string) bool {
	if allSameDigits(digits) {
		return false
	}
	switch len(digits) {
	case 10:
		return validateINN10(digits)
	case 12:
		return validateINN12(digits)
	default:
		return false
	}
}

func validateINN10(digits string) bool {
	coef := []int{2, 4, 10, 3, 5, 9, 4, 6, 8}
	sum := 0
	for i := 0; i < 9; i++ {
		sum += int(digits[i]-'0') * coef[i]
	}
	check := sum % 11 % 10
	return int(digits[9]-'0') == check
}

func validateINN12(digits string) bool {
	coef11 := []int{7, 2, 4, 10, 3, 5, 9, 4, 6, 8}
	sum11 := 0
	for i := 0; i < 10; i++ {
		sum11 += int(digits[i]-'0') * coef11[i]
	}
	check11 := sum11 % 11 % 10
	if int(digits[10]-'0') != check11 {
		return false
	}

	coef12 := []int{3, 7, 2, 4, 10, 3, 5, 9, 4, 6, 8}
	sum12 := 0
	for i := 0; i < 11; i++ {
		sum12 += int(digits[i]-'0') * coef12[i]
	}
	check12 := sum12 % 11 % 10
	return int(digits[11]-'0') == check12
}
