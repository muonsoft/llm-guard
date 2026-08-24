// Package llmguard provides local detection, reversible masking, and restore for
// sensitive entities in text using pluggable detectors.
//
// llm-guard is a precision-oriented prefilter: it reduces the risk of sending
// documented supported PII and secret forms to an LLM. It does not replace
// high-recall DLP, generic NER, or exhaustive data-loss prevention. Supported
// scope includes multi-token Russian FIO, compositional ADDRESS (street+house),
// checksum-valid structured identifiers, and conservative secret patterns.
// Unsupported examples include single given names, city-only addresses,
// checksum-invalid INN/SNILS, and arbitrary unknown credential shapes.
//
// # Import path and toolchain
//
// Canonical module path: github.com/muonsoft/llm-guard (Go 1.26.2+).
// External consumers must use only the public API; internal packages are not
// supported.
//
// # Byte spans
//
// Findings use UTF-8 byte offsets: Start and End form a half-open interval
// [Start, End) into the original input string. A span is valid only when both
// boundaries align with UTF-8 rune boundaries; End may equal len(text).
//
// # Concurrency and state
//
// Guard is immutable after construction and safe for concurrent Detect, Mask,
// and Restore calls. Each registered Detector must also be safe for concurrent
// use because a single Guard may run detectors in parallel across overlapping
// invocations. Reversible state lives only in caller-owned TokenSet values; the
// library does not persist mappings or conversation history.
//
// # Security boundary
//
// Default observer is a no-op. Safe observers expose bounded metadata only and
// must not log original text or TokenSet contents. Secrets block Mask by default
// (ErrBlocked); reversible secret masking requires explicit configuration.
// WithUnsafeDevelopmentObserver leaks sensitive data by design and is not for
// production use.
//
// # Restore semantics
//
// Restore substitutes original substrings byte-for-byte and does not perform
// morphological agreement when the LLM alters surrounding grammar.
package llmguard
