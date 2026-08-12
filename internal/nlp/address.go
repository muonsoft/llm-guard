package nlp

import (
	"context"
	"strings"
	"unicode"
)

type addressLabel struct {
	short string
	full  string
}

var (
	streetPrefixLabels = []addressLabel{
		{short: "ул", full: "улица"},
		{short: "пр-т", full: "проспект"},
		{short: "просп", full: "проспект"},
		{short: "пер", full: "переулок"},
		{short: "", full: "шоссе"},
	}
	streetSuffixLabels = streetPrefixLabels
	settlementLabels   = []addressLabel{{short: "г", full: "город"}}
	houseLabels        = []addressLabel{{short: "д", full: "дом"}}
	corpusLabels       = []addressLabel{{short: "корп", full: "корпус"}}
	buildingLabels     = []addressLabel{{short: "стр", full: "строение"}}
	apartmentLabels    = []addressLabel{{short: "кв", full: "квартира"}}
)

const maxStreetNameTokens = 4
const maxSettlementNameTokens = 2

// DetectAddressSpans finds conservative compositional ADDRESS spans in text.
func DetectAddressSpans(ctx context.Context, text string) ([]Span, error) {
	if err := checkScanCtx(ctx, "pre_tokenize"); err != nil {
		return nil, err
	}

	tokens := Tokenize(text)
	if err := checkScanCtx(ctx, "post_tokenize"); err != nil {
		return nil, err
	}
	if len(tokens) == 0 {
		return nil, nil
	}

	var spans []Span
	for i := 0; i < len(tokens); {
		if err := checkScanCtx(ctx, "scan"); err != nil {
			return nil, err
		}
		if span, consumed, ok := matchAddressAt(text, tokens, i); ok {
			spans = append(spans, span)
			i += consumed
			continue
		}
		i++
	}
	return spans, nil
}

func matchAddressAt(text string, tokens []Token, start int) (Span, int, bool) {
	streetStarts := []int{start}
	if streetStart, ok := parseSettlementPrefixForward(tokens, start); ok {
		streetStarts = appendUniqueInt(streetStarts, streetStart)
	}

	var best Span
	bestConsumed := 0
	found := false

	for _, streetStart := range streetStarts {
		spanStartIdx := streetStart
		if sm, ok := matchSettlementBeforeStreet(tokens, streetStart); ok {
			spanStartIdx = sm.settlementStart
		}
		if spanStartIdx > start {
			continue
		}

		span, consumed, ok := composeAddressFromStreet(text, tokens, spanStartIdx, streetStart, start)
		if !ok {
			continue
		}
		if !found || span.End > best.End || (span.End == best.End && span.Start < best.Start) {
			best = span
			bestConsumed = consumed
			found = true
		}
	}

	if !found {
		return Span{}, 0, false
	}
	return best, bestConsumed, true
}

func composeAddressFromStreet(text string, tokens []Token, spanStartIdx, streetStart, scanStart int) (Span, int, bool) {
	_, afterStreet, ok := matchStreet(tokens, streetStart)
	if !ok {
		return Span{}, 0, false
	}
	houseStart, ok := skipOptionalCommaSeparator(tokens, afterStreet)
	if !ok {
		return Span{}, 0, false
	}
	houseEnd, afterHouse, ok := matchHouse(tokens, houseStart)
	if !ok {
		return Span{}, 0, false
	}

	endIdx := houseEnd
	cursor := afterHouse
	if extEnd, afterExt, ok := matchExtendedParts(tokens, cursor); ok {
		endIdx = extEnd
		cursor = afterExt
	} else if cursor < len(tokens) {
		if _, sepOK := skipOptionalCommaSeparator(tokens, cursor); !sepOK {
			return Span{}, 0, false
		}
		if looksLikeExtendedAddressPart(tokens, cursor) {
			return Span{}, 0, false
		}
	}

	spanStart := tokens[spanStartIdx].Start
	spanEnd := tokens[endIdx].End
	if spanStart >= spanEnd || strings.Contains(text[spanStart:spanEnd], "\n") {
		return Span{}, 0, false
	}
	if !outerBoundariesOK(text, spanStart, spanEnd) {
		return Span{}, 0, false
	}

	consumed := cursor - scanStart
	if consumed <= 0 {
		return Span{}, 0, false
	}
	return Span{Start: spanStart, End: spanEnd}, consumed, true
}

