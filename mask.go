package llmguard

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"
)

const (
	namespaceEntropyBytes = 16
	maxNamespaceAttempts  = 32
	tokenPrefix           = "{{LLMG_"
	tokenSuffix           = "}}"
)

type randomSourceOption struct {
	source io.Reader
}

func (o randomSourceOption) apply(cfg *guardConfig) error {
	if isNilReader(o.source) {
		return newInvalidConfigError("random source is nil")
	}
	cfg.randomSource = o.source
	return nil
}

// WithRandomSource configures the entropy source used for mask namespaces. Reads
// are serialized; callers replacing the secure default must provide sufficient
// cryptographic quality.
func WithRandomSource(source io.Reader) Option {
	return randomSourceOption{source: source}
}

// MaskResult holds masked text, resolved findings for the original input, and a
// caller-owned TokenSet for restore.
type MaskResult struct {
	Text     string
	Findings []Finding
	Tokens   *TokenSet
}

// String returns a safe summary without masked text, findings, or token mappings.
func (r MaskResult) String() string {
	return fmt.Sprintf("llmguard.MaskResult{findings=%d tokens=%s}", len(r.Findings), redactedTokenSetSummary())
}

// GoString returns the same safe summary as String.
func (r MaskResult) GoString() string {
	return r.String()
}

// MarshalJSON returns a safe JSON representation without sensitive values.
func (r MaskResult) MarshalJSON() ([]byte, error) {
	type safeMaskResult struct {
		Text     string `json:"text"`
		Findings int    `json:"findings_count"`
		Tokens   string `json:"tokens"`
	}
	return json.Marshal(safeMaskResult{
		Text:     redactedMaskTextSummary(),
		Findings: len(r.Findings),
		Tokens:   redactedTokenSetSummary(),
	})
}

func redactedMaskTextSummary() string {
	return "llmguard.MaskResult.text"
}

// TokenSet stores opaque placeholder mappings owned by the caller.
type TokenSet struct {
	mappings []tokenMapping
}

type tokenMapping struct {
	token string
	value string
}

func (t *TokenSet) String() string {
	return redactedTokenSetSummary()
}

func (t *TokenSet) GoString() string {
	return redactedTokenSetSummary()
}

// MarshalJSON returns a safe JSON representation without mappings.
func (t *TokenSet) MarshalJSON() ([]byte, error) {
	return json.Marshal(redactedTokenSetSummary())
}

func redactedTokenSetSummary() string {
	return "llmguard.TokenSet"
}

// Mask runs Detect and Resolve, then replaces resolved spans with collision-safe
// placeholders. Findings refer to byte offsets in the original input text.
func (g *Guard) Mask(ctx context.Context, text string) (MaskResult, error) {
	if err := ctx.Err(); err != nil {
		return MaskResult{}, err
	}

	started := time.Now()
	inputBytes := len(text)
	var (
		outcome  Outcome
		result   MaskResult
		err      error
		resolved []Finding
	)
	defer func() {
		event := Event{
			Operation:    OperationMask,
			Outcome:      outcome,
			InputBytes:   inputBytes,
			OutputBytes:  len(result.Text),
			Duration:     time.Since(started),
			FindingCount: len(result.Findings),
		}
		switch outcome {
		case OutcomeBlocked:
			event.OutputBytes = 0
			event.EntityCounts = buildEntityCounts(resolved)
			event.ActionCounts = buildActionCounts(resolved, g.actionForEntity)
		case OutcomeSuccess:
			event.EntityCounts = buildEntityCounts(result.Findings)
			event.ActionCounts = buildActionCounts(result.Findings, g.actionForEntity)
		}
		g.publishSafeEvent(event)
		g.publishUnsafeEvent(UnsafeDevelopmentEvent{
			Operation: OperationMask,
			Outcome:   outcome,
			Input:     text,
			Output:    result.Text,
			Findings:  result.Findings,
		})
	}()

	findings, err := g.Detect(ctx, text)
	if err != nil {
		outcome = OutcomeError
		return MaskResult{}, err
	}

	resolved, err = Resolve(text, findings)
	if err != nil {
		outcome = OutcomeError
		return MaskResult{}, err
	}

	if len(resolved) == 0 {
		outcome = OutcomeSuccess
		result = MaskResult{
			Text:     text,
			Findings: nil,
			Tokens:   newTokenSet(nil),
		}
		return result, nil
	}

	for _, finding := range resolved {
		if g.actionForEntity(finding.Entity) == ActionBlock {
			outcome = OutcomeBlocked
			return MaskResult{}, newBlockError()
		}
	}

	maskFindings := make([]Finding, 0, len(resolved))
	for _, finding := range resolved {
		if g.actionForEntity(finding.Entity) == ActionMask {
			maskFindings = append(maskFindings, finding)
		}
	}

	if len(maskFindings) == 0 {
		outcome = OutcomeSuccess
		result = MaskResult{
			Text:     text,
			Findings: resolved,
			Tokens:   newTokenSet(nil),
		}
		return result, nil
	}

	tokens, replacements, err := g.buildMaskReplacements(ctx, text, maskFindings)
	if err != nil {
		outcome = OutcomeError
		return MaskResult{}, err
	}

	masked := applyReplacementsRightToLeft(text, replacements)
	outcome = OutcomeSuccess
	result = MaskResult{
		Text:     masked,
		Findings: resolved,
		Tokens:   tokens,
	}
	return result, nil
}

