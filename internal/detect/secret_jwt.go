package detect

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"regexp"
	"strings"
)

const jwtDetectorName = "secret_jwt"

var jwtCandidatePattern = regexp.MustCompile(`[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+`)

func JWT(ctx context.Context, text string) ([]Span, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	matches := jwtCandidatePattern.FindAllStringIndex(text, -1)
	if len(matches) == 0 {
		return nil, nil
	}

	findings := make([]Span, 0, len(matches))
	for _, loc := range matches {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		start, end := loc[0], loc[1]
		if !asciiTokenBoundaryOK(text, start, end, jwtInnerBoundary) {
			continue
		}
		if !validateJWTCandidate(text[start:end]) {
			continue
		}

		findings = append(findings, Span{Start: start, End: end})
	}

	if len(findings) == 0 {
		return nil, nil
	}
	return findings, nil
}

func jwtInnerBoundary(r rune) bool {
	return r == '.' || r == '-' || r == '_'
}

func validateJWTCandidate(token string) bool {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return false
	}
	if parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return false
	}
	if !isBase64URLAlphabet(parts[0]) || !isValidUnpaddedBase64URLLength(parts[0]) {
		return false
	}
	if !isBase64URLAlphabet(parts[1]) || !isValidUnpaddedBase64URLLength(parts[1]) {
		return false
	}
	if !isBase64URLAlphabet(parts[2]) || !isValidUnpaddedBase64URLLength(parts[2]) {
		return false
	}

	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil || len(headerBytes) == 0 {
		return false
	}

	var header struct {
		Alg string `json:"alg"`
	}
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return false
	}
	if header.Alg == "" || strings.EqualFold(header.Alg, "none") {
		return false
	}

	return true
}

func isBase64URLAlphabet(segment string) bool {
	for i := 0; i < len(segment); i++ {
		switch segment[i] {
		case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M',
			'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z',
			'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm',
			'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z',
			'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '-', '_':
		default:
			return false
		}
	}
	return true
}

func isValidUnpaddedBase64URLLength(segment string) bool {
	switch len(segment) % 4 {
	case 0, 2, 3:
		return true
	default:
		return false
	}
}
