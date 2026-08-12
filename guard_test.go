package llmguard_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	muerrors "github.com/muonsoft/errors"
	"github.com/muonsoft/llm-guard"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type staticDetector struct {
	name     string
	findings []llmguard.Finding
	err      error
	delay    time.Duration
	onDetect func(context.Context)
	invoked  *atomic.Bool
}

func (d *staticDetector) Name() string { return d.name }

func (d *staticDetector) Detect(ctx context.Context, text string) ([]llmguard.Finding, error) {
	if d.invoked != nil {
		d.invoked.Store(true)
	}
	if d.onDetect != nil {
		d.onDetect(ctx)
	}
	if d.delay > 0 {
		select {
		case <-time.After(d.delay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	if d.err != nil {
		return nil, d.err
	}
	return d.findings, nil
}

func TestGuard_WhenCustomDetector_ExpectFindings(t *testing.T) {
	t.Parallel()

	detector := &staticDetector{
		name: "email",
		findings: []llmguard.Finding{{
			Entity:     llmguard.EntityEmail,
			Start:      7,
			End:        20,
			Confidence: 0.9,
		}},
	}

	guard, err := llmguard.New(llmguard.WithDetector(detector))
	require.NoError(t, err)

	findings, err := guard.Detect(context.Background(), "contact user@example.com today")
	require.NoError(t, err)
	require.Len(t, findings, 1)
	assert.Equal(t, llmguard.EntityEmail, findings[0].Entity)
	assert.Equal(t, 7, findings[0].Start)
	assert.Equal(t, 20, findings[0].End)
	assert.Equal(t, "email", findings[0].Detector)
}

func TestGuard_WhenNoDetectors_ExpectEmptyResult(t *testing.T) {
	t.Parallel()

	guard, err := llmguard.New()
	require.NoError(t, err)

	findings, err := guard.Detect(context.Background(), "plain text")
	require.NoError(t, err)
	assert.Nil(t, findings)
}

func TestGuard_WhenNilDetector_ExpectInvalidConfig(t *testing.T) {
	t.Parallel()

	_, err := llmguard.New(llmguard.WithDetector(nil))
	require.Error(t, err)
	assert.ErrorIs(t, err, llmguard.ErrInvalidConfig)
}

func TestGuard_WhenTypedNilDetector_ExpectInvalidConfig(t *testing.T) {
	t.Parallel()

	var detector *staticDetector
	_, err := llmguard.New(llmguard.WithDetector(detector))
	require.Error(t, err)
	assert.ErrorIs(t, err, llmguard.ErrInvalidConfig)
}

func TestGuard_WhenEmptyDetectorName_ExpectInvalidConfig(t *testing.T) {
	t.Parallel()

	detector := &staticDetector{name: ""}
	_, err := llmguard.New(llmguard.WithDetector(detector))
	require.Error(t, err)
	assert.ErrorIs(t, err, llmguard.ErrInvalidConfig)
}

func TestGuard_WhenDuplicateDetectorName_ExpectInvalidConfig(t *testing.T) {
	t.Parallel()

	first := &staticDetector{name: "dup"}
	second := &staticDetector{name: "dup"}
	_, err := llmguard.New(llmguard.WithDetector(first), llmguard.WithDetector(second))
	require.Error(t, err)
	assert.ErrorIs(t, err, llmguard.ErrInvalidConfig)
}

func TestGuard_WhenInvalidUTF8Input_ExpectInvalidText(t *testing.T) {
	t.Parallel()

	var invoked atomic.Bool
	guard, err := llmguard.New(llmguard.WithDetector(&staticDetector{name: "noop", invoked: &invoked}))
	require.NoError(t, err)

	_, err = guard.Detect(context.Background(), string([]byte{0xff}))
	require.Error(t, err)
	assert.ErrorIs(t, err, llmguard.ErrInvalidText)
	assert.False(t, invoked.Load(), "detector must not run for invalid UTF-8 input")
}

func TestGuard_WhenContextCanceledBeforeDetect_ExpectCanceled(t *testing.T) {
	t.Parallel()

	guard, err := llmguard.New(llmguard.WithDetector(&staticDetector{name: "slow", delay: time.Second}))
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	findings, err := guard.Detect(ctx, "text")
	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
	assert.Nil(t, findings)
}

func TestGuard_WhenContextCanceledDuringDetect_ExpectCallerError(t *testing.T) {
	t.Parallel()

	started := make(chan struct{})
	detector := &staticDetector{
		name: "blocking",
		onDetect: func(ctx context.Context) {
			close(started)
			<-ctx.Done()
		},
	}

	guard, err := llmguard.New(llmguard.WithDetector(detector))
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan struct{})
	var (
		findings  []llmguard.Finding
		detectErr error
	)
	go func() {
		findings, detectErr = guard.Detect(ctx, "text")
		close(done)
	}()

	<-started
	cancel()
	<-done

	require.Error(t, detectErr)
	assert.ErrorIs(t, detectErr, context.Canceled)
	assert.Nil(t, findings)
}

func TestGuard_WhenDetectorReturnsError_ExpectNilFindingsAndDetectorError(t *testing.T) {
	t.Parallel()

	root := errors.New("root cause")
	detector := &staticDetector{name: "fail", err: root}

	guard, err := llmguard.New(llmguard.WithDetector(detector))
	require.NoError(t, err)

	findings, err := guard.Detect(context.Background(), "secret-value-here")
	require.Error(t, err)
	assert.Nil(t, findings)
	assert.ErrorIs(t, err, llmguard.ErrDetector)
	assert.ErrorIs(t, err, root)

	detErr, ok := muerrors.As[*llmguard.DetectorError](err)
	require.True(t, ok)
	assert.Equal(t, "fail", detErr.Detector)
	assert.NotContains(t, detErr.Error(), "secret")
	assert.NotContains(t, detErr.Error(), "root cause")
}

func TestGuard_WhenMultipleDetectorErrors_ExpectLowestRegistrationIndex(t *testing.T) {
	t.Parallel()

	secondDone := make(chan struct{})
	first := &barrierDetector{
		name:    "first",
		err:     errors.New("first failure"),
		waitFor: secondDone,
	}
	second := &barrierDetector{
		name:   "second",
		err:    errors.New("second failure"),
		signal: secondDone,
	}

	guard, err := llmguard.New(llmguard.WithDetector(first), llmguard.WithDetector(second))
	require.NoError(t, err)

	_, err = guard.Detect(context.Background(), "text")
	require.Error(t, err)

	detErr, ok := muerrors.As[*llmguard.DetectorError](err)
	require.True(t, ok)
	assert.Equal(t, "first", detErr.Detector)
}

type barrierDetector struct {
	name    string
	err     error
	waitFor <-chan struct{}
	signal  chan<- struct{}
}

func (d *barrierDetector) Name() string { return d.name }

func (d *barrierDetector) Detect(ctx context.Context, text string) ([]llmguard.Finding, error) {
	if d.signal != nil {
		close(d.signal)
	}
	if d.waitFor != nil {
		<-d.waitFor
	}
	if d.err != nil {
		return nil, d.err
	}
	return nil, nil
}

func TestGuard_WhenDetectorsFinishInDifferentOrder_ExpectStableSort(t *testing.T) {
	t.Parallel()

	slow := &staticDetector{
		name:  "slow",
		delay: 30 * time.Millisecond,
		findings: []llmguard.Finding{{
			Entity: llmguard.EntityPhone, Start: 10, End: 14, Confidence: 0.5,
		}},
	}
	fast := &staticDetector{
		name:  "fast",
		delay: 0,
		findings: []llmguard.Finding{{
			Entity: llmguard.EntityEmail, Start: 0, End: 5, Confidence: 0.9,
		}},
	}

	guard, err := llmguard.New(llmguard.WithDetector(slow), llmguard.WithDetector(fast))
	require.NoError(t, err)

	findings, err := guard.Detect(context.Background(), "email phone1234")
	require.NoError(t, err)
	require.Len(t, findings, 2)
	assert.Equal(t, 0, findings[0].Start)
	assert.Equal(t, 10, findings[1].Start)
}

func TestGuard_WhenConcurrentDetect_ExpectNoDataRace(t *testing.T) {
	t.Parallel()

	detector := &staticDetector{
		name: "concurrent",
		findings: []llmguard.Finding{{
			Entity: llmguard.EntityEmail, Start: 0, End: 1, Confidence: 1,
		}},
	}
	guard, err := llmguard.New(llmguard.WithDetector(detector))
	require.NoError(t, err)

	const workers = 16
	type detectResult struct {
		findings []llmguard.Finding
		err      error
	}
	results := make(chan detectResult, workers)
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			findings, detectErr := guard.Detect(context.Background(), "x")
			results <- detectResult{findings: findings, err: detectErr}
		}()
	}
	wg.Wait()
	close(results)

	for result := range results {
		require.NoError(t, result.err)
		require.Len(t, result.findings, 1)
	}
}

