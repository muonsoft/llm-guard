package nlp

// ValidHouseIdentifierForTest exposes house-identifier validation for black-box tests.
func ValidHouseIdentifierForTest(segment string) bool {
	return validHouseIdentifier(segment)
}
