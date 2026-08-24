package evaluation

import (
	"sort"
)

// byteInterval is a half-open UTF-8 byte range [start, end).
type byteInterval struct {
	start int
	end   int
}

func mergeIntervals(intervals []byteInterval) []byteInterval {
	if len(intervals) == 0 {
		return nil
	}
	sorted := append([]byteInterval(nil), intervals...)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].start != sorted[j].start {
			return sorted[i].start < sorted[j].start
		}
		return sorted[i].end < sorted[j].end
	})
	out := []byteInterval{sorted[0]}
	for _, iv := range sorted[1:] {
		last := &out[len(out)-1]
		if iv.start <= last.end {
			if iv.end > last.end {
				last.end = iv.end
			}
			continue
		}
		out = append(out, iv)
	}
	return out
}

func unionByteCount(intervals []byteInterval) int {
	merged := mergeIntervals(intervals)
	total := 0
	for _, iv := range merged {
		total += iv.end - iv.start
	}
	return total
}

func intersectionByteCount(a, b []byteInterval) int {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	ma := mergeIntervals(a)
	mb := mergeIntervals(b)
	total := 0
	i, j := 0, 0
	for i < len(ma) && j < len(mb) {
		start := ma[i].start
		if mb[j].start > start {
			start = mb[j].start
		}
		end := ma[i].end
		if mb[j].end < end {
			end = mb[j].end
		}
		if start < end {
			total += end - start
		}
		if ma[i].end < mb[j].end {
			i++
		} else if mb[j].end < ma[i].end {
			j++
		} else {
			i++
			j++
		}
	}
	return total
}

func differenceByteCount(a, b []byteInterval) int {
	return unionByteCount(a) - intersectionByteCount(a, b)
}

func intervalDifference(a, b []byteInterval) []byteInterval {
	if len(a) == 0 {
		return nil
	}
	mergedA := mergeIntervals(a)
	mergedB := mergeIntervals(b)
	var out []byteInterval
	for _, iv := range mergedA {
		out = append(out, subtractInterval(iv, mergedB)...)
	}
	return mergeIntervals(out)
}

func subtractInterval(iv byteInterval, subtract []byteInterval) []byteInterval {
	if len(subtract) == 0 {
		return []byteInterval{iv}
	}
	var out []byteInterval
	pos := iv.start
	for _, sub := range subtract {
		if sub.end <= pos {
			continue
		}
		if sub.start > pos {
			out = append(out, byteInterval{start: pos, end: minInt(sub.start, iv.end)})
		}
		if sub.end >= iv.end {
			return out
		}
		pos = sub.end
	}
	if pos < iv.end {
		out = append(out, byteInterval{start: pos, end: iv.end})
	}
	return out
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func spanFullyCovered(span byteInterval, covered []byteInterval) bool {
	if span.start >= span.end {
		return true
	}
	merged := mergeIntervals(covered)
	pos := span.start
	for _, iv := range merged {
		if iv.end <= pos {
			continue
		}
		if iv.start > pos {
			return false
		}
		if iv.end >= span.end {
			return true
		}
		pos = iv.end
	}
	return pos >= span.end
}

func intervalsOverlap(a, b byteInterval) bool {
	return a.start < b.end && b.start < a.end
}

func intervalsEqual(a, b byteInterval) bool {
	return a.start == b.start && a.end == b.end
}
