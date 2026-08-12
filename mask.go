package llmguard

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"strings"
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

	findings, err := g.Detect(ctx, text)
	if err != nil {
		return MaskResult{}, err
	}

	resolved, err := Resolve(text, findings)
	if err != nil {
		return MaskResult{}, err
	}

	if len(resolved) == 0 {
		return MaskResult{
			Text:     text,
			Findings: nil,
			Tokens:   newTokenSet(nil),
		}, nil
	}

	tokens, replacements, err := g.buildMaskReplacements(ctx, text, resolved)
	if err != nil {
		return MaskResult{}, err
	}

	masked := applyReplacementsRightToLeft(text, replacements)
	return MaskResult{
		Text:     masked,
		Findings: resolved,
		Tokens:   tokens,
	}, nil
}

// Restore replaces exact known placeholders in response using the provided
// TokenSet. Unknown, mutated, or cross-set tokens remain unchanged.
func (g *Guard) Restore(ctx context.Context, response string, tokens *TokenSet) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if tokens == nil {
		return "", newInvalidTokenSetError("nil token set")
	}
	if err := validateInputText(response); err != nil {
		return "", err
	}
	if len(tokens.mappings) == 0 {
		return response, nil
	}

	pairs := make([]string, 0, len(tokens.mappings)*2)
	for _, mapping := range tokens.mappings {
		pairs = append(pairs, mapping.token, mapping.value)
	}
	return strings.NewReplacer(pairs...).Replace(response), nil
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
