package llmguard

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"unicode/utf8"
)

func FuzzEmailDetectorBoundaries(f *testing.F) {
	f.Add("a@b.co")
	f.Add("Напишите (Ivan.Petrov+sales@example-domain.ru).")
	f.Add("user@example.company")
	f.Add("")

	f.Fuzz(func(t *testing.T, text string) {
		if !utf8.ValidString(text) {
			return
		}

		detector := NewEmailDetector()
		findings, err := detector.Detect(context.Background(), text)
		if err != nil {
			return
		}

		for _, finding := range findings {
			if finding.Start < 0 || finding.End <= finding.Start || finding.End > len(text) {
				t.Fatalf("invalid span [%d,%d) for len=%d", finding.Start, finding.End, len(text))
			}
			if !utf8.RuneStart(text[finding.Start]) {
				t.Fatalf("start not on rune boundary at %d", finding.Start)
			}
			if finding.End < len(text) && !utf8.RuneStart(text[finding.End]) {
				t.Fatalf("end not on rune boundary at %d", finding.End)
			}
			if finding.Entity != EntityEmail {
				t.Fatalf("unexpected entity %q", finding.Entity)
			}
			mailbox := text[finding.Start:finding.End]
			if !validateEmailMailbox(mailbox) {
				t.Fatalf("invalid mailbox grammar for %q", mailbox)
			}
			if !emailBoundaryOK(text, finding.Start, finding.End) {
				t.Fatalf("boundary check failed for %q", mailbox)
			}
		}
	})
}

func FuzzResolveInvariants(f *testing.F) {
	f.Add("contact a@b.co", byte(0))
	f.Add("0123456789", byte(1))
	f.Add("café", byte(2))

	f.Fuzz(func(t *testing.T, text string, seed byte) {
		if !utf8.ValidString(text) {
			return
		}
		if len(text) == 0 {
			return
		}

		findings := syntheticFindings(text, int(seed))
		if len(findings) == 0 {
			return
		}

		inputCopy := append([]Finding(nil), findings...)
		resolved, err := Resolve(text, findings)
		if err != nil {
			t.Fatalf("resolve failed: %v", err)
		}
		if !findingsEqual(findings, inputCopy) {
			t.Fatalf("resolve mutated input findings slice")
		}

		for i, finding := range resolved {
			if _, validateErr := validateFinding(text, finding.Detector, finding, i); validateErr != nil {
				t.Fatalf("resolved finding %d invalid: %v", i, validateErr)
			}
		}

		for i := 1; i < len(resolved); i++ {
			prev, cur := resolved[i-1], resolved[i]
			if prev.Start > cur.Start ||
				(prev.Start == cur.Start && prev.End > cur.End) ||
				(prev.Start == cur.Start && prev.End == cur.End && prev.Entity > cur.Entity) ||
				(prev.Start == cur.Start && prev.End == cur.End && prev.Entity == cur.Entity && prev.Detector > cur.Detector) ||
				(prev.Start == cur.Start && prev.End == cur.End && prev.Entity == cur.Entity && prev.Detector == cur.Detector && prev.Confidence < cur.Confidence) {
				t.Fatalf("resolved findings not in stable textual order at %d", i)
			}
		}

		for i := 0; i < len(resolved); i++ {
			for j := i + 1; j < len(resolved); j++ {
				if intervalsOverlap(resolved[i].Start, resolved[i].End, resolved[j].Start, resolved[j].End) {
					t.Fatalf("overlap between %d and %d", i, j)
				}
			}
		}

		again, err := Resolve(text, shuffleFindings(inputCopy, int(seed)))
		if err != nil {
			t.Fatalf("permutation resolve failed: %v", err)
		}
		if !findingsEqual(resolved, again) {
			t.Fatalf("non-deterministic resolve output")
		}
	})
}

