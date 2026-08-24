package evaluation

import (
	"context"
	"errors"
	"regexp"
	"strings"

	muerrors "github.com/muonsoft/errors"
	"github.com/muonsoft/llm-guard"
)

var errPlaceholderAlignment = errors.New("mask/input placeholder alignment failed")

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

const collisionFakeToken = "{{LLMG_0123456789abcdef0123456789abcdef_9999}}"

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
		inferred, alignErr := inferPlaceholderMap(rec.Input, result.Text)
		if alignErr != nil {
			return fail(LifecycleOutcomeError, alignErr.Error())
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
			if expected := expectedRecipeRestore(result.Text, inferred); restored != expected {
				return fail(LifecycleOutcomeRestoreMiss, "identity restore mismatch")
			}
			return nil
		case "mutate_placeholder", "delete_placeholder", "collision":
			if response == result.Text {
				return fail(LifecycleOutcomeError, "recipe produced no placeholder change")
			}
			expected := expectedRecipeRestore(response, inferred)
			if restored != expected {
				return fail(LifecycleOutcomeMutation, "recipe restore mismatch")
			}
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

// inferPlaceholderMap aligns masked text with original input and returns
// placeholder token → original substring mappings.
func inferPlaceholderMap(input, masked string) (map[string]string, error) {
	locs := placeholderPattern.FindAllStringIndex(masked, -1)
	if len(locs) == 0 {
		if masked != input {
			return nil, errPlaceholderAlignment
		}
		return map[string]string{}, nil
	}

	result := make(map[string]string, len(locs))
	inputPos := 0
	maskedPos := 0
	for i, loc := range locs {
		plaintext := masked[maskedPos:loc[0]]
		if plaintext != "" {
			if inputPos+len(plaintext) > len(input) || input[inputPos:inputPos+len(plaintext)] != plaintext {
				return nil, errPlaceholderAlignment
			}
			inputPos += len(plaintext)
		}

		token := masked[loc[0]:loc[1]]
		var nextPlaintext string
		if i+1 < len(locs) {
			nextPlaintext = masked[loc[1]:locs[i+1][0]]
		} else {
			nextPlaintext = masked[loc[1]:]
		}

		var valueEnd int
		if nextPlaintext != "" {
			idx := strings.Index(input[inputPos:], nextPlaintext)
			if idx < 0 {
				return nil, errPlaceholderAlignment
			}
			valueEnd = inputPos + idx
		} else {
			valueEnd = len(input)
		}
		result[token] = input[inputPos:valueEnd]
		inputPos = valueEnd
		maskedPos = loc[1]
	}

	if maskedPos < len(masked) {
		trailing := masked[maskedPos:]
		if inputPos+len(trailing) > len(input) || input[inputPos:] != trailing {
			return nil, errPlaceholderAlignment
		}
		inputPos += len(trailing)
	}
	if inputPos != len(input) {
		return nil, errPlaceholderAlignment
	}
	return result, nil
}

// expectedRecipeRestore substitutes known placeholders in recipeText using inferred
// token→value pairs. Unknown placeholders remain unchanged (same as Restore).
func expectedRecipeRestore(recipeText string, inferred map[string]string) string {
	if len(inferred) == 0 {
		return recipeText
	}
	pairs := make([]string, 0, len(inferred)*2)
	for token, value := range inferred {
		pairs = append(pairs, token, value)
	}
	return strings.NewReplacer(pairs...).Replace(recipeText)
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
		return collisionFakeToken + " " + masked
	default:
		return masked
	}
}
