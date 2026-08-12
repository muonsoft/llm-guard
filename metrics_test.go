package llmguard_test

import (
	"context"
	"testing"

	"github.com/muonsoft/llm-guard"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type metricsObserver struct {
	requests      map[llmguard.Operation]int
	outcomes      map[llmguard.Outcome]int
	entityCounts  map[llmguard.EntityType]int
	actionCounts  map[llmguard.Action]int
	restoreMisses int
	blocks        int
}

func newMetricsObserver() *metricsObserver {
	return &metricsObserver{
		requests:     make(map[llmguard.Operation]int),
		outcomes:     make(map[llmguard.Outcome]int),
		entityCounts: make(map[llmguard.EntityType]int),
		actionCounts: make(map[llmguard.Action]int),
	}
}

func (m *metricsObserver) Observe(event llmguard.Event) {
	m.requests[event.Operation]++
	m.outcomes[event.Outcome]++
	for _, count := range event.EntityCounts {
		m.entityCounts[count.Entity] += count.Count
	}
	for _, count := range event.ActionCounts {
		m.actionCounts[count.Action] += count.Count
	}
	m.restoreMisses += event.RestoreMisses
	if event.Operation == llmguard.OperationMask && event.Outcome == llmguard.OutcomeBlocked {
		m.blocks++
	}
}

func TestMetrics_WhenFullLifecycle_ExpectCountableFields(t *testing.T) {
	t.Parallel()

	metrics := newMetricsObserver()
	guard, err := llmguard.New(
		llmguard.WithDetector(llmguard.NewEmailDetector()),
		llmguard.WithDetector(llmguard.NewAPIKeyDetector()),
		llmguard.WithObserver(metrics),
	)
	require.NoError(t, err)

	masked, err := guard.Mask(context.Background(), "mail a@b.co")
	require.NoError(t, err)

	token := "{{LLMG_0123456789abcdef0123456789abcdef_0001}}"
	_, err = guard.Restore(context.Background(), masked.Text+" "+token, masked.Tokens)
	require.NoError(t, err)

	_, err = guard.Mask(context.Background(), "ghp_AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	require.Error(t, err)

	assert.Equal(t, 2, metrics.requests[llmguard.OperationDetect])
	assert.Equal(t, 2, metrics.requests[llmguard.OperationMask])
	assert.Equal(t, 1, metrics.requests[llmguard.OperationRestore])
	assert.Equal(t, 1, metrics.blocks)
	assert.Equal(t, 1, metrics.restoreMisses)
	assert.Positive(t, metrics.entityCounts[llmguard.EntityEmail])
	assert.Positive(t, metrics.actionCounts[llmguard.ActionMask])
}
