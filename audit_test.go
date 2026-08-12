package llmguard_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/muonsoft/llm-guard"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const auditMarker = "AUDIT_MARKER_7F3C2A91_SECRET_FRAGMENT"

type unsafeCapture struct {
	mu     sync.Mutex
	events []llmguard.UnsafeDevelopmentEvent
}

func (c *unsafeCapture) ObserveUnsafe(event llmguard.UnsafeDevelopmentEvent) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.events = append(c.events, event)
}

func TestAudit_WhenSafeObserverOnly_ExpectNoUnsafeCallbacks(t *testing.T) {
	t.Parallel()

	unsafeObserver := &unsafeCapture{}
	guard, err := llmguard.New(
		llmguard.WithDetector(llmguard.NewEmailDetector()),
		llmguard.WithObserver(llmguard.ObserverFunc(func(llmguard.Event) {})),
	)
	require.NoError(t, err)

	_, err = guard.Detect(context.Background(), auditMarker+"@example.com")
	require.NoError(t, err)
	assert.Empty(t, unsafeObserver.events)
}

func TestAudit_WhenUnsafeObserverConfigured_ExpectRawDiagnostics(t *testing.T) {
	t.Parallel()

	unsafeObserver := &unsafeCapture{}
	guard, err := llmguard.New(
		llmguard.WithDetector(llmguard.NewEmailDetector()),
		llmguard.WithUnsafeDevelopmentObserver(unsafeObserver),
	)
	require.NoError(t, err)

	text := "mail " + auditMarker + "@example.com"
	_, err = guard.Detect(context.Background(), text)
	require.NoError(t, err)

	require.Len(t, unsafeObserver.events, 1)
	assert.Contains(t, unsafeObserver.events[0].Input, auditMarker)
}

func TestAudit_WhenNilUnsafeObserver_ExpectInvalidConfig(t *testing.T) {
	t.Parallel()

	_, err := llmguard.New(llmguard.WithUnsafeDevelopmentObserver(nil))
	require.Error(t, err)
	assert.ErrorIs(t, err, llmguard.ErrInvalidConfig)
}

func TestAudit_WhenTypedNilUnsafeObserver_ExpectInvalidConfig(t *testing.T) {
	t.Parallel()

	var observer *unsafeCapture
	_, err := llmguard.New(llmguard.WithUnsafeDevelopmentObserver(observer))
	require.Error(t, err)
	assert.ErrorIs(t, err, llmguard.ErrInvalidConfig)

	var fn llmguard.UnsafeDevelopmentObserverFunc
	_, err = llmguard.New(llmguard.WithUnsafeDevelopmentObserver(fn))
	require.Error(t, err)
	assert.ErrorIs(t, err, llmguard.ErrInvalidConfig)
}

func TestAudit_WhenDuplicateUnsafeObserver_ExpectInvalidConfig(t *testing.T) {
	t.Parallel()

	_, err := llmguard.New(
		llmguard.WithUnsafeDevelopmentObserver(llmguard.UnsafeDevelopmentObserverFunc(func(llmguard.UnsafeDevelopmentEvent) {})),
		llmguard.WithUnsafeDevelopmentObserver(llmguard.UnsafeDevelopmentObserverFunc(func(llmguard.UnsafeDevelopmentEvent) {})),
	)
	require.Error(t, err)
	assert.ErrorIs(t, err, llmguard.ErrInvalidConfig)
}

func TestRedact_WhenSafeLifecyclePaths_ExpectNoMarkerLeakage(t *testing.T) {
	t.Parallel()

	marker := auditMarker
	observer := &captureObserver{}
	guard, err := llmguard.New(
		llmguard.WithDetector(&staticDetector{name: "custom", findings: []llmguard.Finding{{
			Entity: llmguard.EntityEmail, Start: 0, End: len(marker), Confidence: 0.9,
		}}}),
		llmguard.WithDetector(llmguard.NewAPIKeyDetector()),
		llmguard.WithObserver(observer),
	)
	require.NoError(t, err)

	scenarios := []struct {
		name      string
		run       func() error
		wantError bool
	}{
		{
			name: "detect_success",
			run: func() error {
				_, err := guard.Detect(context.Background(), marker)
				return err
			},
		},
		{
			name:      "detect_invalid",
			wantError: true,
			run: func() error {
				_, err := guard.Detect(context.Background(), "\xff")
				return err
			},
		},
		{
			name:      "mask_block",
			wantError: true,
			run: func() error {
				_, err := guard.Mask(context.Background(), marker+" ghp_AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
				return err
			},
		},
		{
			name: "restore_miss",
			run: func() error {
				token := "{{LLMG_0123456789abcdef0123456789abcdef_0001}}"
				_, err := guard.Restore(context.Background(), marker+token, &llmguard.TokenSet{})
				return err
			},
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			err := scenario.run()
			if scenario.wantError {
				require.Error(t, err)
				assertNoMarkerInError(t, marker, err)
			} else {
				require.NoError(t, err)
			}
			for _, event := range observer.snapshot() {
				assertNoMarker(t, marker, event)
			}
			observer.events = nil
		})
	}
}