func TestGuard_WhenSiblingFailureCancelsOthers_ExpectOriginalFailure(t *testing.T) {
	t.Parallel()

	var canceled atomic.Bool
	cooperativeStarted := make(chan struct{})
	substantiveReturned := make(chan struct{})

	cooperative := &staticDetector{
		name: "cooperative",
		onDetect: func(ctx context.Context) {
			close(cooperativeStarted)
			<-ctx.Done()
			canceled.Store(true)
		},
	}
	failing := &staticDetector{
		name: "failing",
		err:  errors.New("boom"),
		onDetect: func(context.Context) {
			<-cooperativeStarted
			close(substantiveReturned)
		},
	}

	guard, err := llmguard.New(llmguard.WithDetector(cooperative), llmguard.WithDetector(failing))
	require.NoError(t, err)

	_, err = guard.Detect(context.Background(), "text")
	require.Error(t, err)

	detErr, ok := muerrors.As[*llmguard.DetectorError](err)
	require.True(t, ok)
	assert.Equal(t, "failing", detErr.Detector)
	assert.True(t, canceled.Load())

	<-substantiveReturned
}

func TestGuard_WhenSensitiveDetectorError_ExpectRedactedPublicMessage(t *testing.T) {
	t.Parallel()

	sensitive := errors.New("matched secret-value-12345")
	detector := &staticDetector{name: "leaky", err: sensitive}

	guard, err := llmguard.New(llmguard.WithDetector(detector))
	require.NoError(t, err)

	_, err = guard.Detect(context.Background(), "secret-value-12345")
	require.Error(t, err)
	assert.ErrorIs(t, err, sensitive)
	assert.NotContains(t, err.Error(), "secret-value-12345")
	assert.NotContains(t, fmt.Sprint(err), "secret-value-12345")
	assert.NotContains(t, fmt.Sprintf("%+v", err), "secret-value-12345")
}

