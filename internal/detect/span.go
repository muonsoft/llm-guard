package detect

// Span is a matched entity byte interval in the original UTF-8 input.
type Span struct {
	Start int
	End   int
}

func overlapsExisting(spans []Span, start, end int) bool {
	for _, span := range spans {
		if intervalsOverlap(start, end, span.Start, span.End) {
			return true
		}
	}
	return false
}

func intervalsOverlap(aStart, aEnd, bStart, bEnd int) bool {
	return aStart < bEnd && bStart < aEnd
}
