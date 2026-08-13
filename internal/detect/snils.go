package detect

import (
	"context"
	"regexp"
	"strconv"
)

const snilsDetectorName = "snils"

var (
	snilsCompactPattern   = regexp.MustCompile(`\d{11}`)
	snilsFormattedPattern = regexp.MustCompile(`\d{3}-\d{3}-\d{3} \d{2}`)
)

func SNILS(ctx context.Context, text string) ([]Span, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var findings []Span

	formatted := snilsFormattedPattern.FindAllStringIndex(text, -1)
	for _, loc := range formatted {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		start, end := loc[0], loc[1]
		if !numericIdentifierBoundaryOK(text, start, end) {
			continue
		}
		segment := text[start:end]
		if !validateSNILSSegment(segment) {
			continue
		}
		findings = append(findings, Span{Start: start, End: end})
	}

	compact := snilsCompactPattern.FindAllStringIndex(text, -1)
	for _, loc := range compact {
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
		segment := text[start:end]
		if !validateSNILSSegment(segment) {
			continue
		}
		findings = append(findings, Span{Start: start, End: end})
	}

	if len(findings) == 0 {
		return nil, nil
	}
	return findings, nil
}

func validateSNILSSegment(segment string) bool {
	digits := collectASCIIDigits(segment)
	if len(digits) != 11 {
		return false
	}
	if segment != digits {
		if !snilsFormattedPattern.MatchString(segment) {
			return false
		}
	}
	return validateSNILSDigits(digits)
}

func validateSNILSDigits(digits string) bool {
	numberPart, err := strconv.Atoi(digits[:9])
	if err != nil {
		return false
	}
	if numberPart < 1001998 {
		return false
	}

	sum := 0
	for i := 0; i < 9; i++ {
		sum += int(digits[i]-'0') * (9 - i)
	}

	var check int
	switch {
	case sum < 100:
		check = sum
	case sum == 100 || sum == 101:
		check = 0
	default:
		check = sum % 101
		if check == 100 {
			check = 0
		}
	}

	expected, err := strconv.Atoi(digits[9:11])
	if err != nil {
		return false
	}
	return check == expected
}
