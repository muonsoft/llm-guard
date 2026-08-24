package evaluation

import (
	"context"
	"regexp"
	"strings"

	muerrors "github.com/muonsoft/errors"
	"github.com/muonsoft/llm-guard"
)

// LifecycleCaseOutcome is the terminal lifecycle result for one case.
type LifecycleCaseOutcome string

const (
	LifecycleOutcomeOK          LifecycleCaseOutcome = "ok"
	LifecycleOutcomeBlock       LifecycleCaseOutcome = "block"
	LifecycleOutcomeMutation    LifecycleCaseOutcome = "mutation"
	LifecycleOutcomeRestoreMiss LifecycleCaseOutcome = "restore_miss"
	LifecycleOutcomeError       LifecycleCaseOutcome = "error"
)

// LifecycleFailureDiagnostic describes one lifecycle failure without raw input.
type LifecycleFailureDiagnostic struct {
	SourceRecordID string
	ExpectedAction string
	ResponseRecipe string
	Outcome        LifecycleCaseOutcome
	Detail         string
	InputSHA256    string
}

// LifecycleReport is the lifecycle profile evaluation output.
type LifecycleReport struct {
	Profile        string
	SuiteID        string
	MappingVersion string
	SourceIDs      []string
	Cases          int
	OK             int
	Blocked        int
	MutationMiss   int
	RestoreMiss    int
	Errors         int
	Diagnostics    []LifecycleFailureDiagnostic
	ThresholdID    string
	Status         string
}

// HasLifecycleRegression returns true when any lifecycle case failed expectations.
func (r LifecycleReport) HasLifecycleRegression() bool {
	return len(r.Diagnostics) > 0
}

var placeholderPattern = regexp.MustCompile(`\{\{LLMG_[0-9a-f]{32}_\d+\}\}`)

// EvaluateLifecycle runs mask/block/restore checks on normalized suite cases.
func EvaluateLifecycle(ctx context.Context, guard *llmguard.Guard, suite Suite) (LifecycleReport, error) {
	report := LifecycleReport{
		Profile:        string(ProfileLifecycle),
		SuiteID:        suite.SuiteID,
		MappingVersion: suite.MappingVersion,
		SourceIDs:      append([]string(nil), suite.SourceIDs...),
		Cases:          len(suite.Records),
		Status:         StatusPass,
	}
	for _, rec := range suite.Records {
		lc := defaultLifecycleExpectations(rec)
		if lc == nil {
			continue
		}
		diag := evaluateLifecycleCase(ctx, guard, rec, *lc)
		if diag != nil {
			report.Diagnostics = append(report.Diagnostics, *diag)
			switch diag.Outcome {
			case LifecycleOutcomeError:
				report.Errors++
			case LifecycleOutcomeRestoreMiss, LifecycleOutcomeMutation:
				report.RestoreMiss++
			default:
				report.Errors++
			}
			continue
		}
		switch lc.ExpectedAction {
		case "block":
			report.Blocked++
		case "mask":
			switch lc.ResponseRecipe {
			case "mutate_placeholder", "delete_placeholder", "collision":
				report.MutationMiss++
			default:
				report.OK++
			}
		}
	}
	if report.HasLifecycleRegression() {
		report.Status = StatusFail
	}
	return report, nil
}

func defaultLifecycleExpectations(rec SuiteRecord) *SuiteLifecycle {
	if rec.Lifecycle != nil {
		return rec.Lifecycle
	}
	hasSupportedPII := false
	hasSupportedSecret := false
	for _, ann := range rec.Annotations {
		if ann.Disposition != DispositionSupported {
			continue
		}
		switch llmguard.EntityType(ann.MappedEntity) {
		case llmguard.EntitySecretJWT, llmguard.EntitySecretPrivateKey, llmguard.EntitySecretAPIKey, llmguard.EntityConnectionString:
			hasSupportedSecret = true
		default:
			hasSupportedPII = true
		}
	}
	if hasSupportedSecret {
		return &SuiteLifecycle{ExpectedAction: "block"}
	}
	if hasSupportedPII {
		return &SuiteLifecycle{ExpectedAction: "mask", ResponseRecipe: "identity"}
	}
	return nil
}

