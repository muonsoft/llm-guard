package evaluation

import (
	"unicode"
	"unicode/utf8"
)

// RuneIntervalToUTF8 converts a [start,end) Unicode code-point interval to UTF-8 byte offsets.
func RuneIntervalToUTF8(text string, runeStart, runeEnd int) (byteStart, byteEnd int, ok bool) {
	if runeStart < 0 || runeEnd < runeStart {
		return 0, 0, false
	}
	byteStart = -1
	byteEnd = -1
	runeIdx := 0
	for offset := 0; offset < len(text); {
		if runeIdx == runeStart {
			byteStart = offset
		}
		if runeIdx == runeEnd {
			byteEnd = offset
			return byteStart, byteEnd, byteStart >= 0
		}
		_, size := utf8.DecodeRuneInString(text[offset:])
		offset += size
		runeIdx++
	}
	if runeIdx == runeEnd {
		byteEnd = len(text)
		return byteStart, byteEnd, byteStart >= 0
	}
	return 0, 0, false
}

// LocateTokenInText finds token in remaining text: exact prefix or skip Unicode whitespace.
func LocateTokenInText(text string, token string) (byteStart, byteEnd int, rest string, ok bool) {
	if stringsHasPrefix(text, token) {
		return 0, len(token), text[len(token):], true
	}
	pos := 0
	for pos < len(text) && isUnicodeSpace(text[pos:]) {
		_, size := utf8.DecodeRuneInString(text[pos:])
		pos += size
	}
	restAfterSpace := text[pos:]
	if stringsHasPrefix(restAfterSpace, token) {
		return pos, pos + len(token), restAfterSpace[len(token):], true
	}
	return 0, 0, text, false
}

func stringsHasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

func isUnicodeSpace(s string) bool {
	r, _ := utf8.DecodeRuneInString(s)
	return unicode.IsSpace(r)
}
