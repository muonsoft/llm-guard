package llmguard_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	muerrors "github.com/muonsoft/errors"
	"github.com/muonsoft/llm-guard"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func syntheticJWT(headerJSON, payloadJSON, signature string) string {
	encode := func(s string) string {
		return base64.RawURLEncoding.EncodeToString([]byte(s))
	}
	return encode(headerJSON) + "." + encode(payloadJSON) + "." + encode(signature)
}

func TestJWTDetector_WhenStructurallyValid_ExpectExactSpan(t *testing.T) {
	t.Parallel()

	token := syntheticJWT(`{"alg":"HS256","typ":"JWT"}`, `{"sub":"synthetic"}`, "sig")
	text := "prefix " + token + " suffix"
	detector := llmguard.NewJWTDetector()

	findings, err := detector.Detect(context.Background(), text)
	require.NoError(t, err)
	require.Len(t, findings, 1)
	assert.Equal(t, llmguard.EntitySecretJWT, findings[0].Entity)
	assert.Equal(t, 7, findings[0].Start)
	assert.Equal(t, 7+len(token), findings[0].End)
}

func TestJWTDetector_WhenAlgNone_ExpectNoMatch(t *testing.T) {
	t.Parallel()

	token := syntheticJWT(`{"alg":"none","typ":"JWT"}`, `{"sub":"x"}`, "sig")
	findings, err := llmguard.NewJWTDetector().Detect(context.Background(), token)
	require.NoError(t, err)
	assert.Empty(t, findings)
}

func TestJWTDetector_WhenMalformedSegments_ExpectNoMatch(t *testing.T) {
	t.Parallel()
	detector := llmguard.NewJWTDetector()

	cases := []struct {
		name string
		text string
	}{
		{name: "two_segments", text: "only.two"},
		{name: "four_segments", text: "a.b.c.d"},
		{name: "empty_segments", text: ".."},
		{name: "invalid_header_json", text: syntheticJWT(`not-json`, `{"sub":"x"}`, "sig")},
		{name: "missing_alg", text: syntheticJWT(`{"typ":"JWT"}`, `{"sub":"x"}`, "sig")},
	}
	for _, tc := range cases {
		findings, err := detector.Detect(context.Background(), tc.text)
		require.NoError(t, err, tc.name)
		assert.Empty(t, findings, tc.name)
	}
}

func TestJWTDetector_WhenInvalidSegmentLength_ExpectNoMatch(t *testing.T) {
	t.Parallel()

	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256"}`))
	sig := base64.RawURLEncoding.EncodeToString([]byte("sig"))
	token := header + ".Y." + sig
	findings, err := llmguard.NewJWTDetector().Detect(context.Background(), token)
	require.NoError(t, err)
	assert.Empty(t, findings)
}

func TestJWTDetector_WhenEmbeddedInTokenChars_ExpectBoundaryRespected(t *testing.T) {
	t.Parallel()

	token := syntheticJWT(`{"alg":"HS256"}`, `{"sub":"x"}`, "sig")
	embedded := "x" + token + "y"
	findings, err := llmguard.NewJWTDetector().Detect(context.Background(), embedded)
	require.NoError(t, err)
	assert.Empty(t, findings)
}