func TestGuard_WhenEmptyMetadata_ExpectRegisteredDetectorName(t *testing.T) {
	t.Parallel()

	detector := &staticDetector{
		name: "registered",
		findings: []llmguard.Finding{{
			Entity: llmguard.EntityEmail, Start: 0, End: 1, Confidence: 1,
		}},
	}
	guard, err := llmguard.New(llmguard.WithDetector(detector))
	require.NoError(t, err)

	findings, err := guard.Detect(context.Background(), "a")
	require.NoError(t, err)
	assert.Equal(t, "registered", findings[0].Detector)
}

func TestGuard_WhenConflictingMetadata_ExpectInvalidFinding(t *testing.T) {
	t.Parallel()

	detector := &staticDetector{
		name: "registered",
		findings: []llmguard.Finding{{
			Entity: llmguard.EntityEmail, Start: 0, End: 1, Confidence: 1, Detector: "other",
		}},
	}
	guard, err := llmguard.New(llmguard.WithDetector(detector))
	require.NoError(t, err)

	_, err = guard.Detect(context.Background(), "a")
	require.Error(t, err)
	assert.ErrorIs(t, err, llmguard.ErrInvalidFinding)
}

func TestGuard_WhenSortTieBreakers_ExpectDeterministicOrder(t *testing.T) {
	t.Parallel()

	detector := &staticDetector{
		name: "multi",
		findings: []llmguard.Finding{
			{Entity: llmguard.EntityEmail, Start: 0, End: 2, Confidence: 0.5},
			{Entity: llmguard.EntityPhone, Start: 0, End: 2, Confidence: 0.9},
			{Entity: llmguard.EntityEmail, Start: 0, End: 2, Confidence: 0.9},
		},
	}
	guard, err := llmguard.New(llmguard.WithDetector(detector))
	require.NoError(t, err)

	findings, err := guard.Detect(context.Background(), "ab")
	require.NoError(t, err)
	require.Len(t, findings, 3)
	assert.Equal(t, llmguard.EntityEmail, findings[0].Entity)
	assert.InDelta(t, 0.9, findings[0].Confidence, 0)
	assert.Equal(t, llmguard.EntityEmail, findings[1].Entity)
	assert.InDelta(t, 0.5, findings[1].Confidence, 0)
	assert.Equal(t, llmguard.EntityPhone, findings[2].Entity)
}