// Restore replaces exact known placeholders in response using the provided
// TokenSet. Unknown, mutated, or cross-set tokens remain unchanged.
func (g *Guard) Restore(ctx context.Context, response string, tokens *TokenSet) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	started := time.Now()
	inputBytes := len(response)
	var (
		outcome  Outcome
		restored string
		err      error
		misses   int
	)
	defer func() {
		event := Event{
			Operation:     OperationRestore,
			Outcome:       outcome,
			InputBytes:    inputBytes,
			OutputBytes:   len(restored),
			Duration:      time.Since(started),
			RestoreMisses: misses,
		}
		g.publishSafeEvent(event)
		g.publishUnsafeEvent(UnsafeDevelopmentEvent{
			Operation: OperationRestore,
			Outcome:   outcome,
			Input:     response,
			Output:    restored,
		})
	}()

	if tokens == nil {
		outcome = OutcomeError
		err = newInvalidTokenSetError("nil token set")
		return "", err
	}
	if err = validateInputText(response); err != nil {
		outcome = OutcomeError
		return "", err
	}
	misses = countRestoreMisses(response, tokens)
	if len(tokens.mappings) == 0 {
		if misses > 0 {
			outcome = OutcomeRestoreMiss
		} else {
			outcome = OutcomeSuccess
		}
		restored = response
		return restored, nil
	}

	pairs := make([]string, 0, len(tokens.mappings)*2)
	for _, mapping := range tokens.mappings {
		pairs = append(pairs, mapping.token, mapping.value)
	}
	restored = strings.NewReplacer(pairs...).Replace(response)
	if misses > 0 {
		outcome = OutcomeRestoreMiss
	} else {
		outcome = OutcomeSuccess
	}
	return restored, nil
}

type maskReplacement struct {
	start int
	end   int
	token string
}

func (g *Guard) buildMaskReplacements(ctx context.Context, text string, findings []Finding) (*TokenSet, []maskReplacement, error) {
	for attempt := 0; attempt < maxNamespaceAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return nil, nil, err
		}

		namespace, err := g.readNamespace(ctx)
		if err != nil {
			return nil, nil, err
		}

		replacements := make([]maskReplacement, 0, len(findings))
		mappings := make([]tokenMapping, 0, len(findings))
		collision := false

		for i, finding := range findings {
			token := formatToken(namespace, i+1)
			if strings.Contains(text, token) {
				collision = true
				break
			}
			value := text[finding.Start:finding.End]
			replacements = append(replacements, maskReplacement{
				start: finding.Start,
				end:   finding.End,
				token: token,
			})
			mappings = append(mappings, tokenMapping{
				token: token,
				value: value,
			})
		}
		if collision {
			continue
		}

		return newTokenSet(mappings), replacements, nil
	}

	return nil, nil, newNamespaceCollisionError()
}

func newTokenSet(mappings []tokenMapping) *TokenSet {
	if len(mappings) == 0 {
		return &TokenSet{}
	}
	copied := make([]tokenMapping, len(mappings))
	copy(copied, mappings)
	return &TokenSet{mappings: copied}
}

func formatToken(namespace string, counter int) string {
	return fmt.Sprintf("%s%s_%04d%s", tokenPrefix, namespace, counter, tokenSuffix)
}

func applyReplacementsRightToLeft(text string, replacements []maskReplacement) string {
	if len(replacements) == 0 {
		return text
	}

	sorted := append([]maskReplacement(nil), replacements...)
	sortReplacementsDesc(sorted)

	segments := make([]string, 0, len(sorted)*2+1)
	end := len(text)
	for _, replacement := range sorted {
		segments = append(segments, text[replacement.end:end])
		segments = append(segments, replacement.token)
		end = replacement.start
	}
	segments = append(segments, text[:end])

	for i, j := 0, len(segments)-1; i < j; i, j = i+1, j-1 {
		segments[i], segments[j] = segments[j], segments[i]
	}
	return strings.Join(segments, "")
}

func sortReplacementsDesc(replacements []maskReplacement) {
	for i := 0; i < len(replacements); i++ {
		for j := i + 1; j < len(replacements); j++ {
			if replacements[j].start > replacements[i].start {
				replacements[i], replacements[j] = replacements[j], replacements[i]
			}
		}
	}
}

func (g *Guard) readNamespace(ctx context.Context) (string, error) {
	buf := make([]byte, namespaceEntropyBytes)

	g.randomMu.Lock()
	_, err := io.ReadFull(g.randomSource, buf)
	g.randomMu.Unlock()
	if err != nil {
		if ctx.Err() != nil {
			return "", ctx.Err()
		}
		return "", newNamespaceSourceError(err)
	}

	return hex.EncodeToString(buf), nil
}