func parseSettlementPrefixForward(tokens []Token, start int) (streetStart int, ok bool) {
	idx := start
	if n, ok := matchAddressLabel(tokens, idx, settlementLabels[0]); ok {
		idx = n
	}
	_, afterName, ok := matchSettlementNameForward(tokens, idx)
	if !ok {
		return 0, false
	}
	if afterName >= len(tokens) || tokens[afterName].Kind != KindPunctuation || tokens[afterName].Text != "," {
		return 0, false
	}
	return afterName + 1, true
}

func matchSettlementNameForward(tokens []Token, start int) (endIdx, next int, ok bool) {
	if start >= len(tokens) || !isSettlementNameToken(tokens[start]) {
		return 0, 0, false
	}
	end := start
	count := 1
	for i := start + 1; i < len(tokens) && count < maxSettlementNameTokens; i++ {
		if !addressGapOK(tokens, i-1, i) {
			break
		}
		if !isSettlementNameToken(tokens[i]) {
			break
		}
		end = i
		count++
	}
	return end, end + 1, true
}

func appendUniqueInt(values []int, value int) []int {
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}

func matchStreet(tokens []Token, start int) (endIdx, next int, ok bool) {
	if end, n, ok := matchStreetPrefix(tokens, start); ok {
		return end, n, true
	}
	if end, n, ok := matchStreetSuffix(tokens, start); ok {
		return end, n, true
	}
	return 0, 0, false
}

func matchStreetPrefix(tokens []Token, start int) (endIdx, next int, ok bool) {
	idx := start
	matchedLabel := false
	for _, label := range streetPrefixLabels {
		n, ok := matchAddressLabel(tokens, idx, label)
		if !ok {
			continue
		}
		idx = n
		matchedLabel = true
		break
	}
	if !matchedLabel {
		return 0, 0, false
	}

	nameEnd, afterName, ok := matchStreetNameTokens(tokens, idx)
	if !ok {
		return 0, 0, false
	}
	return nameEnd, afterName, true
}

func matchStreetSuffix(tokens []Token, start int) (endIdx, next int, ok bool) {
	_, afterName, ok := matchStreetNameTokens(tokens, start)
	if !ok {
		return 0, 0, false
	}
	for _, label := range streetSuffixLabels {
		if n, ok := matchAddressLabel(tokens, afterName, label); ok {
			return n - 1, n, true
		}
	}
	return 0, 0, false
}

func matchStreetNameTokens(tokens []Token, start int) (endIdx, next int, ok bool) {
	if start >= len(tokens) || !isStreetNameToken(tokens[start]) {
		return 0, 0, false
	}
	end := start
	count := 1
	for i := start + 1; i < len(tokens) && count < maxStreetNameTokens; i++ {
		if !addressGapOK(tokens, i-1, i) {
			break
		}
		if !isStreetNameToken(tokens[i]) {
			break
		}
		end = i
		count++
	}
	return end, end + 1, true
}

func isStreetNameToken(tok Token) bool {
	if tok.Kind != KindWord {
		return false
	}
	if isAddressLabelWord(tok.Folded) {
		return false
	}
	return tok.Capitalized || isOrdinalStreetToken(tok)
}

func isOrdinalStreetToken(tok Token) bool {
	// Genitive/descriptive street-name tokens like "Академика".
	if !tok.Capitalized {
		return false
	}
	runes := []rune(tok.Text)
	return len(runes) >= 4 && strings.HasSuffix(tok.Folded, "а")
}

func matchHouse(tokens []Token, start int) (endIdx, next int, ok bool) {
	if start >= len(tokens) {
		return 0, 0, false
	}
	if n, ok := matchAddressLabel(tokens, start, houseLabels[0]); ok {
		idEnd, afterID, ok := matchHouseIdentifier(tokens, n)
		if !ok {
			return 0, 0, false
		}
		return idEnd, afterID, true
	}
	return matchHouseIdentifier(tokens, start)
}

func matchHouseIdentifier(tokens []Token, start int) (endIdx, next int, ok bool) {
	if start >= len(tokens) {
		return 0, 0, false
	}
	if tokens[start].Kind == KindInteger && start+1 < len(tokens) &&
		tokens[start].End == tokens[start+1].Start && isSingleLetterHouseSuffix(tokens[start+1]) {
		combined := tokens[start].Text + tokens[start+1].Text
		if validHouseIdentifier(combined) {
			return start + 1, start + 2, true
		}
		return 0, 0, false
	}
	if isHouseIdentifierToken(tokens[start]) {
		return start, start + 1, true
	}
	return 0, 0, false
}

