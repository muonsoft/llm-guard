package detect

import (
	"context"
	"regexp"
)

const bankCardDetectorName = "bank_card"

var bankCardCandidatePattern = regexp.MustCompile(
	`(?:\d[ -]?){12,18}\d`,
)

func BankCard(ctx context.Context, text string) ([]Span, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	matches := bankCardCandidatePattern.FindAllStringIndex(text, -1)
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

		segment := text[start:end]
		if !validateBankCardSegment(segment) {
			continue
		}

		findings = append(findings, Span{Start: start, End: end})
	}

	if len(findings) == 0 {
		return nil, nil
	}
	return findings, nil
}

func validateBankCardSegment(segment string) bool {
	if !consistentDigitSeparators(segment) {
		return false
	}
	digits := collectASCIIDigits(segment)
	if len(digits) < 13 || len(digits) > 19 {
		return false
	}
	if allSameDigits(digits) {
		return false
	}
	return luhnValid(digits)
}

func luhnValid(digits string) bool {
	sum := 0
	parity := len(digits) % 2
	for i, ch := range digits {
		n := int(ch - '0')
		if i%2 == parity {
			n *= 2
			if n > 9 {
				n -= 9
			}
		}
		sum += n
	}
	return sum%10 == 0
}
