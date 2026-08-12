package llmguard

// ScanValidPlaceholdersForTest exposes placeholder scanning for tests.
func ScanValidPlaceholdersForTest(text string) []string {
	return scanValidPlaceholders(text)
}

// CountRestoreMissesForTest exposes restore-miss counting for tests.
func CountRestoreMissesForTest(response string, tokens *TokenSet) int {
	return countRestoreMisses(response, tokens)
}

// NewTokenSetForTest builds a TokenSet from explicit token/value pairs for tests.
func NewTokenSetForTest(pairs ...string) *TokenSet {
	if len(pairs)%2 != 0 {
		panic("token/value pairs must be even")
	}
	mappings := make([]tokenMapping, 0, len(pairs)/2)
	for i := 0; i < len(pairs); i += 2 {
		mappings = append(mappings, tokenMapping{token: pairs[i], value: pairs[i+1]})
	}
	return newTokenSet(mappings)
}
