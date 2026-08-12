// Package llmguard provides local detection, reversible masking, and restore for
// sensitive entities in text using pluggable detectors.
//
// Findings use UTF-8 byte offsets: Start and End form a half-open interval
// [Start, End) into the original input string. A span is valid only when both
// boundaries align with UTF-8 rune boundaries; End may equal len(text).
//
// Guard is immutable after construction and safe for concurrent Detect, Mask,
// and Restore calls. Each registered Detector must also be safe for concurrent
// use because a single Guard may run detectors in parallel across overlapping
// invocations. Reversible state lives only in caller-owned TokenSet values.
package llmguard
