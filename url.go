package llmguard

import (
	"context"
	"net"
	"net/url"
	"regexp"
	"strings"
)

const urlDetectorName = "url"

var urlCandidatePattern = regexp.MustCompile(`https?://(?:[^\s<>"{}|\\^` + "`" + `\[\]]|\[[^\]]*\])+`)

type urlDetector struct{}

// NewURLDetector returns an immutable built-in URL detector.
func NewURLDetector() Detector {
	return urlDetector{}
}

func (urlDetector) Name() string {
	return urlDetectorName
}

func (urlDetector) Detect(ctx context.Context, text string) ([]Finding, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	matches := urlCandidatePattern.FindAllStringIndex(text, -1)
	if len(matches) == 0 {
		return nil, nil
	}

	findings := make([]Finding, 0, len(matches))
	for _, loc := range matches {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		start := loc[0]
		rawEnd := loc[1]
		trimmedEnd, ok := trimURLTrailingPunctuation(text, start, rawEnd)
		if !ok {
			continue
		}
		if !urlBoundaryOK(text, start, trimmedEnd) {
			continue
		}

		segment := text[start:trimmedEnd]
		if !validateURLSegment(segment) {
			continue
		}

		findings = append(findings, Finding{
			Entity:     EntityURL,
			Start:      start,
			End:        trimmedEnd,
			Confidence: 0.88,
			Detector:   urlDetectorName,
		})
	}

	if len(findings) == 0 {
		return nil, nil
	}
	return findings, nil
}

func urlBoundaryOK(text string, start, end int) bool {
	return asciiTokenBoundaryOK(text, start, end, nil)
}

func validateURLSegment(segment string) bool {
	if strings.ContainsAny(segment, " \t\r\n") {
		return false
	}

	parsed, err := url.Parse(segment)
	if err != nil || parsed == nil {
		return false
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return false
	}
	if parsed.Host == "" {
		return false
	}
	host := parsed.Hostname()
	if host == "" {
		return false
	}
	if ip := net.ParseIP(host); ip != nil {
		return true
	}
	if strings.Contains(host, ".") && dnsLikeHostOK(host) {
		return true
	}
	return false
}

func dnsLikeHostOK(host string) bool {
	if host == "" || strings.HasPrefix(host, ".") || strings.HasSuffix(host, ".") {
		return false
	}
	if strings.Contains(host, "..") {
		return false
	}
	labels := strings.Split(host, ".")
	if len(labels) < 2 {
		return false
	}
	for _, label := range labels {
		if label == "" {
			return false
		}
		for _, r := range label {
			switch {
			case r >= 'a' && r <= 'z':
			case r >= 'A' && r <= 'Z':
			case r >= '0' && r <= '9':
			case r == '-':
			default:
				return false
			}
		}
		if label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
	}
	return true
}

var urlTrailingPunctuation = map[rune]struct{}{
	'.': {}, ',': {}, ';': {}, ':': {}, '!': {}, '?': {},
	')': {}, ']': {}, '}': {}, '>': {},
}

func trimURLTrailingPunctuation(text string, start, end int) (int, bool) {
	for end > start {
		r, size := precedingRune(text, end)
		if _, ok := urlTrailingPunctuation[r]; !ok {
			break
		}
		if !balancedAfterTrim(text[start:end], r) {
			break
		}
		end -= size
	}
	if end <= start {
		return 0, false
	}
	return end, true
}

func balancedAfterTrim(segment string, trimmed rune) bool {
	switch trimmed {
	case ')':
		return countRune(segment, ')') > countRune(segment, '(')
	case ']':
		return countRune(segment, ']') > countRune(segment, '[')
	case '}':
		return countRune(segment, '}') > countRune(segment, '{')
	case '>':
		return countRune(segment, '>') > countRune(segment, '<')
	default:
		return true
	}
}

func countRune(s string, target rune) int {
	count := 0
	for _, r := range s {
		if r == target {
			count++
		}
	}
	return count
}
