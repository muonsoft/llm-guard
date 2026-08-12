package llmguard

import (
	"context"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

const dateOfBirthDetectorName = "date_of_birth"

var (
	dateOfBirthNumericPattern = regexp.MustCompile(`\d{2}[./]\d{2}[./]\d{4}`)
	dateOfBirthTextPattern    = regexp.MustCompile(
		`(?i)\d{1,2}\s+(?:января|февраля|марта|апреля|мая|июня|июля|августа|сентября|октября|ноября|декабря)\s+\d{4}`,
	)
)

var dateOfBirthContextMarkers = []string{
	"дата рождения",
	"д.р.",
	"родился",
	"родилась",
}

var russianMonthGenitive = map[string]time.Month{
	"января":   time.January,
	"февраля":  time.February,
	"марта":    time.March,
	"апреля":   time.April,
	"мая":      time.May,
	"июня":     time.June,
	"июля":     time.July,
	"августа":  time.August,
	"сентября": time.September,
	"октября":  time.October,
	"ноября":   time.November,
	"декабря":  time.December,
}

type dateOfBirthDetector struct{}

// NewDateOfBirthDetector returns an immutable built-in DATE_OF_BIRTH detector.
func NewDateOfBirthDetector() Detector {
	return dateOfBirthDetector{}
}

func (dateOfBirthDetector) Name() string {
	return dateOfBirthDetectorName
}

func (dateOfBirthDetector) Detect(ctx context.Context, text string) ([]Finding, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var findings []Finding

	numericMatches := dateOfBirthNumericPattern.FindAllStringIndex(text, -1)
	for _, loc := range numericMatches {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		start, end := loc[0], loc[1]
		if overlapsExisting(findings, start, end) {
			continue
		}
		if !dateTokenBoundaryOK(text, start, end) {
			continue
		}
		segment := text[start:end]
		if !validateNumericDateOfBirth(segment) {
			continue
		}
		if !hasBoundedRUContext(text, start, dateOfBirthContextMarkers) {
			continue
		}
		findings = append(findings, Finding{
			Entity:     EntityDateOfBirth,
			Start:      start,
			End:        end,
			Confidence: 0.84,
			Detector:   dateOfBirthDetectorName,
		})
	}

	textMatches := dateOfBirthTextPattern.FindAllStringIndex(text, -1)
	for _, loc := range textMatches {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		start, end := loc[0], loc[1]
		if overlapsExisting(findings, start, end) {
			continue
		}
		if !dateTokenBoundaryOK(text, start, end) {
			continue
		}
		segment := text[start:end]
		if strings.ContainsAny(segment, "\n\r") {
			continue
		}
		if !validateTextualDateOfBirth(segment) {
			continue
		}
		if !hasBoundedRUContext(text, start, dateOfBirthContextMarkers) {
			continue
		}
		findings = append(findings, Finding{
			Entity:     EntityDateOfBirth,
			Start:      start,
			End:        end,
			Confidence: 0.84,
			Detector:   dateOfBirthDetectorName,
		})
	}

	if len(findings) == 0 {
		return nil, nil
	}
	return findings, nil
}

func dateTokenBoundaryOK(text string, start, end int) bool {
	if start > 0 {
		r, size := precedingRune(text, start)
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return false
		}
		if r == '.' || r == '/' {
			prev, prevSize := precedingRune(text, start-size)
			if prevSize > 0 && unicode.IsDigit(prev) {
				return false
			}
		}
	}
	if end < len(text) {
		r, size := utf8RuneAt(text, end)
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return false
		}
		if r == '.' || r == '/' {
			next, nextSize := utf8RuneAt(text, end+size)
			if nextSize > 0 && unicode.IsDigit(next) {
				return false
			}
		}
	}
	return true
}

func validateNumericDateOfBirth(segment string) bool {
	sep := '/'
	if strings.Contains(segment, ".") {
		sep = '.'
	}
	parts := strings.Split(segment, string(sep))
	if len(parts) != 3 {
		return false
	}
	if len(parts[2]) != 4 {
		return false
	}
	day, err := strconv.Atoi(parts[0])
	if err != nil {
		return false
	}
	month, err := strconv.Atoi(parts[1])
	if err != nil {
		return false
	}
	year, err := strconv.Atoi(parts[2])
	if err != nil {
		return false
	}
	return calendarDateValid(day, month, year)
}

func validateTextualDateOfBirth(segment string) bool {
	lower := strings.ToLower(segment)
	fields := strings.Fields(lower)
	if len(fields) != 3 {
		return false
	}
	day, err := strconv.Atoi(fields[0])
	if err != nil {
		return false
	}
	month, ok := russianMonthGenitive[fields[1]]
	if !ok {
		return false
	}
	year, err := strconv.Atoi(fields[2])
	if err != nil || len(fields[2]) != 4 {
		return false
	}
	return calendarDateValid(day, int(month), year)
}

func calendarDateValid(day, month, year int) bool {
	if year < 1 {
		return false
	}
	t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	return t.Year() == year && int(t.Month()) == month && t.Day() == day
}