func TestJWTDetector_WhenContextCancelled_ExpectContextError(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := llmguard.NewJWTDetector().Detect(ctx, "any")
	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestJWTDetector_WhenConcurrent_ExpectDeterministic(t *testing.T) {
	t.Parallel()

	token := syntheticJWT(`{"alg":"RS256"}`, `{"sub":"x"}`, "sigbody")
	text := "token=" + token
	detector := llmguard.NewJWTDetector()

	const workers = 8
	results := make([][]llmguard.Finding, workers)
	errs := make([]error, workers)
	done := make(chan struct{})
	for i := 0; i < workers; i++ {
		go func(idx int) {
			results[idx], errs[idx] = detector.Detect(context.Background(), text)
			done <- struct{}{}
		}(i)
	}
	for i := 0; i < workers; i++ {
		<-done
	}
	for i := 0; i < workers; i++ {
		require.NoError(t, errs[i])
		assert.Equal(t, results[0], results[i])
	}
}

func TestPEMDetector_WhenSupportedLabel_ExpectFullBlockSpan(t *testing.T) {
	t.Parallel()

	block := "-----BEGIN RSA PRIVATE KEY-----\nAQIDBA==\n-----END RSA PRIVATE KEY-----"
	text := "stored:\n" + block + "\nend"
	findings, err := llmguard.NewPEMPrivateKeyDetector().Detect(context.Background(), text)
	require.NoError(t, err)
	require.Len(t, findings, 1)
	assert.Equal(t, llmguard.EntitySecretPrivateKey, findings[0].Entity)
	assert.Equal(t, strings.Index(text, "-----BEGIN"), findings[0].Start)
	assert.Equal(t, strings.Index(text, "-----BEGIN")+len(block), findings[0].End)
}

func TestPEMDetector_WhenPublicKey_ExpectNoMatch(t *testing.T) {
	t.Parallel()

	block := "-----BEGIN PUBLIC KEY-----\nAQIDBA==\n-----END PUBLIC KEY-----"
	findings, err := llmguard.NewPEMPrivateKeyDetector().Detect(context.Background(), block)
	require.NoError(t, err)
	assert.Empty(t, findings)
}

func TestPEMDetector_WhenMismatchedFooter_ExpectNoMatch(t *testing.T) {
	t.Parallel()

	block := "-----BEGIN RSA PRIVATE KEY-----\nAQIDBA==\n-----END EC PRIVATE KEY-----"
	findings, err := llmguard.NewPEMPrivateKeyDetector().Detect(context.Background(), block)
	require.NoError(t, err)
	assert.Empty(t, findings)
}

func TestPEMDetector_WhenMalformedBody_ExpectNoMatch(t *testing.T) {
	t.Parallel()

	block := "-----BEGIN PRIVATE KEY-----\n!!!\n-----END PRIVATE KEY-----"
	findings, err := llmguard.NewPEMPrivateKeyDetector().Detect(context.Background(), block)
	require.NoError(t, err)
	assert.Empty(t, findings)
}

func TestPEMDetector_WhenConcatenatedBlock_ExpectNoMatch(t *testing.T) {
	t.Parallel()

	block := "-----BEGIN PRIVATE KEY-----AQIDBA==-----END PRIVATE KEY-----"
	findings, err := llmguard.NewPEMPrivateKeyDetector().Detect(context.Background(), block)
	require.NoError(t, err)
	assert.Empty(t, findings)
}

func TestPEMDetector_WhenCRLFLineEndings_ExpectMatch(t *testing.T) {
	t.Parallel()

	block := "-----BEGIN PRIVATE KEY-----\r\nAQIDBA==\r\n-----END PRIVATE KEY-----"
	findings, err := llmguard.NewPEMPrivateKeyDetector().Detect(context.Background(), block)
	require.NoError(t, err)
	require.Len(t, findings, 1)
}

func TestPEMDetector_WhenAllSupportedLabels_ExpectMatch(t *testing.T) {
	t.Parallel()

	labels := []string{
		"PRIVATE KEY",
		"RSA PRIVATE KEY",
		"EC PRIVATE KEY",
		"OPENSSH PRIVATE KEY",
		"PGP PRIVATE KEY BLOCK",
	}
	for _, label := range labels {
		block := fmt.Sprintf("-----BEGIN %s-----\nAQIDBA==\n-----END %s-----", label, label)
		findings, err := llmguard.NewPEMPrivateKeyDetector().Detect(context.Background(), block)
		require.NoError(t, err, label)
		require.Len(t, findings, 1, label)
	}
}

func TestPEMDetector_WhenContextCancelled_ExpectContextError(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := llmguard.NewPEMPrivateKeyDetector().Detect(ctx, "x")
	require.Error(t, err)
}

func TestPEMDetector_WhenConcurrent_ExpectDeterministic(t *testing.T) {
	t.Parallel()

	block := "-----BEGIN PRIVATE KEY-----\nAQIDBA==\n-----END PRIVATE KEY-----"
	detector := llmguard.NewPEMPrivateKeyDetector()
	const workers = 8
	results := make([][]llmguard.Finding, workers)
	errs := make([]error, workers)
	done := make(chan struct{})
	for i := 0; i < workers; i++ {
		go func(idx int) {
			results[idx], errs[idx] = detector.Detect(context.Background(), block)
			done <- struct{}{}
		}(i)
	}
	for i := 0; i < workers; i++ {
		<-done
	}
	for i := 0; i < workers; i++ {
		require.NoError(t, errs[i])
		assert.Equal(t, results[0], results[i])
	}
}

func syntheticGitHubClassic() string { return "ghp_" + strings.Repeat("A", 36) }
func syntheticGitHubFine() string {
	return "github_pat_" + strings.Repeat("A", 22) + "_" + strings.Repeat("B", 59)
}
func syntheticGitLabPAT() string    { return "glpat-" + strings.Repeat("A", 20) }
func syntheticAWSKey() string       { return "AKIA" + strings.Repeat("A", 16) }
func syntheticOpenAISk() string     { return "sk-" + strings.Repeat("A", 20) }
func syntheticOpenAIProjSk() string { return "sk-proj-" + strings.Repeat("A", 20) }

func TestAPIKeyDetector_WhenSupportedShapes_ExpectExactSpans(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		token string
	}{
		{"github_classic", syntheticGitHubClassic()},
		{"github_fine", syntheticGitHubFine()},
		{"gitlab", syntheticGitLabPAT()},
		{"aws", syntheticAWSKey()},
		{"aws_asia", "ASIA" + strings.Repeat("B", 16)},
		{"openai_sk", syntheticOpenAISk()},
		{"openai_proj", syntheticOpenAIProjSk()},
	}
	detector := llmguard.NewAPIKeyDetector()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			text := "key=" + tc.token + "!"
			findings, err := detector.Detect(context.Background(), text)
			require.NoError(t, err)
			require.Len(t, findings, 1)
			assert.Equal(t, 4, findings[0].Start)
			assert.Equal(t, 4+len(tc.token), findings[0].End)
		})
	}
}

