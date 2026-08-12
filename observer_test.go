package llmguard_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/muonsoft/llm-guard"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type captureObserver struct {
	mu     sync.Mutex
	events []llmguard.Event
}

func (c *captureObserver) Observe(event llmguard.Event) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.events = append(c.events, event)
}

func (c *captureObserver) snapshot() []llmguard.Event {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]llmguard.Event, len(c.events))
	copy(out, c.events)
	return out
}

func TestObserver_WhenNoObserverConfigured_ExpectNoCallbacks(t *testing.T) {
	t.Parallel()

	guard, err := llmguard.New(llmguard.WithDetector(llmguard.NewEmailDetector()))
	require.NoError(t, err)

	_, err = guard.Detect(context.Background(), "contact a@b.co")
	require.NoError(t, err)
}

func TestObserver_WhenDetectSuccess_ExpectTerminalEvent(t *testing.T) {
	t.Parallel()

	observer := &captureObserver{}
	guard, err := llmguard.New(
		llmguard.WithDetector(llmguard.NewEmailDetector()),
		llmguard.WithObserver(observer),
	)
	require.NoError(t, err)

	text := "contact a@b.co"
	_, err = guard.Detect(context.Background(), text)
	require.NoError(t, err)

	events := observer.snapshot()
	require.Len(t, events, 1)
	event := events[0]
	assert.Equal(t, llmguard.OperationDetect, event.Operation)
	assert.Equal(t, llmguard.OutcomeSuccess, event.Outcome)
	assert.Equal(t, len(text), event.InputBytes)
	assert.Zero(t, event.OutputBytes)
	assert.Equal(t, 1, event.FindingCount)
	require.Len(t, event.EntityCounts, 1)
	assert.Equal(t, llmguard.EntityEmail, event.EntityCounts[0].Entity)
	assert.Positive(t, event.Duration)
}

func TestObserver_WhenDetectInvalidText_ExpectErrorOutcome(t *testing.T) {
	t.Parallel()

	observer := &captureObserver{}
	guard, err := llmguard.New(
		llmguard.WithDetector(llmguard.NewEmailDetector()),
		llmguard.WithObserver(observer),
	)
	require.NoError(t, err)

	_, err = guard.Detect(context.Background(), "\xff")
	require.Error(t, err)
	events := observer.snapshot()
	require.Len(t, events, 1)
	assert.Equal(t, llmguard.OutcomeError, events[0].Outcome)
}

func TestObserver_WhenDetectDetectorError_ExpectErrorOutcome(t *testing.T) {
	t.Parallel()

	observer := &captureObserver{}
	guard, err := llmguard.New(
		llmguard.WithDetector(&staticDetector{name: "broken", err: errors.New("boom")}),
		llmguard.WithObserver(observer),
	)
	require.NoError(t, err)

	_, err = guard.Detect(context.Background(), "plain")
	require.Error(t, err)
	events := observer.snapshot()
	require.Len(t, events, 1)
	assert.Equal(t, llmguard.OutcomeError, events[0].Outcome)
}

func TestObserver_WhenNoDetectors_ExpectSuccessEvent(t *testing.T) {
	t.Parallel()

	observer := &captureObserver{}
	guard, err := llmguard.New(llmguard.WithObserver(observer))
	require.NoError(t, err)

	_, err = guard.Detect(context.Background(), "plain")
	require.NoError(t, err)
	events := observer.snapshot()
	require.Len(t, events, 1)
	assert.Equal(t, llmguard.OutcomeSuccess, events[0].Outcome)
	assert.Zero(t, events[0].FindingCount)
}

