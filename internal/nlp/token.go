package nlp

import (
	"strings"
	"unicode"

	"github.com/muonsoft/go-razdel"
)

// Kind is a closed token kind used by matchers.
type Kind uint8

const (
	KindWord Kind = iota
	KindPunctuation
	KindInteger
	KindOther
)

// Token is an annotated source token with original UTF-8 byte offsets.
type Token struct {
	Text        string
	Folded      string
	Start       int
	End         int
	Kind        Kind
	Role        Role
	Forms       FormClass
	Capitalized bool
	Initial     bool
	Hyphenated  bool
}

// Span is a matched PERSON byte interval in source text.
type Span struct {
	Start int
	End   int
}

// Tokenize adapts go-razdel output into annotated tokens preserving byte spans.
func Tokenize(text string) []Token {
	raw := razdel.Tokenize(text)
	if len(raw) == 0 {
		return nil
	}

	out := make([]Token, 0, len(raw))
	for i := 0; i < len(raw); i++ {
		tok := raw[i]
		segment := text[tok.Start:tok.End]
		if segment != tok.Text {
			continue
		}

		if isSingleLetterWord(segment) && i+1 < len(raw) {
			next := raw[i+1]
			if isByteAdjacentDot(text, tok, next) {
				out = append(out, annotateInitial(tok.Start, next.End, segment))
				i++
				continue
			}
		}

		out = append(out, annotateToken(tok.Start, tok.End, segment, classifyRazdelToken(segment)))
	}
	return out
}

func classifyRazdelToken(segment string) Kind {
	if segment == "" {
		return KindOther
	}
	allDigits := true
	allPunct := true
	for _, r := range segment {
		if unicode.IsDigit(r) {
			allPunct = false
			continue
		}
		allDigits = false
		if !unicode.IsPunct(r) {
			allPunct = false
		}
	}
	if allDigits {
		return KindInteger
	}
	if allPunct {
		return KindPunctuation
	}
	if isWordToken(segment) {
		return KindWord
	}
	return KindOther
}

func annotateToken(start, end int, segment string, kind Kind) Token {
	token := Token{
		Text:   segment,
		Folded: foldToken(segment),
		Start:  start,
		End:    end,
		Kind:   kind,
	}
	if kind != KindWord {
		return token
	}
	if strings.Contains(segment, "-") {
		token.Hyphenated = true
	}
	if !isCapitalizedCyrillicWord(segment) {
		return token
	}
	token.Capitalized = true
	role, forms, ok := lookupNameRole(token.Folded)
	if ok {
		token.Role = role
		token.Forms = forms
	}
	return token
}

func annotateInitial(letterStart, dotEnd int, letter string) Token {
	segment := letter + "."
	return Token{
		Text:        segment,
		Folded:      foldToken(letter),
		Start:       letterStart,
		End:         dotEnd,
		Kind:        KindWord,
		Capitalized: true,
		Initial:     true,
	}
}

func isByteAdjacentDot(text string, letterTok, dotTok razdel.Token) bool {
	if dotTok.Start != letterTok.End {
		return false
	}
	segment := text[dotTok.Start:dotTok.End]
	return segment == "." && dotTok.Text == "."
}

func isSingleLetterWord(segment string) bool {
	runes := []rune(segment)
	return len(runes) == 1 && isCyrillicUpper(runes[0])
}

func isWordToken(segment string) bool {
	if segment == "" {
		return false
	}
	for _, r := range segment {
		if !unicode.IsLetter(r) && r != '-' && r != '\'' {
			return false
		}
	}
	return true
}

func isCapitalizedCyrillicWord(segment string) bool {
	runes := []rune(segment)
	if len(runes) == 0 {
		return false
	}
	if !isCyrillicUpper(runes[0]) {
		return false
	}
	for i := 1; i < len(runes); i++ {
		r := runes[i]
		if r == '-' {
			if i+1 >= len(runes) || !isCyrillicUpper(runes[i+1]) {
				return false
			}
			continue
		}
		if isCyrillicLetter(r) && isCyrillicUpper(r) {
			return false
		}
		if !isCyrillicLetter(r) && !unicode.IsLower(r) {
			return false
		}
	}
	return true
}

func isCyrillicLetter(r rune) bool {
	return (r >= 'А' && r <= 'я') || r == 'ё' || r == 'Ё'
}

func isCyrillicUpper(r rune) bool {
	return (r >= 'А' && r <= 'Я') || r == 'Ё'
}

func foldToken(segment string) string {
	return strings.ToLower(segment)
}