func isSingleLetterHouseSuffix(tok Token) bool {
	if tok.Kind != KindWord {
		return false
	}
	runes := []rune(tok.Text)
	if len(runes) != 1 {
		return false
	}
	r := runes[0]
	return unicode.IsLetter(r) && (r <= unicode.MaxASCII || isCyrillicLetter(r))
}

func matchExtendedParts(tokens []Token, start int) (endIdx, next int, ok bool) {
	idx, ok := skipOptionalCommaSeparator(tokens, start)
	if !ok {
		return 0, 0, false
	}
	extended := false

	if _, idEnd, ok := matchLabeledIdentifier(tokens, idx, corpusLabels[0]); ok {
		idx, ok = skipOptionalCommaSeparator(tokens, idEnd+1)
		if !ok {
			return 0, 0, false
		}
		endIdx = idEnd
		extended = true
	}
	if _, idEnd, ok := matchLabeledIdentifier(tokens, idx, buildingLabels[0]); ok {
		idx, ok = skipOptionalCommaSeparator(tokens, idEnd+1)
		if !ok {
			return 0, 0, false
		}
		endIdx = idEnd
		extended = true
	}
	if _, idEnd, ok := matchLabeledIdentifier(tokens, idx, apartmentLabels[0]); ok {
		endIdx = idEnd
		extended = true
	}
	if !extended {
		return 0, 0, false
	}
	return endIdx, endIdx + 1, true
}

func matchLabeledIdentifier(tokens []Token, start int, label addressLabel) (labelEnd, idEnd int, ok bool) {
	if start >= len(tokens) {
		return 0, 0, false
	}
	if start > 0 && !addressPartGapOK(tokens, start-1, start) {
		return 0, 0, false
	}
	n, ok := matchAddressLabel(tokens, start, label)
	if !ok {
		return 0, 0, false
	}
	if n >= len(tokens) {
		return 0, 0, false
	}
	houseEnd, _, matched := matchHouseIdentifier(tokens, n)
	if !matched {
		return 0, 0, false
	}
	return n - 1, houseEnd, true
}

type settlementMatch struct {
	settlementStart int
	streetStart     int
}

func matchSettlementBeforeStreet(tokens []Token, streetStart int) (settlementMatch, bool) {
	if streetStart == 0 {
		return settlementMatch{}, false
	}
	commaIdx := streetStart - 1
	if tokens[commaIdx].Kind != KindPunctuation || tokens[commaIdx].Text != "," {
		return settlementMatch{}, false
	}
	_, nameStart, ok := matchSettlementNameBackward(tokens, commaIdx-1)
	if !ok {
		return settlementMatch{}, false
	}
	settlementStart := nameStart
	if labelStart, ok := matchSettlementLabelBefore(tokens, nameStart); ok {
		settlementStart = labelStart
	}
	return settlementMatch{settlementStart: settlementStart, streetStart: streetStart}, true
}

func matchSettlementLabelBefore(tokens []Token, nameStart int) (int, bool) {
	if nameStart == 0 {
		return 0, false
	}
	idx := nameStart - 1
	if tokens[idx].Kind == KindPunctuation && tokens[idx].Text == "." {
		if idx == 0 {
			return 0, false
		}
		if tokens[idx-1].Kind == KindWord && tokens[idx-1].Folded == "г" {
			return idx - 1, true
		}
		return 0, false
	}
	if n, ok := matchAddressLabel(tokens, nameStart-1, settlementLabels[0]); ok && n == nameStart {
		return nameStart - 1, true
	}
	return 0, false
}

func matchSettlementNameBackward(tokens []Token, end int) (nameEnd, nameStart int, ok bool) {
	if end < 0 || !isSettlementNameToken(tokens[end]) {
		return 0, 0, false
	}
	start := end
	count := 1
	for i := end - 1; i >= 0 && count < maxSettlementNameTokens; i-- {
		if !addressGapOK(tokens, i, i+1) {
			break
		}
		if !isSettlementNameToken(tokens[i]) {
			break
		}
		start = i
		count++
	}
	return end, start, true
}

