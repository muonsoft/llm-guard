// Package llmguard provides a detection-only core for finding sensitive entities
// in text using pluggable detectors.
//
// Findings use UTF-8 byte offsets: Start and End form a half-open interval
// [Start, End) into the original input string. A span is valid only when both
// boundaries align with UTF-8 rune boundaries; End may equal len(text).
//
// Guard is immutable after construction and safe for concurrent Detect calls.
// Each registered Detector must also be safe for concurrent use because a single
// Guard may run detectors in parallel across overlapping Detect invocations.
//
// Masking, restore, and built-in detectors are out of scope for this package;
// callers own any state needed for downstream processing.
package llmguard