func TestGuard_WhenMultipleDetectorsSameSpan_ExpectRegistrationOrderTieBreak(t *testing.T) {
	t.Parallel()

	first := &staticDetector{
		name: "alpha",
		findings: []llmguard.Finding{{
			Entity: llmguard.EntityEmail, Start: 0, End: 1, Confidence: 0.5,
		}},
	}
	second := &staticDetector{
		name: "beta",
		findings: []llmguard.Finding{{
			Entity: llmguard.EntityEmail, Start: 0, End: 1, Confidence: 0.5,
		}},
	}

	guard, err := llmguard.New(llmguard.WithDetector(first), llmguard.WithDetector(second))
	require.NoError(t, err)

	findings, err := guard.Detect(context.Background(), "a")
	require.NoError(t, err)
	require.Len(t, findings, 2)
	assert.Equal(t, "alpha", findings[0].Detector)
	assert.Equal(t, "beta", findings[1].Detector)
}

func TestGuard_WhenDetectorNameHasControlChar_ExpectInvalidConfig(t *testing.T) {
	t.Parallel()

	detector := &staticDetector{name: "bad\nname"}
	_, err := llmguard.New(llmguard.WithDetector(detector))
	require.Error(t, err)
	assert.ErrorIs(t, err, llmguard.ErrInvalidConfig)
}

func TestGuard_WhenNilOption_ExpectInvalidConfig(t *testing.T) {
	t.Parallel()

	_, err := llmguard.New(nil)
	require.Error(t, err)
	assert.ErrorIs(t, err, llmguard.ErrInvalidConfig)
}

func TestGuard_WhenPartialFailure_ExpectNoPartialFindings(t *testing.T) {
	t.Parallel()

	ok := &staticDetector{
		name: "ok",
		findings: []llmguard.Finding{{
			Entity: llmguard.EntityEmail, Start: 0, End: 1, Confidence: 1,
		}},
	}
	bad := &staticDetector{name: "bad", err: errors.New("fail")}

	guard, err := llmguard.New(llmguard.WithDetector(ok), llmguard.WithDetector(bad))
	require.NoError(t, err)

	findings, err := guard.Detect(context.Background(), "a")
	require.Error(t, err)
	assert.Nil(t, findings)
}

func TestGuard_WhenUnicodeText_ExpectByteOffsetsPreserved(t *testing.T) {
	t.Parallel()

	text := "Привет мир"
	start := strings.Index(text, "мир")
	end := start + len("мир")

	detector := &staticDetector{
		name: "ru",
		findings: []llmguard.Finding{{
			Entity: llmguard.EntityPerson, Start: start, End: end, Confidence: 1,
		}},
	}
	guard, err := llmguard.New(llmguard.WithDetector(detector))
	require.NoError(t, err)

	findings, err := guard.Detect(context.Background(), text)
	require.NoError(t, err)
	require.Len(t, findings, 1)
	assert.Equal(t, start, findings[0].Start)
	assert.Equal(t, end, findings[0].End)
}

