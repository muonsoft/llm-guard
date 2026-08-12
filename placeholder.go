package llmguard

import (
	"unicode"
)

// scanValidPlaceholders returns distinct syntactically valid llm-guard placeholders
// found in text using the bounded grammar {{LLMG_<32 lowercase hex>_<4+ digits>}}.
func scanValidPlaceholders(text string) []string {
	if text == "" {
		return nil
	}

	seen := make(map[string]struct{})
	var tokens []string
	i := 0
	for i < len(text) {
		token, next := parsePlaceholderAt(text, i)
		if token == "" {
			i++
			continue
		}
		if _, ok := seen[token]; !ok {
			seen[token] = struct{}{}
			tokens = append(tokens, token)
		}
		i = next
	}
	return tokens
}

func hasTokenPrefixAt(text string, start int) bool {
	return start+len(tokenPrefix) <= len(text) && text[start:start+len(tokenPrefix)] == tokenPrefix
}

func parsePlaceholderAt(text string, start int) (token string, next int) {
	if !hasTokenPrefixAt(text, start) {
		return "", start + 1
	}

	pos := start + len(tokenPrefix)
	if pos+32+1+4+len(tokenSuffix) > len(text) {
		return "", start + 1
	}

	for j := 0; j < 32; j++ {
		if !isLowerHex(text[pos+j]) {
			return "", start + 1
		}
	}
	pos += 32
	if text[pos] != '_' {
		return "", start + 1
	}
	pos++

	digitStart := pos
	for pos < len(text) && unicode.IsDigit(rune(text[pos])) {
		pos++
	}
	if pos-digitStart < 4 {
		return "", start + 1
	}
	if pos+len(tokenSuffix) > len(text) || text[pos:pos+len(tokenSuffix)] != tokenSuffix {
		return "", start + 1
	}
	pos += len(tokenSuffix)
	return text[start:pos], pos
}

func isLowerHex(b byte) bool {
	return (b >= '0' && b <= '9') || (b >= 'a' && b <= 'f')
}

func countRestoreMisses(response string, tokens *TokenSet) int {
	if response == "" {
		return 0
	}

	misses := 0
	i := 0
	for i < len(response) {
		token, next := parsePlaceholderAt(response, i)
		if token == "" {
			i++
			continue
		}
		if !tokens.hasToken(token) {
			misses++
		}
		i = next
	}
	return misses
}

func (t *TokenSet) hasToken(token string) bool {
	if t == nil {
		return false
	}
	for _, mapping := range t.mappings {
		if mapping.token == token {
			return true
		}
	}
	return false
}
