package detect

import (
	"context"
	"net/url"
	"regexp"
	"strings"
)

const dsnDetectorName = "secret_dsn"

var dsnCandidatePattern = regexp.MustCompile(
	`(?i)(?:postgres|postgresql|mysql|mongodb\+srv|mongodb|rediss|redis|amqps|amqp)://(?:[^\s<>"'{}|\\^` + "`" + `\[\]]|\[[^\]]*\])+`,
)

var dsnAllowedSchemes = map[string]struct{}{
	"postgres":    {},
	"postgresql":  {},
	"mysql":       {},
	"mongodb":     {},
	"mongodb+srv": {},
	"redis":       {},
	"rediss":      {},
	"amqp":        {},
	"amqps":       {},
}

func DSN(ctx context.Context, text string) ([]Span, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	matches := dsnCandidatePattern.FindAllStringIndex(text, -1)
	if len(matches) == 0 {
		return nil, nil
	}

	findings := make([]Span, 0, len(matches))
	for _, loc := range matches {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		start, end := loc[0], loc[1]
		end = trimDSNTrailingPunctuation(text, start, end)
		if end <= start {
			continue
		}
		if !asciiTokenBoundaryOK(text, start, end, dsnInnerBoundary) {
			continue
		}
		if !validateDSNCandidate(text[start:end]) {
			continue
		}

		findings = append(findings, Span{Start: start, End: end})
	}

	if len(findings) == 0 {
		return nil, nil
	}
	return findings, nil
}

func dsnInnerBoundary(r rune) bool {
	return r == ':' || r == '/' || r == '@' || r == '?' || r == '&' || r == '=' ||
		r == '%' || r == '+' || r == '-' || r == '_' || r == '.'
}

func trimDSNTrailingPunctuation(text string, start, end int) int {
	for end > start {
		r, size := precedingRune(text, end)
		switch r {
		case '.', ',', ';', ':', ')', '}', '"', '\'', '»', '”', '’':
			end -= size
			continue
		case ']':
			candidate := text[start : end-size]
			if strings.Count(candidate, "[") > strings.Count(candidate, "]") {
				break
			}
			end -= size
			continue
		}
		break
	}
	return end
}

func validateDSNCandidate(candidate string) bool {
	parsed, err := url.Parse(candidate)
	if err != nil {
		return false
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return false
	}
	scheme := strings.ToLower(parsed.Scheme)
	if _, ok := dsnAllowedSchemes[scheme]; !ok {
		return false
	}
	if parsed.User == nil {
		return false
	}
	username := parsed.User.Username()
	password, hasPassword := parsed.User.Password()
	if username == "" || !hasPassword || password == "" {
		return false
	}
	return true
}
