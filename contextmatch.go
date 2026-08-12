package llmguard

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

const ruContextGapMaxBytes = 48

var ruContextGapWords = []string{"рф", "серия", "номер"}

func lineStartIndex(text string, index int) int {
	for index > 0 && text[index-1] != '\n' {
		index--
	}
	return index
}

func lineEndIndex(text string, index int) int {
	for index < len(text) && text[index] != '\n' {
		index++
	}
	return index
}

func hasBoundedRUContext(text string, candidateStart int, markers []string) bool {
	lineStart := lineStartIndex(text, candidateStart)
	prefix := text[lineStart:candidateStart]
	for _, marker := range markers {
		idx := lastCaseInsensitiveIndex(prefix, marker)
		if idx < 0 {
			continue
		}
		if !isValidMarkerOccurrence(prefix, idx, marker) {
			continue
		}
		markerEnd := lineStart + byteIndexAfterCaseInsensitiveMatch(prefix, idx, marker)
		gap := text[markerEnd:candidateStart]
		if len(gap) > ruContextGapMaxBytes {
			continue
		}
		if validRUContextGap(gap) {
			return true
		}
	}
	return false
}

func byteIndexAfterCaseInsensitiveMatch(text string, start int, marker string) int {
	end := start
	remaining := marker
	for len(remaining) > 0 {
		if end >= len(text) {
			return start
		}
		mr, mSize := utf8.DecodeRuneInString(remaining)
		tr, tSize := utf8.DecodeRuneInString(text[end:])
		if unicode.ToLower(mr) != unicode.ToLower(tr) {
			return start
		}
		end += tSize
		remaining = remaining[mSize:]
	}
	return end
}

func isValidMarkerOccurrence(text string, idx int, marker string) bool {
	if idx > 0 {
		prev, _ := precedingRune(text, idx)
		if unicode.IsLetter(prev) || unicode.IsDigit(prev) {
			return false
		}
	}
	markerEnd := byteIndexAfterCaseInsensitiveMatch(text, idx, marker)
	if markerEnd == idx {
		return false
	}
	if markerEnd < len(text) {
		next, _ := utf8RuneAt(text, markerEnd)
		if unicode.IsLetter(next) {
			return false
		}
	}
	return true
}

func lastCaseInsensitiveIndex(text, substr string) int {
	if substr == "" || len(text) < len(substr) {
		return -1
	}
	return strings.LastIndex(strings.ToLower(text), strings.ToLower(substr))
}

func validRUContextGap(gap string) bool {
	if gap == "" {
		return true
	}
	i := 0
	for i < len(gap) {
		r, size := utf8.DecodeRuneInString(gap[i:])
		if size == 0 {
			break
		}
		if unicode.IsSpace(r) {
			i += size
			continue
		}
		switch r {
		case ':', '№', '#', '-', ',':
			i += size
			continue
		}
		if unicode.IsLetter(r) {
			wordStart := i
			wordEnd := i + size
			for wordEnd < len(gap) {
				nr, nsize := utf8.DecodeRuneInString(gap[wordEnd:])
				if nsize == 0 || !unicode.IsLetter(nr) {
					break
				}
				wordEnd += nsize
			}
			word := strings.ToLower(gap[wordStart:wordEnd])
			if !isAllowedRUContextWord(word) {
				return false
			}
			i = wordEnd
			continue
		}
		return false
	}
	return true
}

func isAllowedRUContextWord(word string) bool {
	for _, allowed := range ruContextGapWords {
		if word == allowed {
			return true
		}
	}
	return false
}