func TestObserver_WhenMaskBlocked_ExpectBlockedOutcome(t *testing.T) {
	t.Parallel()

	observer := &captureObserver{}
	guard, err := llmguard.New(
		llmguard.WithDetector(llmguard.NewAPIKeyDetector()),
		llmguard.WithObserver(observer),
	)
	require.NoError(t, err)

	text := "export KEY=ghp_AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
	_, err = guard.Mask(context.Background(), text)
	require.Error(t, err)
	assert.ErrorIs(t, err, llmguard.ErrBlocked)

	events := observer.snapshot()
	require.Len(t, events, 2)
	maskEvent := events[1]
	assert.Equal(t, llmguard.OperationMask, maskEvent.Operation)
	assert.Equal(t, llmguard.OutcomeBlocked, maskEvent.Outcome)
	assert.Zero(t, maskEvent.OutputBytes)
}

func TestObserver_WhenRestoreMiss_ExpectRestoreMissOutcome(t *testing.T) {
	t.Parallel()

	observer := &captureObserver{}
	guard, err := llmguard.New(
		llmguard.WithDetector(llmguard.NewEmailDetector()),
		llmguard.WithObserver(observer),
	)
	require.NoError(t, err)

	token := "{{LLMG_0123456789abcdef0123456789abcdef_0001}}"
	response := "hello " + token + " " + token
	restored, err := guard.Restore(context.Background(), response, &llmguard.TokenSet{})
	require.NoError(t, err)
	assert.Equal(t, response, restored)

	events := observer.snapshot()
	require.Len(t, events, 1)
	assert.Equal(t, llmguard.OperationRestore, events[0].Operation)
	assert.Equal(t, llmguard.OutcomeRestoreMiss, events[0].Outcome)
	assert.Equal(t, 2, events[0].RestoreMisses)
}

func TestObserver_WhenRestoredLiteralPlaceholderValue_ExpectSuccessOutcome(t *testing.T) {
	t.Parallel()

	observer := &captureObserver{}
	guard, err := llmguard.New(llmguard.WithObserver(observer))
	require.NoError(t, err)

	token := "{{LLMG_0123456789abcdef0123456789abcdef_0001}}"
	literal := "{{LLMG_0123456789abcdef0123456789abcdef_9999}}"
	tokens := llmguard.NewTokenSetForTest(token, literal)

	restored, err := guard.Restore(context.Background(), "value "+token, tokens)
	require.NoError(t, err)
	assert.Equal(t, "value "+literal, restored)

	events := observer.snapshot()
	require.Len(t, events, 1)
	assert.Equal(t, llmguard.OutcomeSuccess, events[0].Outcome)
	assert.Zero(t, events[0].RestoreMisses)
}

func TestObserver_WhenNilObserverOption_ExpectInvalidConfig(t *testing.T) {
	t.Parallel()

	_, err := llmguard.New(llmguard.WithObserver(nil))
	require.Error(t, err)
	assert.ErrorIs(t, err, llmguard.ErrInvalidConfig)
}

func TestObserver_WhenTypedNilObserver_ExpectInvalidConfig(t *testing.T) {
	t.Parallel()

	var observer *captureObserver
	_, err := llmguard.New(llmguard.WithObserver(observer))
	require.Error(t, err)
	assert.ErrorIs(t, err, llmguard.ErrInvalidConfig)

	var fn llmguard.ObserverFunc
	_, err = llmguard.New(llmguard.WithObserver(fn))
	require.Error(t, err)
	assert.ErrorIs(t, err, llmguard.ErrInvalidConfig)
}

func TestObserver_WhenDuplicateObserver_ExpectInvalidConfig(t *testing.T) {
	t.Parallel()

	_, err := llmguard.New(
		llmguard.WithObserver(llmguard.ObserverFunc(func(llmguard.Event) {})),
		llmguard.WithObserver(llmguard.ObserverFunc(func(llmguard.Event) {})),
	)
	require.Error(t, err)
	assert.ErrorIs(t, err, llmguard.ErrInvalidConfig)
}