func FuzzMaskRestoreRoundTrip(f *testing.F) {
	f.Add("a@b.co")
	f.Add("Привет a@b.co мир")
	f.Add("a@b.co and c@d.org")

	f.Fuzz(func(t *testing.T, text string) {
		if !utf8.ValidString(text) {
			return
		}

		reader := &deterministicReader{seed: []byte(text)}
		guard, err := New(WithDetector(NewEmailDetector()), WithRandomSource(reader))
		if err != nil {
			t.Fatalf("new guard: %v", err)
		}

		result, err := guard.Mask(context.Background(), text)
		if err != nil {
			return
		}

		restored, err := guard.Restore(context.Background(), result.Text, result.Tokens)
		if err != nil {
			t.Fatalf("restore: %v", err)
		}
		if restored != text {
			t.Fatalf("round trip mismatch")
		}
	})
}

func syntheticFindings(text string, seed int) []Finding {
	boundaries := runeBoundaryIndices(text)
	if len(boundaries) < 2 {
		return nil
	}

	max := len(boundaries) - 1
	if max > 6 {
		max = 6
	}

	out := make([]Finding, 0, max)
	for i := 0; i < max; i++ {
		startIdx := (seed + i*3) % (len(boundaries) - 1)
		endIdx := startIdx + 1 + ((seed + i) % 2)
		if endIdx >= len(boundaries) {
			endIdx = len(boundaries) - 1
		}
		if endIdx <= startIdx {
			continue
		}

		start := boundaries[startIdx]
		end := boundaries[endIdx]
		if end <= start {
			continue
		}

		out = append(out, Finding{
			Entity:     EntityType(string(rune('A' + (i % 3)))),
			Start:      start,
			End:        end,
			Confidence: float64(i%10) / 10,
			Detector:   "fuzz",
		})
	}
	return out
}

func runeBoundaryIndices(text string) []int {
	indices := []int{0}
	for i := 0; i < len(text); {
		_, size := utf8.DecodeRuneInString(text[i:])
		if size == 0 {
			break
		}
		i += size
		indices = append(indices, i)
	}
	return indices
}

func shuffleFindings(findings []Finding, seed int) []Finding {
	if len(findings) <= 1 {
		return append([]Finding(nil), findings...)
	}
	out := append([]Finding(nil), findings...)
	for i := len(out) - 1; i > 0; i-- {
		j := (seed + i*7) % (i + 1)
		out[i], out[j] = out[j], out[i]
	}
	return out
}

func findingsEqual(a, b []Finding) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

type deterministicReader struct {
	seed []byte
	pos  int
}

func (r *deterministicReader) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	for i := range p {
		if len(r.seed) == 0 {
			p[i] = byte(i)
			continue
		}
		p[i] = r.seed[r.pos%len(r.seed)]
		r.pos++
	}
	return len(p), nil
}

func TestFuzzHelpers_DeterministicReader_ExpectStable(t *testing.T) {
	reader := &deterministicReader{seed: []byte("abc")}
	buf := make([]byte, 4)
	_, err := reader.Read(buf)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(buf, []byte{'a', 'b', 'c', 'a'}) {
		t.Fatalf("unexpected %v", buf)
	}
}

func TestFuzzHelpers_ShuffleFindings_ExpectPermutation(t *testing.T) {
	original := []Finding{{Start: 1}, {Start: 2}, {Start: 3}}
	shuffled := shuffleFindings(original, 3)
	if len(shuffled) != len(original) {
		t.Fatal("length mismatch")
	}
}

func TestFuzzHelpers_SyntheticFindings_ExpectWithinText(t *testing.T) {
	findings := syntheticFindings("abcdef", 2)
	for _, finding := range findings {
		if finding.End > len("abcdef") {
			t.Fatalf("span out of range")
		}
		if !utf8.RuneStart("abcdef"[finding.Start]) {
			t.Fatalf("start not on rune boundary")
		}
		if finding.End < len("abcdef") && !utf8.RuneStart("abcdef"[finding.End]) {
			t.Fatalf("end not on rune boundary")
		}
	}
}

func TestFuzzHelpers_EmailMailboxValidation_ExpectAsciiOnly(t *testing.T) {
	if validateEmailMailbox("a@b.co") != true {
		t.Fatal("expected valid mailbox")
	}
	if validateEmailMailbox(strings.Repeat("a", 50)+"@b.co") != true {
		t.Fatal("expected long local part valid")
	}
	if validateEmailMailbox("user@example.company") != true {
		t.Fatal("expected company suffix valid")
	}
}
