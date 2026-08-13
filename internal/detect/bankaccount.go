package detect

import (
	"context"
	"regexp"
	"unicode"
	"unicode/utf8"
)

const bankAccountDetectorName = "bank_account"

var (
	bankAccountCompactPattern = regexp.MustCompile(`\d{20}`)
	bankAccountGroupedPattern = regexp.MustCompile(`\d{4} \d{4} \d{4} \d{4} \d{4}`)
	bikLabelPattern           = regexp.MustCompile(`(?i)бик`)
)

var bankAccountContextMarkers = []string{
	"банковский счёт",
	"банковский счет",
	"расчётный счёт",
	"расчетный счет",
	"р/с",
}

func BankAccount(ctx context.Context, text string) ([]Span, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var findings []Span
	for _, pattern := range []*regexp.Regexp{bankAccountGroupedPattern, bankAccountCompactPattern} {
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
			segment := text[start:end]
			digits, ok := normalizeBankAccountSegment(segment)
			if !ok {
				continue
			}
			if !hasBoundedRUContext(text, start, bankAccountContextMarkers) {
				continue
			}
			if bik, found := findLabeledBIKOnLine(text, start, end); found {
				if !validateBankAccountBIKChecksum(bik, digits) {
					continue
				}
			}
			findings = append(findings, Span{Start: start, End: end})
		}
	}

	if len(findings) == 0 {
		return nil, nil
	}
	return findings, nil
}

func normalizeBankAccountSegment(segment string) (string, bool) {
	if bankAccountCompactPattern.MatchString(segment) {
		if segment != collectASCIIDigits(segment) {
			return "", false
		}
		digits := segment
		if allSameDigits(digits) {
			return "", false
		}
		return digits, true
	}
	if !bankAccountGroupedPattern.MatchString(segment) {
		return "", false
	}
	digits := collectASCIIDigits(segment)
	if len(digits) != 20 || allSameDigits(digits) {
		return "", false
	}
	return digits, true
}

func findLabeledBIKOnLine(text string, accountStart, accountEnd int) (string, bool) {
	lineStart := lineStartIndex(text, accountStart)
	lineEnd := lineEndIndex(text, accountEnd)
	line := text[lineStart:lineEnd]

	labelLocs := bikLabelPattern.FindAllStringIndex(line, -1)
	for _, loc := range labelLocs {
		labelOffset := loc[0]
		if !isValidMarkerOccurrence(line, labelOffset, "бик") {
			continue
		}
		labelEnd := byteIndexAfterCaseInsensitiveMatch(line, labelOffset, "бик")
		digits, relStart, relEnd, ok := parseLabeledBIKDigits(line, labelEnd)
		if !ok {
			continue
		}
		bikStart := lineStart + relStart
		bikEnd := lineStart + relEnd
		if !numericIdentifierBoundaryOK(text, bikStart, bikEnd) {
			continue
		}
		if intervalsOverlap(bikStart, bikEnd, accountStart, accountEnd) {
			continue
		}
		return digits, true
	}
	return "", false
}

func parseLabeledBIKDigits(line string, afterLabel int) (string, int, int, bool) {
	i := afterLabel
	gapBytes := 0
	for i < len(line) {
		r, size := utf8.DecodeRuneInString(line[i:])
		if size == 0 {
			break
		}
		if unicode.IsSpace(r) {
			gapBytes += size
			if gapBytes > 8 {
				return "", 0, 0, false
			}
			i += size
			continue
		}
		switch r {
		case ':', '№', '#', '-', ',':
			gapBytes += size
			if gapBytes > 8 {
				return "", 0, 0, false
			}
			i += size
			continue
		}
		break
	}
	if i+9 > len(line) {
		return "", 0, 0, false
	}
	digits := line[i : i+9]
	for j := 0; j < 9; j++ {
		if digits[j] < '0' || digits[j] > '9' {
			return "", 0, 0, false
		}
	}
	return digits, i, i + 9, true
}

func validateBankAccountBIKChecksum(bik, account string) bool {
	if len(bik) != 9 || len(account) != 20 {
		return false
	}
	combined := bik[6:] + account
	if len(combined) != 23 {
		return false
	}
	weights := []int{7, 1, 3}
	sum := 0
	for i := 0; i < len(combined); i++ {
		d := int(combined[i] - '0')
		if combined[i] < '0' || combined[i] > '9' {
			return false
		}
		sum += d * weights[i%3]
	}
	return sum%10 == 0
}
