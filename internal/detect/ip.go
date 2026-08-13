package detect

import (
	"context"
	"net"
	"regexp"
	"sort"
	"strings"
)

const ipDetectorName = "ip"

var (
	ipv4CandidatePattern = regexp.MustCompile(`(?:[0-9]{1,3}\.){3}[0-9]{1,3}`)
	ipv6CandidatePattern = regexp.MustCompile(`(?i)(?:[0-9a-f]{0,4}:){1,6}(?:[0-9a-f]{0,4}|(?:(?:[0-9]{1,3}\.){3}[0-9]{1,3}))`)
)

func IP(ctx context.Context, text string) ([]Span, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	candidates := make([]Span, 0, 4)

	bracketed := findBracketedIPv6(text)
	for _, loc := range bracketed {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		candidates = append(candidates, Span{Start: loc[0], End: loc[1]})
	}

	ipv6Matches := ipv6CandidatePattern.FindAllStringIndex(text, -1)
	for _, loc := range ipv6Matches {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		start, end := loc[0], loc[1]
		if isInsideBracketedIPv6(text, start, end) {
			continue
		}
		segment := text[start:end]
		if strings.Contains(segment, "%") || (end < len(text) && text[end] == '%') {
			continue
		}
		end = extendIPv6End(text, start, end)
		if !ipAddressBoundaryOK(text, start, end) {
			continue
		}
		segment = text[start:end]
		if !validateIPv6Segment(segment) {
			continue
		}
		candidates = append(candidates, Span{Start: start, End: end})
	}

	ipv4Matches := ipv4CandidatePattern.FindAllStringIndex(text, -1)
	for _, loc := range ipv4Matches {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		start, end := loc[0], loc[1]
		if spanInsideAny(start, end, candidates) {
			continue
		}
		if !ipAddressBoundaryOK(text, start, end) {
			continue
		}
		segment := text[start:end]
		if !validateIPv4Segment(segment) {
			continue
		}
		candidates = append(candidates, Span{Start: start, End: end})
	}

	findings := selectNonOverlappingIPFindings(candidates)
	if len(findings) == 0 {
		return nil, nil
	}
	return findings, nil
}

func extendIPv6End(text string, start, end int) int {
	for end < len(text) {
		candidate := text[start:end]
		if !validateIPv6Segment(candidate) {
			break
		}
		next := end
		if next < len(text) && (text[next] == ':' || text[next] == '.') {
			extended := end + 1
			for extended < len(text) {
				ch := text[extended]
				if (ch >= '0' && ch <= '9') || ch == '.' || ch == ':' || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F') {
					extended++
					continue
				}
				break
			}
			if validateIPv6Segment(text[start:extended]) {
				end = extended
				continue
			}
		}
		break
	}
	return end
}

func ipAddressBoundaryOK(text string, start, end int) bool {
	if !digitTokenBoundaryOK(text, start, end) {
		return false
	}
	if end < len(text) {
		switch text[end] {
		case '.':
			if end+1 < len(text) && text[end+1] >= '0' && text[end+1] <= '9' {
				return false
			}
		case ':':
			return false
		}
	}
	return true
}

func spanInsideAny(start, end int, findings []Span) bool {
	for _, f := range findings {
		if start >= f.Start && end <= f.End {
			return true
		}
	}
	return false
}

func selectNonOverlappingIPFindings(candidates []Span) []Span {
	if len(candidates) == 0 {
		return nil
	}
	sort.Slice(candidates, func(i, j int) bool {
		li, lj := candidates[i].End-candidates[i].Start, candidates[j].End-candidates[j].Start
		if li != lj {
			return li > lj
		}
		if candidates[i].Start != candidates[j].Start {
			return candidates[i].Start < candidates[j].Start
		}
		return candidates[i].End > candidates[j].End
	})

	selected := make([]Span, 0, len(candidates))
	for _, candidate := range candidates {
		if spanInsideAny(candidate.Start, candidate.End, selected) {
			continue
		}
		overlap := false
		for _, existing := range selected {
			if intervalsOverlap(candidate.Start, candidate.End, existing.Start, existing.End) {
				overlap = true
				break
			}
		}
		if overlap {
			continue
		}
		selected = append(selected, candidate)
	}

	sort.Slice(selected, func(i, j int) bool {
		if selected[i].Start != selected[j].Start {
			return selected[i].Start < selected[j].Start
		}
		return selected[i].End < selected[j].End
	})
	return selected
}

func validateIPv4Segment(segment string) bool {
	parts := strings.Split(segment, ".")
	if len(parts) != 4 {
		return false
	}
	for _, part := range parts {
		if len(part) == 0 || len(part) > 3 {
			return false
		}
		if len(part) > 1 && part[0] == '0' {
			return false
		}
		for _, r := range part {
			if r < '0' || r > '9' {
				return false
			}
		}
	}
	ip := net.ParseIP(segment)
	if ip == nil || ip.To4() == nil {
		return false
	}
	return true
}

func validateIPv6Segment(segment string) bool {
	if !strings.Contains(segment, ":") {
		return false
	}
	if strings.Contains(segment, "%") {
		return false
	}
	ip := net.ParseIP(segment)
	if ip == nil {
		return false
	}
	return ip.To16() != nil
}

func findBracketedIPv6(text string) [][2]int {
	var out [][2]int
	for i := 0; i < len(text); i++ {
		if text[i] != '[' {
			continue
		}
		close := strings.IndexByte(text[i+1:], ']')
		if close < 0 {
			continue
		}
		close += i + 1
		innerStart := i + 1
		innerEnd := close
		segment := text[innerStart:innerEnd]
		if !validateIPv6Segment(segment) {
			continue
		}
		if !digitTokenBoundaryOK(text, innerStart, innerEnd) {
			continue
		}
		out = append(out, [2]int{innerStart, innerEnd})
	}
	return out
}

func isInsideBracketedIPv6(text string, start, end int) bool {
	for i := 0; i < len(text); i++ {
		if text[i] != '[' {
			continue
		}
		close := strings.IndexByte(text[i+1:], ']')
		if close < 0 {
			continue
		}
		close += i + 1
		innerStart := i + 1
		innerEnd := close
		if start >= innerStart && end <= innerEnd {
			return true
		}
	}
	return false
}