func TestAPIKeyDetector_WhenMixedShapes_ExpectStableTextualOrder(t *testing.T) {
	t.Parallel()

	aws := syntheticAWSKey()
	ghp := syntheticGitHubClassic()
	text := aws + " " + ghp
	findings, err := llmguard.NewAPIKeyDetector().Detect(context.Background(), text)
	require.NoError(t, err)
	require.Len(t, findings, 2)
	assert.Equal(t, 0, findings[0].Start)
	assert.True(t, findings[0].Start < findings[1].Start)

	reversed := ghp + " " + aws
	reversedFindings, err := llmguard.NewAPIKeyDetector().Detect(context.Background(), reversed)
	require.NoError(t, err)
	require.Len(t, reversedFindings, 2)
	assert.Equal(t, 0, reversedFindings[0].Start)
	assert.True(t, reversedFindings[0].Start < reversedFindings[1].Start)
}

func TestAPIKeyDetector_WhenTruncatedOrInvalidAlphabet_ExpectNoMatch(t *testing.T) {
	t.Parallel()

	detector := llmguard.NewAPIKeyDetector()
	cases := []struct {
		name string
		text string
	}{
		{name: "ghp_truncated", text: "ghp_short"},
		{name: "github_pat_prefix_only", text: "github_pat_"},
		{name: "github_pat_truncated", text: "github_pat_" + strings.Repeat("A", 22) + "_" + strings.Repeat("B", 10)},
		{name: "gitlab_truncated", text: "glpat-tooshort"},
		{name: "aws_truncated", text: "AKIA123"},
		{name: "openai_sk_truncated", text: "sk-tooshort"},
		{name: "openai_proj_truncated", text: "sk-proj-short"},
		{name: "ghp_invalid_alphabet", text: "ghp_" + strings.Repeat("!", 36)},
	}
	for _, tc := range cases {
		findings, err := detector.Detect(context.Background(), tc.text)
		require.NoError(t, err, tc.name)
		assert.Empty(t, findings, tc.name)
	}
}