func evaluateLifecycleCase(ctx context.Context, guard *llmguard.Guard, rec SuiteRecord, lc SuiteLifecycle) *LifecycleFailureDiagnostic {
	fail := func(outcome LifecycleCaseOutcome, detail string) *LifecycleFailureDiagnostic {
		return &LifecycleFailureDiagnostic{
			SourceRecordID: rec.SourceRecordID,
			ExpectedAction: lc.ExpectedAction,
			ResponseRecipe: lc.ResponseRecipe,
			Outcome:        outcome,
			Detail:         detail,
			InputSHA256:    rec.InputSHA256,
		}
	}

	switch lc.ExpectedAction {
	case "block":
		result, err := guard.Mask(ctx, rec.Input)
		if err == nil {
			if result.Text != "" {
				return fail(LifecycleOutcomeError, "block expected zero outbound text")
			}
			return fail(LifecycleOutcomeError, "block expected error")
		}
		if !muerrors.Is(err, llmguard.ErrBlocked) {
			return fail(LifecycleOutcomeError, "block error not ErrBlocked")
		}
		if result.Text != "" {
			return fail(LifecycleOutcomeError, "block returned non-empty result text")
		}
		if errStrContainsSecret(err, rec) {
			return fail(LifecycleOutcomeError, "block error contains secret value")
		}
		return nil
	case "mask":
		result, err := guard.Mask(ctx, rec.Input)
		if err != nil {
			return fail(LifecycleOutcomeError, "mask failed")
		}
		if containsProtectedSubstrings(rec, result.Text) {
			return fail(LifecycleOutcomeMutation, "masked text contains protected original substring")
		}
		response := applyResponseRecipe(result.Text, lc.ResponseRecipe)
		restored, err := guard.Restore(ctx, response, result.Tokens)
		if err != nil {
			return fail(LifecycleOutcomeError, "restore failed")
		}
		switch lc.ResponseRecipe {
		case "", "identity":
			if restored != rec.Input {
				return fail(LifecycleOutcomeRestoreMiss, "identity restore mismatch")
			}
			return nil
		case "mutate_placeholder", "delete_placeholder", "collision":
			if restored == rec.Input {
				return fail(LifecycleOutcomeMutation, "expected mutation/miss outcome")
			}
			return nil
		default:
			return fail(LifecycleOutcomeError, "unknown response recipe")
		}
	default:
		return fail(LifecycleOutcomeError, "unknown expected action")
	}
}

func containsProtectedSubstrings(rec SuiteRecord, text string) bool {
	for _, ann := range rec.Annotations {
		if ann.Disposition != DispositionSupported {
			continue
		}
		if ann.Start < 0 || ann.End > len(rec.Input) || ann.End <= ann.Start {
			continue
		}
		sub := rec.Input[ann.Start:ann.End]
		if sub != "" && strings.Contains(text, sub) {
			return true
		}
	}
	return false
}

func errStrContainsSecret(err error, rec SuiteRecord) bool {
	msg := err.Error()
	for _, ann := range rec.Annotations {
		if ann.Disposition != DispositionSupported {
			continue
		}
		if ann.Start < 0 || ann.End > len(rec.Input) {
			continue
		}
		sub := rec.Input[ann.Start:ann.End]
		if len(sub) >= 8 && strings.Contains(msg, sub) {
			return true
		}
	}
	return false
}

func applyResponseRecipe(masked, recipe string) string {
	switch recipe {
	case "", "identity":
		return masked
	case "mutate_placeholder":
		loc := placeholderPattern.FindStringIndex(masked)
		if loc == nil {
			return masked
		}
		token := masked[loc[0]:loc[1]]
		lastDigit := -1
		for i := len(token) - 1; i >= 0; i-- {
			if token[i] >= '0' && token[i] <= '9' {
				lastDigit = i
				break
			}
		}
		if lastDigit < 0 {
			return masked
		}
		replacement := byte('0')
		if token[lastDigit] == '0' {
			replacement = '1'
		}
		mutated := token[:lastDigit] + string(replacement) + token[lastDigit+1:]
		return masked[:loc[0]] + mutated + masked[loc[1]:]
	case "delete_placeholder":
		loc := placeholderPattern.FindStringIndex(masked)
		if loc == nil {
			return masked
		}
		return masked[:loc[0]] + masked[loc[1]:]
	case "collision":
		fake := "{{LLMG_0123456789abcdef0123456789abcdef_9999}}"
		return fake + " " + masked
	default:
		return masked
	}
}
