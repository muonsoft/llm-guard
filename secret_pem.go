package llmguard

import (
	"context"
	"encoding/base64"
	"strings"
)

const pemPrivateKeyDetectorName = "secret_private_key"

var pemPrivateKeyLabels = []string{
	"PRIVATE KEY",
	"RSA PRIVATE KEY",
	"EC PRIVATE KEY",
	"OPENSSH PRIVATE KEY",
	"PGP PRIVATE KEY BLOCK",
}

type pemPrivateKeyDetector struct{}

// NewPEMPrivateKeyDetector returns an immutable built-in SECRET_PRIVATE_KEY detector.
func NewPEMPrivateKeyDetector() Detector {
	return pemPrivateKeyDetector{}
}

func (pemPrivateKeyDetector) Name() string {
	return pemPrivateKeyDetectorName
}

func (pemPrivateKeyDetector) Detect(ctx context.Context, text string) ([]Finding, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var findings []Finding
	searchFrom := 0
	for searchFrom < len(text) {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		beginIdx := strings.Index(text[searchFrom:], "-----BEGIN ")
		if beginIdx == -1 {
			break
		}
		start := searchFrom + beginIdx
		closeBegin := strings.Index(text[start+len("-----BEGIN "):], "-----")
		if closeBegin == -1 {
			searchFrom = start + len("-----BEGIN ")
			continue
		}
		headerEnd := start + len("-----BEGIN ") + closeBegin + len("-----")

		label, ok := matchPEMPrivateLabel(text[start:headerEnd])
		if !ok {
			searchFrom = start + len("-----BEGIN ")
			continue
		}
		if !hasPEMLineBreakAfter(text, headerEnd) {
			searchFrom = start + len("-----BEGIN ")
			continue
		}

		bodyStart := skipPEMLineBreak(text, headerEnd)
		footerMarker := "-----END " + label + "-----"
		footerIdx := strings.Index(text[bodyStart:], footerMarker)
		if footerIdx == -1 {
			searchFrom = start + len("-----BEGIN ")
			continue
		}
		footerStart := bodyStart + footerIdx
		if !hasPEMLineBreakBefore(text, footerStart) {
			searchFrom = start + len("-----BEGIN ")
			continue
		}
		end := footerStart + len(footerMarker)

		body := strings.TrimSpace(text[bodyStart:footerStart])
		if !validatePEMBody(body) {
			searchFrom = start + len("-----BEGIN ")
			continue
		}

		findings = append(findings, Finding{
			Entity:     EntitySecretPrivateKey,
			Start:      start,
			End:        end,
			Confidence: 0.95,
			Detector:   pemPrivateKeyDetectorName,
		})
		searchFrom = end
	}

	if len(findings) == 0 {
		return nil, nil
	}
	return findings, nil
}

func hasPEMLineBreakAfter(text string, index int) bool {
	if index >= len(text) {
		return false
	}
	if text[index] == '\n' {
		return true
	}
	return text[index] == '\r' && index+1 < len(text) && text[index+1] == '\n'
}

func hasPEMLineBreakBefore(text string, index int) bool {
	if index <= 0 {
		return false
	}
	if text[index-1] == '\n' {
		return true
	}
	return index >= 2 && text[index-2] == '\r' && text[index-1] == '\n'
}

func skipPEMLineBreak(text string, index int) int {
	if index < len(text) && text[index] == '\n' {
		return index + 1
	}
	if index+1 < len(text) && text[index] == '\r' && text[index+1] == '\n' {
		return index + 2
	}
	return index
}

func matchPEMPrivateLabel(header string) (string, bool) {
	for _, label := range pemPrivateKeyLabels {
		expected := "-----BEGIN " + label + "-----"
		if header == expected {
			return label, true
		}
	}
	return "", false
}

func validatePEMBody(body string) bool {
	if body == "" {
		return false
	}
	normalized := strings.ReplaceAll(body, "\r\n", "\n")
	lines := strings.Split(normalized, "\n")
	var encoded strings.Builder
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		for i := 0; i < len(line); i++ {
			switch line[i] {
			case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M',
				'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z',
				'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm',
				'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z',
				'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '+', '/', '=':
			default:
				return false
			}
		}
		encoded.WriteString(line)
	}
	if encoded.Len() == 0 {
		return false
	}
	_, err := base64.StdEncoding.DecodeString(encoded.String())
	return err == nil
}