func TestAPIKeyDetector_WhenEmbedded_ExpectNoMatch(t *testing.T) {
	t.Parallel()

	token := syntheticGitHubClassic()
	findings, err := llmguard.NewAPIKeyDetector().Detect(context.Background(), "x"+token+"y")
	require.NoError(t, err)
	assert.Empty(t, findings)
}

func TestAPIKeyDetector_WhenContextCancelled_ExpectContextError(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := llmguard.NewAPIKeyDetector().Detect(ctx, "x")
	require.Error(t, err)
}

func TestAPIKeyDetector_WhenConcurrent_ExpectDeterministic(t *testing.T) {
	t.Parallel()

	text := "token " + syntheticGitLabPAT()
	detector := llmguard.NewAPIKeyDetector()
	const workers = 8
	results := make([][]llmguard.Finding, workers)
	errs := make([]error, workers)
	done := make(chan struct{})
	for i := 0; i < workers; i++ {
		go func(idx int) {
			results[idx], errs[idx] = detector.Detect(context.Background(), text)
			done <- struct{}{}
		}(i)
	}
	for i := 0; i < workers; i++ {
		<-done
	}
	for i := 0; i < workers; i++ {
		require.NoError(t, errs[i])
		assert.Equal(t, results[0], results[i])
	}
}

func TestDSNDetector_WhenCredentialBearing_ExpectFullSpan(t *testing.T) {
	t.Parallel()

	dsn := "postgres://dbuser:dbpass@db.example.com:5432/app?sslmode=disable"
	text := "connect " + dsn + " now"
	findings, err := llmguard.NewDSNDetector().Detect(context.Background(), text)
	require.NoError(t, err)
	require.Len(t, findings, 1)
	assert.Equal(t, 8, findings[0].Start)
	assert.Equal(t, 8+len(dsn), findings[0].End)
}

func TestDSNDetector_WhenPercentEncodingAndIPv6_ExpectMatch(t *testing.T) {
	t.Parallel()

	dsn := "postgresql://user%40corp:p%40ss@[::1]:5432/db"
	findings, err := llmguard.NewDSNDetector().Detect(context.Background(), dsn)
	require.NoError(t, err)
	require.Len(t, findings, 1)
	assert.Equal(t, 0, findings[0].Start)
	assert.Equal(t, len(dsn), findings[0].End)
}

func TestDSNDetector_WhenIPv6WithoutPath_ExpectClosingBracketPreserved(t *testing.T) {
	t.Parallel()

	dsn := "postgres://u:p@[::1]"
	findings, err := llmguard.NewDSNDetector().Detect(context.Background(), dsn)
	require.NoError(t, err)
	require.Len(t, findings, 1)
	assert.Equal(t, 0, findings[0].Start)
	assert.Equal(t, len(dsn), findings[0].End)
}

func TestDSNDetector_WhenHTTPSOrPasswordless_ExpectNoMatch(t *testing.T) {
	t.Parallel()

	detector := llmguard.NewDSNDetector()
	cases := []struct {
		name string
		text string
	}{
		{name: "https_scheme", text: "https://user:pass@example.com"},
		{name: "passwordless", text: "postgres://user@localhost/db"},
		{name: "query_only_password", text: "postgres://localhost/db?password=secret"},
		{name: "invalid_scheme", text: "postgres:invalid"},
	}
	for _, tc := range cases {
		findings, err := detector.Detect(context.Background(), tc.text)
		require.NoError(t, err, tc.name)
		assert.Empty(t, findings, tc.name)
	}
}

