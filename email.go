package llmguard

import (
	"context"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

const emailDetectorName = "email"

var emailCandidatePattern = regexp.MustCompile(
	`[a-zA-Z0-9](?:[a-zA-Z0-9._+-]*[a-zA-Z0-9])?@[a-zA-Z0-9](?:[a-zA-Z0-9-]*[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]*[a-zA-Z0-9])?)+`,
)

type emailDetector struct{}

// NewEmailDetector returns an immutable built-in EMAIL detector. Register it with
// WithDetector when constructing a Guard.
func NewEmailDetector() Detector {
	return emailDetector{}
}

func (emailDetector) Name() string {
	return emailDetectorName
}

func (emailDetector) Detect(ctx context.Context, text string) ([]Finding, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	matches := emailCandidatePattern.FindAllStringIndex(text, -1)
	if len(matches) == 0 {
		return nil, nil
	}

	findings := make([]Finding, 0, len(matches))
	for _, loc := range matches {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		start, end := loc[0], loc[1]
		if !emailBoundaryOK(text, start, end) {
			continue
		}

		mailbox := text[start:end]
		if !validateEmailMailbox(mailbox) {
			continue
		}

		findings = append(findings, Finding{
			Entity:     EntityEmail,
			Start:      start,
			End:        end,
			Confidence: 0.9,
			Detector:   emailDetectorName,
		})
	}

	if len(findings) == 0 {
		return nil, nil
	}
	return findings, nil
}

func emailBoundaryOK(text string, start, end int) bool {
	if start > 0 {
		if r, _ := precedingRune(text, start); !isEmailOuterBoundary(r) {
			return false
		}
	}
	if end < len(text) {
		if r, _ := utf8RuneAt(text, end); !isEmailOuterBoundary(r) {
			return false
		}
	}
	return true
}

func precedingRune(text string, index int) (rune, int) {
	if index <= 0 {
		return 0, 0
	}
	r, size := utf8.DecodeLastRuneInString(text[:index])
	return r, size
}

func utf8RuneAt(text string, index int) (rune, int) {
	if index < 0 || index >= len(text) {
		return 0, 0
	}
	return utf8.DecodeRuneInString(text[index:])
}

func isEmailOuterBoundary(r rune) bool {
	if unicode.IsLetter(r) || unicode.IsDigit(r) {
		return false
	}
	switch r {
	case '.', '+', '-', '_', '@':
		return false
	default:
		return true
	}
}

func validateEmailMailbox(mailbox string) bool {
	at := strings.LastIndex(mailbox, "@")
	if at <= 0 || at >= len(mailbox)-1 {
		return false
	}

	local := mailbox[:at]
	domain := mailbox[at+1:]

	return validateEmailLocalPart(local) && validateEmailDomain(domain)
}

func validateEmailLocalPart(local string) bool {
	if local == "" {
		return false
	}
	if local[0] == '.' || local[len(local)-1] == '.' {
		return false
	}
	if strings.Contains(local, "..") {
		return false
	}
	for i := 0; i < len(local); i++ {
		switch local[i] {
		case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm',
			'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z',
			'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M',
			'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z',
			'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '.', '+', '-', '_':
		default:
			return false
		}
	}
	return true
}

func validateEmailDomain(domain string) bool {
	if !strings.Contains(domain, ".") {
		return false
	}
	labels := strings.Split(domain, ".")
	for _, label := range labels {
		if label == "" {
			return false
		}
		if label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
		for j := 0; j < len(label); j++ {
			switch label[j] {
			case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm',
				'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z',
				'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M',
				'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z',
				'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '-':
			default:
				return false
			}
		}
	}
	return true
}