func isSettlementNameToken(tok Token) bool {
	if tok.Kind != KindWord {
		return false
	}
	return tok.Capitalized || tok.Hyphenated
}

func matchAddressLabel(tokens []Token, idx int, label addressLabel) (next int, ok bool) {
	if idx >= len(tokens) {
		return idx, false
	}
	tok := tokens[idx]
	if tok.Kind != KindWord && tok.Kind != KindOther {
		return idx, false
	}
	folded := tok.Folded
	if label.full != "" && folded == label.full {
		return idx + 1, true
	}
	if label.short != "" && folded == label.short {
		next := idx + 1
		if next < len(tokens) && tokens[next].Kind == KindPunctuation && tokens[next].Text == "." {
			return next + 1, true
		}
		return next, true
	}
	if label.short != "" && strings.HasSuffix(tok.Text, ".") && folded == label.short+"." {
		return idx + 1, true
	}
	if label.short == "" && label.full != "" && folded == label.full {
		return idx + 1, true
	}
	return idx, false
}

func isHouseIdentifierToken(tok Token) bool {
	switch tok.Kind {
	case KindInteger, KindWord, KindOther:
		return validHouseIdentifier(tok.Text)
	default:
		return false
	}
}

func validHouseIdentifier(segment string) bool {
	if segment == "" {
		return false
	}
	if strings.Contains(segment, "/") {
		if strings.Contains(segment, "-") {
			return false
		}
		parts := strings.Split(segment, "/")
		if len(parts) != 2 {
			return false
		}
		return validHouseIdentifierPart(parts[0]) && validHouseIdentifierPart(parts[1])
	}
	if strings.Contains(segment, "-") {
		parts := strings.Split(segment, "-")
		if len(parts) != 2 {
			return false
		}
		return validHouseIdentifierPart(parts[0]) && validHouseIdentifierPart(parts[1])
	}
	return validHouseIdentifierPart(segment)
}

func validHouseIdentifierPart(part string) bool {
	if part == "" {
		return false
	}
	digits := 0
	letters := 0
	for _, r := range part {
		switch {
		case unicode.IsDigit(r):
			digits++
		case unicode.IsLetter(r) && (r <= unicode.MaxASCII || isCyrillicLetter(r)):
			letters++
			if letters > 1 {
				return false
			}
		default:
			return false
		}
	}
	return digits > 0
}

func isAddressLabelWord(folded string) bool {
	for _, label := range streetPrefixLabels {
		if folded == label.full || folded == label.short {
			return true
		}
	}
	for _, label := range settlementLabels {
		if folded == label.full || folded == label.short {
			return true
		}
	}
	for _, tables := range [][]addressLabel{houseLabels, corpusLabels, buildingLabels, apartmentLabels} {
		for _, label := range tables {
			if folded == label.full || folded == label.short {
				return true
			}
		}
	}
	return false
}

func looksLikeExtendedAddressPart(tokens []Token, idx int) bool {
	if idx >= len(tokens) {
		return false
	}
	peek, ok := skipOptionalCommaSeparator(tokens, idx)
	if !ok {
		return true
	}
	for _, labels := range [][]addressLabel{corpusLabels, buildingLabels, apartmentLabels} {
		if _, ok := matchAddressLabel(tokens, peek, labels[0]); ok {
			return true
		}
	}
	return false
}

func skipOptionalCommaSeparator(tokens []Token, idx int) (int, bool) {
	if idx >= len(tokens) {
		return idx, true
	}
	if tokens[idx].Kind != KindPunctuation || tokens[idx].Text != "," {
		return idx, true
	}
	idx++
	if idx < len(tokens) && tokens[idx].Kind == KindPunctuation && tokens[idx].Text == "," {
		return 0, false
	}
	return idx, true
}

func addressGapOK(tokens []Token, left, right int) bool {
	if left < 0 || right >= len(tokens) {
		return false
	}
	if tokens[left].End > tokens[right].Start {
		return tokens[left].End == tokens[right].Start
	}
	return tokens[left].End <= tokens[right].Start
}

func addressPartGapOK(tokens []Token, left, right int) bool {
	if !addressGapOK(tokens, left, right) {
		return false
	}
	if tokens[left].End == tokens[right].Start {
		return true
	}
	// Allow comma and whitespace between address parts.
	return true
}