func TestDSNDetector_WhenTrailingPunctuation_ExpectTrimmedSpan(t *testing.T) {
	t.Parallel()

	dsn := "mysql://u:p@host/db"
	text := "(" + dsn + ")."
	findings, err := llmguard.NewDSNDetector().Detect(context.Background(), text)
	require.NoError(t, err)
	require.Len(t, findings, 1)
	assert.Equal(t, 1, findings[0].Start)
	assert.Equal(t, 1+len(dsn), findings[0].End)
}

func TestDSNDetector_WhenContextCancelled_ExpectContextError(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := llmguard.NewDSNDetector().Detect(ctx, "x")
	require.Error(t, err)
}

func TestDSNDetector_WhenConcurrent_ExpectDeterministic(t *testing.T) {
	t.Parallel()

	dsn := "redis://u:p@127.0.0.1:6379/0"
	detector := llmguard.NewDSNDetector()
	const workers = 8
	results := make([][]llmguard.Finding, workers)
	errs := make([]error, workers)
	done := make(chan struct{})
	for i := 0; i < workers; i++ {
		go func(idx int) {
			results[idx], errs[idx] = detector.Detect(context.Background(), dsn)
			done <- struct{}{}
		}(i)
	}
	for i := 0; i < workers; i++ {
		<-done
	}
	for i := 0; i < workers; i++ {
		require.NoError(t, errs[i])
		assert.Equal(t, results[0], results[i])
	}
}

func TestPolicy_WhenValidActions_ExpectConstants(t *testing.T) {
	t.Parallel()

	assert.Equal(t, llmguard.Action("allow"), llmguard.ActionAllow)
	assert.Equal(t, llmguard.Action("mask"), llmguard.ActionMask)
	assert.Equal(t, llmguard.Action("block"), llmguard.ActionBlock)
}

func TestPolicy_WhenInvalidConfiguration_ExpectError(t *testing.T) {
	t.Parallel()

	_, err := llmguard.New(llmguard.WithSecretAction(llmguard.Action("deny")))
	require.Error(t, err)
	assert.True(t, muerrors.Is(err, llmguard.ErrInvalidConfig))

	_, err = llmguard.New(llmguard.WithEntityAction("", llmguard.ActionMask))
	require.Error(t, err)

	_, err = llmguard.New(
		llmguard.WithEntityAction(llmguard.EntityEmail, llmguard.ActionAllow),
		llmguard.WithEntityAction(llmguard.EntityEmail, llmguard.ActionMask),
	)
	require.Error(t, err)

	_, err = llmguard.New(nil)
	require.Error(t, err)

	_, err = llmguard.New(
		llmguard.WithSecretAction(llmguard.ActionMask),
		llmguard.WithSecretAction(llmguard.ActionMask),
	)
	require.Error(t, err)
	assert.True(t, muerrors.Is(err, llmguard.ErrInvalidConfig))

	_, err = llmguard.New(
		llmguard.WithSecretAction(llmguard.ActionBlock),
		llmguard.WithSecretAction(llmguard.ActionMask),
	)
	require.Error(t, err)
	assert.True(t, muerrors.Is(err, llmguard.ErrInvalidConfig))
}

func TestPolicy_WhenEntityOverride_ExpectPrecedence(t *testing.T) {
	t.Parallel()

	guard, err := llmguard.New(
		llmguard.WithDetector(llmguard.NewJWTDetector()),
		llmguard.WithSecretAction(llmguard.ActionMask),
		llmguard.WithEntityAction(llmguard.EntitySecretJWT, llmguard.ActionBlock),
	)
	require.NoError(t, err)

	token := syntheticJWT(`{"alg":"HS256"}`, `{"sub":"x"}`, "sig")
	_, err = guard.Mask(context.Background(), token)
	require.Error(t, err)
	assert.True(t, muerrors.Is(err, llmguard.ErrBlocked))
}

