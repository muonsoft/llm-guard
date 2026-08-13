package detect

import (
	"context"
	"regexp"
	"strings"
	"unicode"
)

const phoneDetectorName = "phone"

var phoneCandidatePattern = regexp.MustCompile(
	`\+[0-9][0-9\s\-().]{5,22}[0-9]|(?:\+7|8)[0-9\s\-().]{8,22}[0-9]`,
)

func Phone(ctx context.Context, text string) ([]Span, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	matches := phoneCandidatePattern.FindAllStringIndex(text, -1)
	if len(matches) == 0 {
		return nil, nil
	}

	findings := make([]Span, 0, len(matches))
	for _, loc := range matches {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		start, end := loc[0], loc[1]
		if !phoneBoundaryOK(text, start, end) {
			continue
		}

		segment := text[start:end]
		if !validatePhoneSegment(segment) {
			continue
		}

		findings = append(findings, Span{Start: start, End: end})
	}

	if len(findings) == 0 {
		return nil, nil
	}
	return findings, nil
}

func phoneBoundaryOK(text string, start, end int) bool {
	if start > 0 {
		if r, _ := precedingRune(text, start); isPhoneAdjacentRune(r) {
			return false
		}
	}
	if end < len(text) {
		if r, _ := utf8RuneAt(text, end); isPhoneAdjacentRune(r) {
			return false
		}
	}
	return true
}

func isPhoneAdjacentRune(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '+' || r == '(' || r == ')' || r == '-'
}

func validatePhoneSegment(segment string) bool {
	if segment == "" {
		return false
	}
	if strings.Contains(segment, ".") {
		return false
	}
	for _, r := range segment {
		if unicode.IsLetter(r) {
			return false
		}
	}

	digits := collectASCIIDigits(segment)
	if len(digits) < 8 || len(digits) > 15 {
		return false
	}

	if segment[0] == '+' {
		if len(digits) >= 1 && digits[0] == '7' {
			return validateRUPhone(segment, digits)
		}
		return validateInternationalPhone(segment, digits)
	}
	if segment[0] == '8' {
		return validateRUPhone(segment, digits)
	}
	return false
}

func validateRUPhone(segment string, digits string) bool {
	if len(digits) != 11 {
		return false
	}
	if digits[0] != '7' && digits[0] != '8' {
		return false
	}
	return ruPhoneFormattingOK(segment, digits)
}

func validateInternationalPhone(segment string, digits string) bool {
	if len(digits) < 8 || len(digits) > 15 {
		return false
	}
	return internationalPhoneFormattingOK(segment)
}

func internationalPhoneFormattingOK(segment string) bool {
	open := strings.IndexByte(segment, '(')
	close := strings.IndexByte(segment, ')')
	if open == -1 && close == -1 {
		return true
	}
	if open == -1 || close == -1 || close < open {
		return false
	}
	if strings.Count(segment, "(") != 1 || strings.Count(segment, ")") != 1 {
		return false
	}
	group := segment[open+1 : close]
	if group == "" {
		return false
	}
	for _, r := range group {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func ruPhoneFormattingOK(segment string, digits string) bool {
	open := strings.IndexByte(segment, '(')
	close := strings.IndexByte(segment, ')')
	if open == -1 && close == -1 {
		return true
	}
	if open == -1 || close == -1 || close < open {
		return false
	}
	if strings.Count(segment, "(") != 1 || strings.Count(segment, ")") != 1 {
		return false
	}
	area := segment[open+1 : close]
	if len(area) != 3 {
		return false
	}
	for _, r := range area {
		if r < '0' || r > '9' {
			return false
		}
	}
	national := digits[1:]
	return strings.HasPrefix(national, area)
}