func TestGuard_WhenDetectorReturnsContextCanceled_ExpectSubstantiveDetectorError(t *testing.T) {
	t.Parallel()

	detector := &staticDetector{name: "self-cancel", err: context.Canceled}
	guard, err := llmguard.New(llmguard.WithDetector(detector))
	require.NoError(t, err)

	findings, err := guard.Detect(context.Background(), "text")
	require.Error(t, err)
	assert.Nil(t, findings)
	assert.ErrorIs(t, err, llmguard.ErrDetector)
	assert.ErrorIs(t, err, context.Canceled)

	detErr, ok := muerrors.As[*llmguard.DetectorError](err)
	require.True(t, ok)
	assert.Equal(t, "self-cancel", detErr.Detector)
}

func TestGuard_WhenDetectorReturnsDeadlineExceeded_ExpectSubstantiveDetectorError(t *testing.T) {
	t.Parallel()

	detector := &staticDetector{name: "timeout", err: context.DeadlineExceeded}
	guard, err := llmguard.New(llmguard.WithDetector(detector))
	require.NoError(t, err)

	findings, err := guard.Detect(context.Background(), "text")
	require.Error(t, err)
	assert.Nil(t, findings)
	assert.ErrorIs(t, err, llmguard.ErrDetector)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestGuard_WhenLaterSubstantiveFailure_ExpectNotHiddenByCooperativeCancel(t *testing.T) {
	t.Parallel()

	cooperativeStarted := make(chan struct{})
	substantiveReady := make(chan struct{})

	cooperative := &staticDetector{
		name: "cooperative",
		onDetect: func(ctx context.Context) {
			close(cooperativeStarted)
			<-ctx.Done()
		},
	}
	failing := &staticDetector{
		name: "failing",
		err:  errors.New("substantive"),
		onDetect: func(context.Context) {
			<-cooperativeStarted
			close(substantiveReady)
		},
	}

	guard, err := llmguard.New(llmguard.WithDetector(cooperative), llmguard.WithDetector(failing))
	require.NoError(t, err)

	_, err = guard.Detect(context.Background(), "text")
	require.Error(t, err)

	detErr, ok := muerrors.As[*llmguard.DetectorError](err)
	require.True(t, ok)
	assert.Equal(t, "failing", detErr.Detector)

	<-substantiveReady
}

type cooperativeDetector struct {
	name    string
	started chan struct{}
}

func (d *cooperativeDetector) Name() string { return d.name }

func (d *cooperativeDetector) Detect(ctx context.Context, text string) ([]llmguard.Finding, error) {
	close(d.started)
	<-ctx.Done()
	return nil, ctx.Err()
}

func TestGuard_WhenInvalidFinding_ExpectNotHiddenByCooperativeCancel(t *testing.T) {
	t.Parallel()

	cooperativeStarted := make(chan struct{})
	invalidReady := make(chan struct{})

	cooperative := &cooperativeDetector{
		name:    "cooperative",
		started: cooperativeStarted,
	}
	invalid := &staticDetector{
		name: "invalid",
		findings: []llmguard.Finding{{
			Entity: llmguard.EntityEmail, Start: 0, End: 0, Confidence: 1,
		}},
		onDetect: func(context.Context) {
			<-cooperativeStarted
			close(invalidReady)
		},
	}

	guard, err := llmguard.New(llmguard.WithDetector(cooperative), llmguard.WithDetector(invalid))
	require.NoError(t, err)

	findings, err := guard.Detect(context.Background(), "text")
	require.Error(t, err)
	assert.Nil(t, findings)
	assert.ErrorIs(t, err, llmguard.ErrInvalidFinding)

	_, isDetectorErr := muerrors.As[*llmguard.DetectorError](err)
	assert.False(t, isDetectorErr)

	<-invalidReady
}