func TestResolve_WhenDSNOverlapsURL_ExpectDSNWins(t *testing.T) {
	t.Parallel()

	dsn := "postgres://u:p@host/db"
	text := "x " + dsn
	dsnFinding := llmguard.Finding{
		Entity: llmguard.EntityConnectionString, Start: 2, End: 2 + len(dsn), Confidence: 0.9, Detector: "secret_dsn",
	}
	urlFinding := llmguard.Finding{
		Entity: llmguard.EntityURL, Start: 2, End: 2 + len(dsn), Confidence: 0.9, Detector: "url",
	}
	resolved, err := llmguard.Resolve(text, []llmguard.Finding{urlFinding, dsnFinding})
	require.NoError(t, err)
	require.Len(t, resolved, 1)
	assert.Equal(t, llmguard.EntityConnectionString, resolved[0].Entity)
}

func TestResolve_WhenSecretOverlapsCustom_ExpectSecretWins(t *testing.T) {
	t.Parallel()

	token := syntheticGitHubClassic()
	text := token
	secret := llmguard.Finding{
		Entity: llmguard.EntitySecretAPIKey, Start: 0, End: len(token), Confidence: 0.9, Detector: "secret_api_key",
	}
	custom := llmguard.Finding{
		Entity: llmguard.EntityType("CUSTOM"), Start: 4, End: 10, Confidence: 1, Detector: "custom",
	}
	resolved, err := llmguard.Resolve(text, []llmguard.Finding{custom, secret})
	require.NoError(t, err)
	require.Len(t, resolved, 1)
	assert.Equal(t, llmguard.EntitySecretAPIKey, resolved[0].Entity)
}

func TestMask_WhenDefaultSecret_ExpectBlock(t *testing.T) {
	t.Parallel()

	guard, err := llmguard.New(llmguard.WithDetector(llmguard.NewAPIKeyDetector()))
	require.NoError(t, err)

	text := "key " + syntheticGitHubClassic()
	result, err := guard.Mask(context.Background(), text)
	require.Error(t, err)
	assert.True(t, muerrors.Is(err, llmguard.ErrBlocked))
	assert.Equal(t, llmguard.MaskResult{}, result)
}

func TestMask_WhenExplicitSecretMask_ExpectReversible(t *testing.T) {
	t.Parallel()

	guard, err := llmguard.New(
		llmguard.WithDetector(llmguard.NewAPIKeyDetector()),
		llmguard.WithSecretAction(llmguard.ActionMask),
	)
	require.NoError(t, err)

	text := "key " + syntheticGitHubClassic()
	result, err := guard.Mask(context.Background(), text)
	require.NoError(t, err)
	assert.NotContains(t, result.Text, "ghp_")
	require.Len(t, result.Findings, 1)

	restored, err := guard.Restore(context.Background(), result.Text, result.Tokens)
	require.NoError(t, err)
	assert.Equal(t, text, restored)
}

func TestMask_WhenMixedAllowMaskBlock_ExpectBlockWithoutPartial(t *testing.T) {
	t.Parallel()

	guard, err := llmguard.New(
		llmguard.WithDetector(llmguard.NewEmailDetector()),
		llmguard.WithDetector(llmguard.NewAPIKeyDetector()),
		llmguard.WithEntityAction(llmguard.EntityEmail, llmguard.ActionAllow),
	)
	require.NoError(t, err)

	text := "mail a@b.co token " + syntheticGitLabPAT()
	result, err := guard.Mask(context.Background(), text)
	require.Error(t, err)
	assert.True(t, muerrors.Is(err, llmguard.ErrBlocked))
	assert.Equal(t, llmguard.MaskResult{}, result)
	assert.NotContains(t, result.Text, "{{LLMG_")
}

func TestMask_WhenBlock_ExpectNoEntropyRead(t *testing.T) {
	t.Parallel()

	guard, err := llmguard.New(
		llmguard.WithDetector(llmguard.NewAPIKeyDetector()),
		llmguard.WithRandomSource(blockingEntropyReader{t: t}),
	)
	require.NoError(t, err)

	text := "key " + syntheticGitHubClassic()
	result, err := guard.Mask(context.Background(), text)
	require.Error(t, err)
	assert.True(t, muerrors.Is(err, llmguard.ErrBlocked))
	assert.Equal(t, llmguard.MaskResult{}, result)
}