func TestRedact_WhenSafeErrorsFormatted_ExpectNoMarkerLeakage(t *testing.T) {
	t.Parallel()

	marker := auditMarker
	detectorCause := "DETECTOR_CAUSE_SHOULD_NOT_LEAK_7F3C"

	scenarios := []struct {
		name string
		run  func() error
	}{
		{
			name: "detector_error",
			run: func() error {
				guard, err := llmguard.New(llmguard.WithDetector(&staticDetector{
					name: "broken",
					err:  errors.New(detectorCause),
				}))
				if err != nil {
					return err
				}
				_, err = guard.Detect(context.Background(), marker)
				return err
			},
		},
		{
			name: "policy_block",
			run: func() error {
				guard, err := llmguard.New(llmguard.WithDetector(llmguard.NewAPIKeyDetector()))
				if err != nil {
					return err
				}
				_, err = guard.Mask(context.Background(), marker+" ghp_AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
				return err
			},
		},
		{
			name: "invalid_token_set",
			run: func() error {
				guard, err := llmguard.New()
				if err != nil {
					return err
				}
				_, err = guard.Restore(context.Background(), marker, nil)
				return err
			},
		},
		{
			name: "invalid_restore_text",
			run: func() error {
				guard, err := llmguard.New()
				if err != nil {
					return err
				}
				_, err = guard.Restore(context.Background(), marker+"\xff", &llmguard.TokenSet{})
				return err
			},
		},
		{
			name: "invalid_detect_text",
			run: func() error {
				guard, err := llmguard.New(llmguard.WithDetector(llmguard.NewEmailDetector()))
				if err != nil {
					return err
				}
				_, err = guard.Detect(context.Background(), marker+"\xff")
				return err
			},
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			err := scenario.run()
			require.Error(t, err)
			assertNoMarkerInError(t, marker, err)
			assertNotContainsAny(t, err, detectorCause)
		})
	}
}

func assertNoMarkerInError(t *testing.T, marker string, err error) {
	t.Helper()
	require.Error(t, err)
	for _, form := range []string{
		err.Error(),
		fmt.Sprintf("%v", err),
		fmt.Sprintf("%+v", err),
		fmt.Sprintf("%#v", err),
	} {
		assert.NotContains(t, form, marker)
	}
}

func assertNotContainsAny(t *testing.T, err error, values ...string) {
	t.Helper()
	for _, value := range values {
		for _, form := range []string{
			err.Error(),
			fmt.Sprintf("%v", err),
			fmt.Sprintf("%+v", err),
			fmt.Sprintf("%#v", err),
		} {
			assert.NotContains(t, form, value)
		}
	}
}

func assertNoMarker(t *testing.T, marker string, event llmguard.Event) {
	t.Helper()
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

func TestRedact_WhenMaskResultAndTokenSetFormatted_ExpectRedacted(t *testing.T) {
	t.Parallel()

	marker := auditMarker
	guard, err := llmguard.New(
		llmguard.WithDetector(&staticDetector{name: "custom", findings: []llmguard.Finding{{
			Entity: llmguard.EntityEmail, Start: 0, End: len(marker), Confidence: 0.9,
		}}}),
		llmguard.WithSecretAction(llmguard.ActionMask),
	)
	require.NoError(t, err)

	result, err := guard.Mask(context.Background(), marker)
	require.NoError(t, err)

	for _, value := range []string{
		fmt.Sprintf("%v", result),
		fmt.Sprintf("%+v", result),
		fmt.Sprintf("%#v", result),
	} {
		assert.NotContains(t, value, marker)
		assert.NotContains(t, value, "{{LLMG_")
	}
	data, err := json.Marshal(result)
	require.NoError(t, err)
	assert.NotContains(t, string(data), marker)
	assert.NotContains(t, string(data), "{{LLMG_")

	tokenForms := []string{
		fmt.Sprintf("%v", result.Tokens),
		fmt.Sprintf("%+v", result.Tokens),
		fmt.Sprintf("%#v", result.Tokens),
		result.Tokens.String(),
		result.Tokens.GoString(),
	}
	tokenJSON, err := json.Marshal(result.Tokens)
	require.NoError(t, err)
	tokenForms = append(tokenForms, string(tokenJSON))
	for _, form := range tokenForms {
		assert.NotContains(t, form, marker)
		assert.True(t, strings.Contains(form, "TokenSet") || strings.Contains(form, "tokens"))
	}
}

func TestAudit_WhenConcurrentUnsafeObserver_ExpectRaceSafeCapture(t *testing.T) {
	t.Parallel()

	unsafeObserver := &unsafeCapture{}
	guard, err := llmguard.New(
		llmguard.WithDetector(llmguard.NewEmailDetector()),
		llmguard.WithUnsafeDevelopmentObserver(unsafeObserver),
	)
	require.NoError(t, err)

	var wg sync.WaitGroup
	errCh := make(chan error, 16)
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, detectErr := guard.Detect(context.Background(), "a@b.co")
			errCh <- detectErr
		}()
	}
	wg.Wait()
	close(errCh)
	for detectErr := range errCh {
		require.NoError(t, detectErr)
	}
	assert.Len(t, unsafeObserver.events, 16)
}