func TestObserver_WhenConcurrentCalls_ExpectSeparateEvents(t *testing.T) {
	t.Parallel()

	observer := &captureObserver{}
	guard, err := llmguard.New(
		llmguard.WithDetector(llmguard.NewEmailDetector()),
		llmguard.WithObserver(observer),
	)
	require.NoError(t, err)

	var wg sync.WaitGroup
	errCh := make(chan error, 8)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, detectErr := guard.Detect(context.Background(), "mail a@b.co")
			errCh <- detectErr
		}()
	}
	wg.Wait()
	close(errCh)
	for detectErr := range errCh {
		require.NoError(t, detectErr)
	}
	assert.Len(t, observer.snapshot(), 8)
}

func TestObserver_WhenCustomEntityMarker_ExpectCustomBucketOnly(t *testing.T) {
	t.Parallel()

	marker := "SENSITIVE_MARKER_OBSERVER_7F3C"
	customEntity := llmguard.EntityType(marker)
	observer := &captureObserver{}
	guard, err := llmguard.New(
		llmguard.WithDetector(&staticDetector{name: "custom", findings: []llmguard.Finding{{
			Entity: customEntity, Start: 0, End: len(marker), Confidence: 0.9,
		}}}),
		llmguard.WithObserver(observer),
	)
	require.NoError(t, err)

	_, err = guard.Detect(context.Background(), marker)
	require.NoError(t, err)
	event := observer.snapshot()[0]
	require.Len(t, event.EntityCounts, 1)
	assert.Equal(t, llmguard.EntityCustom, event.EntityCounts[0].Entity)

	data, err := json.Marshal(event)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"entity":"CUSTOM"`)
	assert.NotContains(t, string(data), marker)
	for _, form := range []string{fmt.Sprintf("%v", event), fmt.Sprintf("%+v", event), fmt.Sprintf("%#v", event)} {
		assert.NotContains(t, form, marker)
	}
}

func TestObserver_WhenEventFormatted_ExpectNoSensitiveSubstrings(t *testing.T) {
	t.Parallel()

	marker := "SENSITIVE_MARKER_OBSERVER_7F3C"
	observer := &captureObserver{}
	guard, err := llmguard.New(
		llmguard.WithDetector(&staticDetector{name: "custom", findings: []llmguard.Finding{{
			Entity: llmguard.EntityType("CUSTOM"), Start: 0, End: len(marker), Confidence: 0.9,
		}}}),
		llmguard.WithObserver(observer),
	)
	require.NoError(t, err)

	_, err = guard.Detect(context.Background(), marker)
	require.NoError(t, err)
	event := observer.snapshot()[0]

	forms := []string{
		fmt.Sprintf("%v", event),
		fmt.Sprintf("%+v", event),
		fmt.Sprintf("%#v", event),
	}
	data, err := json.Marshal(event)
	require.NoError(t, err)
	forms = append(forms, string(data))

	for _, form := range forms {
		assert.NotContains(t, form, marker)
	}
}

type slowDetector struct {
	name    string
	invoked atomic.Bool
	release chan struct{}
}

func (d *slowDetector) Name() string { return d.name }

func (d *slowDetector) Detect(ctx context.Context, text string) ([]llmguard.Finding, error) {
	d.invoked.Store(true)
	select {
	case <-d.release:
		return nil, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func TestObserver_WhenDetectCancelled_ExpectErrorOutcome(t *testing.T) {
	t.Parallel()

	observer := &captureObserver{}
	release := make(chan struct{})
	detector := &slowDetector{name: "slow", release: release}
	guard, err := llmguard.New(
		llmguard.WithDetector(detector),
		llmguard.WithObserver(observer),
	)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() {
		_, detectErr := guard.Detect(ctx, "wait")
		errCh <- detectErr
	}()

	require.Eventually(t, detector.invoked.Load, time.Second, 10*time.Millisecond)
	cancel()
	require.Error(t, <-errCh)

	events := observer.snapshot()
	require.Len(t, events, 1)
	assert.Equal(t, llmguard.OutcomeError, events[0].Outcome)
	close(release)
}