type blockingEntropyReader struct {
	t *testing.T
}

func (r blockingEntropyReader) Read([]byte) (int, error) {
	r.t.Helper()
	r.t.Fatal("entropy reader accessed during block")
	return 0, nil
}

func TestMask_WhenAllAllow_ExpectUnchangedBytes(t *testing.T) {
	t.Parallel()

	guard, err := llmguard.New(
		llmguard.WithDetector(llmguard.NewEmailDetector()),
		llmguard.WithEntityAction(llmguard.EntityEmail, llmguard.ActionAllow),
	)
	require.NoError(t, err)

	text := "mail a@b.co"
	result, err := guard.Mask(context.Background(), text)
	require.NoError(t, err)
	assert.Equal(t, text, result.Text)
	require.Len(t, result.Findings, 1)
}

func TestMask_WhenMixedAllowAndMask_ExpectPartialReplacement(t *testing.T) {
	t.Parallel()

	guard, err := llmguard.New(
		llmguard.WithDetector(llmguard.NewEmailDetector()),
		llmguard.WithDetector(llmguard.NewAPIKeyDetector()),
		llmguard.WithSecretAction(llmguard.ActionMask),
		llmguard.WithEntityAction(llmguard.EntityEmail, llmguard.ActionAllow),
	)
	require.NoError(t, err)

	text := "mail a@b.co key " + syntheticAWSKey()
	result, err := guard.Mask(context.Background(), text)
	require.NoError(t, err)
	assert.Contains(t, result.Text, "a@b.co")
	assert.NotContains(t, result.Text, "AKIA")
	require.Len(t, result.Findings, 2)
}

func TestBlock_WhenFormatted_ExpectNoSecretLeakage(t *testing.T) {
	t.Parallel()

	guard, err := llmguard.New(llmguard.WithDetector(llmguard.NewDSNDetector()))
	require.NoError(t, err)

	secretDSN := "postgres://leakuser:leakpass@host/db"
	_, err = guard.Mask(context.Background(), secretDSN)
	require.Error(t, err)

	representations := []string{
		fmt.Sprintf("%v", err),
		fmt.Sprintf("%+v", err),
		fmt.Sprintf("%#v", err),
	}
	for _, repr := range representations {
		assert.NotContains(t, repr, "leakuser")
		assert.NotContains(t, repr, "leakpass")
		assert.NotContains(t, repr, secretDSN)
	}

	blockErr, ok := muerrors.As[*llmguard.BlockError](err)
	require.True(t, ok)
	assert.NotNil(t, blockErr)
	assert.True(t, muerrors.Is(err, llmguard.ErrBlocked))
}

func TestBlock_WhenJSONAdjacent_ExpectNoSensitiveFields(t *testing.T) {
	t.Parallel()

	guard, err := llmguard.New(llmguard.WithDetector(llmguard.NewDSNDetector()))
	require.NoError(t, err)

	secretDSN := "postgres://leakuser:leakpass@host/db"
	_, err = guard.Mask(context.Background(), secretDSN)
	require.Error(t, err)

	envelope := map[string]string{
		"status":  "blocked",
		"message": err.Error(),
	}
	raw, marshalErr := json.Marshal(envelope)
	require.NoError(t, marshalErr)
	assert.NotContains(t, string(raw), "leakuser")
	assert.NotContains(t, string(raw), "leakpass")
	assert.NotContains(t, string(raw), secretDSN)
}

func TestMask_WhenPIIRegression_ExpectDefaultMask(t *testing.T) {
	t.Parallel()

	guard, err := llmguard.New(llmguard.WithDetector(llmguard.NewEmailDetector()))
	require.NoError(t, err)

	text := "Contact a@b.co"
	result, err := guard.Mask(context.Background(), text)
	require.NoError(t, err)
	assert.NotEqual(t, text, result.Text)
}
