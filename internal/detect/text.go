package detect

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

func digitTokenBoundaryOK(text string, start, end int) bool {
	if start > 0 {
		if r, _ := precedingRune(text, start); unicode.IsLetter(r) || unicode.IsDigit(r) {
			return false
		}
	}
	if end < len(text) {
		if r, _ := utf8RuneAt(text, end); unicode.IsLetter(r) || unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

func numericIdentifierBoundaryOK(text string, start, end int) bool {
	if !digitTokenBoundaryOK(text, start, end) {
		return false
	}
	if start > 0 {
		r, size := precedingRune(text, start)
		if isNumericExtensionSeparator(r) {
			prev, prevSize := precedingRune(text, start-size)
			if prevSize > 0 && prev >= '0' && prev <= '9' {
				return false
			}
		}
	}
	if end < len(text) {
		r, size := utf8RuneAt(text, end)
		if isNumericExtensionSeparator(r) {
			next, nextSize := utf8RuneAt(text, end+size)
			if nextSize > 0 && next >= '0' && next <= '9' {
				return false
			}
		}
	}
	return true
}

func isNumericExtensionSeparator(r rune) bool {
	return r == '-' || r == '.' || r == '_'
}

func asciiTokenBoundaryOK(text string, start, end int, extraInner func(rune) bool) bool {
	if start > 0 {
		if r, _ := precedingRune(text, start); !isOuterBoundary(r, extraInner) {
			return false
		}
	}
	if end < len(text) {
		if r, _ := utf8RuneAt(text, end); !isOuterBoundary(r, extraInner) {
			return false
		}
	}
	return true
}

func isOuterBoundary(r rune, extraInner func(rune) bool) bool {
	if unicode.IsLetter(r) || unicode.IsDigit(r) {
		return false
	}
	if extraInner != nil && extraInner(r) {
		return false
	}
	return true
}

func collectASCIIDigits(segment string) string {
	var b strings.Builder
	b.Grow(len(segment))
	for _, r := range segment {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func allSameDigits(digits string) bool {
	if digits == "" {
		return false
	}
	first := digits[0]
	for i := 1; i < len(digits); i++ {
		if digits[i] != first {
			return false
		}
	}
	return true
}

func separatorKind(r rune) int {
	switch r {
	case ' ', '\t':
		return 1
	case '-':
		return 2
	default:
		return 0
	}
}

func consistentDigitSeparators(segment string) bool {
	kind := 0
	for _, r := range segment {
		switch separatorKind(r) {
		case 0:
			continue
		case 1, 2:
			if kind == 0 {
				kind = separatorKind(r)
				continue
			}
			if separatorKind(r) != kind {
				return false
			}
		default:
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
