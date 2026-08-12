package nlp

import (
	"context"
	"strings"
	"unicode"
	"unicode/utf8"
)

var streetMarkers = []string{
	"улица",
	"ул",
	"проспект",
	"пр-т",
	"просп",
	"переулок",
	"пер",
	"бульвар",
	"б-р",
	"шоссе",
	"площадь",
	"пл",
}

// DetectPersonSpans finds conservative PERSON spans in text.
func DetectPersonSpans(ctx context.Context, text string) ([]Span, error) {
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
		if span, consumed, ok := matchPersonAt(text, tokens, i); ok {
			spans = append(spans, span)
			i += consumed
			continue
		}
		i++
	}
	return spans, nil
}

func matchPersonAt(text string, tokens []Token, start int) (Span, int, bool) {
	type candidate struct {
		end      int
		consumed int
	}

	var best candidate
	try := func(endIdx int, consumed int, roles []Role, forms FormClass) {
		if endIdx <= start || consumed <= 0 {
			return
		}
		if !rolesCompatible(roles, forms) {
			return
		}
		spanStart := tokens[start].Start
		spanEnd := tokens[endIdx].End
		if spanStart >= spanEnd || strings.Contains(text[spanStart:spanEnd], "\n") {
			return
		}
		if hasStreetContext(tokens, start) {
			return
		}
		if !outerBoundariesOK(text, spanStart, spanEnd) {
			return
		}
		if consumed > best.consumed || (consumed == best.consumed && spanEnd > tokens[best.end].End) {
			best = candidate{end: endIdx, consumed: consumed}
		}
	}

	if end, ok := matchSequenceInText(text, tokens, start, []componentSpec{
		{want: RoleFirst, allowInitial: false},
		{want: RolePatronymic, allowInitial: false},
		{want: RoleSurname, allowInitial: false},
	}); ok {
		try(end, end-start+1, []Role{RoleFirst, RolePatronymic, RoleSurname}, mergeTokenForms(tokens[start:end+1]))
	}
	if end, ok := matchSequenceInText(text, tokens, start, []componentSpec{
		{want: RoleSurname, allowInitial: false},
		{want: RoleFirst, allowInitial: false},
		{want: RolePatronymic, allowInitial: false},
	}); ok {
		try(end, end-start+1, []Role{RoleSurname, RoleFirst, RolePatronymic}, mergeTokenForms(tokens[start:end+1]))
	}
	if end, ok := matchSequenceInText(text, tokens, start, []componentSpec{
		{want: RoleSurname, allowInitial: false},
		{want: RoleNone, allowInitial: true},
		{want: RoleNone, allowInitial: true},
	}); ok {
		try(end, end-start+1, []Role{RoleSurname, RoleNone, RoleNone}, 0)
	}
	if end, ok := matchSequenceInText(text, tokens, start, []componentSpec{
		{want: RoleNone, allowInitial: true},
		{want: RoleNone, allowInitial: true},
		{want: RoleSurname, allowInitial: false},
	}); ok {
		try(end, end-start+1, []Role{RoleNone, RoleNone, RoleSurname}, 0)
	}
	if end, ok := matchSequenceInText(text, tokens, start, []componentSpec{
		{want: RoleFirst, allowInitial: false},
		{want: RoleSurname, allowInitial: false},
	}); ok {
		try(end, end-start+1, []Role{RoleFirst, RoleSurname}, mergeTokenForms(tokens[start:end+1]))
	}
	if end, ok := matchSequenceInText(text, tokens, start, []componentSpec{
		{want: RoleSurname, allowInitial: false},
		{want: RoleFirst, allowInitial: false},
	}); ok {
		try(end, end-start+1, []Role{RoleSurname, RoleFirst}, mergeTokenForms(tokens[start:end+1]))
	}

	if best.consumed == 0 {
		return Span{}, 0, false
	}
	return Span{Start: tokens[start].Start, End: tokens[best.end].End}, best.consumed, true
}

type componentSpec struct {
	want         Role
	allowInitial bool
}

func matchSequenceInText(text string, tokens []Token, start int, specs []componentSpec) (int, bool) {
	idx := start
	for _, spec := range specs {
		if idx >= len(tokens) {
			return 0, false
		}
		if idx > start && !gapTextOK(text, tokens, idx-1, idx) {
			return 0, false
		}
		tok := tokens[idx]
		switch {
		case spec.allowInitial:
			if tok.Kind != KindWord || !tok.Initial {
				return 0, false
			}
		case spec.want == RoleNone:
			return 0, false
		default:
			if tok.Kind != KindWord || tok.Initial || tok.Role != spec.want {
				return 0, false
			}
		}
		idx++
	}
	return idx - 1, true
}

func gapOK(tokens []Token, left, right int) bool {
	if left < 0 || right >= len(tokens) {
		return false
	}
	return tokens[left].End <= tokens[right].Start
}

func gapTextOK(text string, tokens []Token, left, right int) bool {
	if !gapOK(tokens, left, right) {
		return false
	}
	gap := text[tokens[left].End:tokens[right].Start]
	if gap == "" {
		return true
	}
	for _, r := range gap {
		if r != ' ' && r != '\t' {
			return false
		}
	}
	return true
}

func mergeTokenForms(tokens []Token) FormClass {
	if len(tokens) == 0 {
		return 0
	}
	merged := tokens[0].Forms
	for _, tok := range tokens[1:] {
		if tok.Kind != KindWord || tok.Initial {
			continue
		}
		if merged == 0 || tok.Forms == 0 {
			continue
		}
		merged &= tok.Forms
		if merged == 0 {
			return 0
		}
	}
	return merged
}

func rolesCompatible(roles []Role, forms FormClass) bool {
	hasSurname := false
	hasFirst := false
	initialCount := 0
	for _, role := range roles {
		switch role {
		case RoleSurname:
			hasSurname = true
		case RoleFirst:
			hasFirst = true
		case RoleNone:
			initialCount++
		}
	}
	if !hasSurname {
		return false
	}
	if initialCount == 2 {
		return true
	}
	if !hasFirst {
		return false
	}
	return forms != 0
}

func hasStreetContext(tokens []Token, start int) bool {
	if start == 0 {
		return false
	}
	prev := tokens[start-1]
	if prev.Kind != KindWord && prev.Kind != KindPunctuation {
		return false
	}
	folded := strings.ToLower(prev.Text)
	for _, marker := range streetMarkers {
		if folded == marker {
			return true
		}
	}
	if start >= 2 {
		prev2 := tokens[start-2]
		if strings.ToLower(prev2.Text) == "ул" && prev.Text == "." {
			return true
		}
	}
	return false
}

func outerBoundariesOK(text string, start, end int) bool {
	if start > 0 {
		r, _ := utf8.DecodeLastRuneInString(text[:start])
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return false
		}
	}
	if end < len(text) {
		r, _ := utf8.DecodeRuneInString(text[end:])
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return false
		}
	}
	return true
}
