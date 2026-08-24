package evaluation

import (
	"context"

	"github.com/muonsoft/llm-guard"
)

// NewMVPGuard constructs a Guard with all built-in MVP detectors registered.
func NewMVPGuard() (*llmguard.Guard, error) {
	return llmguard.New(
		llmguard.WithDetector(llmguard.NewPersonDetector()),
		llmguard.WithDetector(llmguard.NewAddressDetector()),
		llmguard.WithDetector(llmguard.NewEmailDetector()),
		llmguard.WithDetector(llmguard.NewPhoneDetector()),
		llmguard.WithDetector(llmguard.NewIPDetector()),
		llmguard.WithDetector(llmguard.NewURLDetector()),
		llmguard.WithDetector(llmguard.NewINNDetector()),
		llmguard.WithDetector(llmguard.NewSNILSDetector()),
		llmguard.WithDetector(llmguard.NewPassportDetector()),
		llmguard.WithDetector(llmguard.NewBankCardDetector()),
		llmguard.WithDetector(llmguard.NewBankAccountDetector()),
		llmguard.WithDetector(llmguard.NewDateOfBirthDetector()),
		llmguard.WithDetector(llmguard.NewJWTDetector()),
		llmguard.WithDetector(llmguard.NewPEMPrivateKeyDetector()),
		llmguard.WithDetector(llmguard.NewAPIKeyDetector()),
		llmguard.WithDetector(llmguard.NewDSNDetector()),
	)
}

// Evaluate runs Detect → Resolve matching for each corpus case.
func Evaluate(ctx context.Context, guard *llmguard.Guard, cases []Case) (Report, error) {
	stats := make(map[llmguard.EntityType]*entityAccumulator, len(mvpEntityOrder))
	for _, entity := range mvpEntityOrder {
		stats[entity] = &entityAccumulator{}
	}

	for _, tc := range cases {
		resolved, err := DetectResolve(ctx, guard, tc.Input)
		if err != nil {
			return Report{}, err
		}

		expectedByEntity := groupExpected(tc.Expected)
		predictedByEntity := groupPredicted(resolved)

		for _, entity := range mvpEntityOrder[:] {
			acc := stats[entity]
			expected := expectedByEntity[entity]
			predicted := predictedByEntity[entity]

			if len(expected) == 0 {
				acc.negativeCases++
				if len(predicted) > 0 {
					acc.falsePositiveCases++
				}
			}

			tp, fp, fn := compareSpans(expected, predicted)
			acc.tp += tp
			acc.fp += fp
			acc.fn += fn
		}
	}

	report := Report{Cases: len(cases), Coverage: BuildCoverage(cases)}
	for _, entity := range mvpEntityOrder[:] {
		acc := stats[entity]
		metrics := EntityMetrics{
			Entity:             entity,
			TP:                 acc.tp,
			FP:                 acc.fp,
			FN:                 acc.fn,
			NegativeCases:      acc.negativeCases,
			FalsePositiveCases: acc.falsePositiveCases,
			Precision:          ratio(acc.tp, acc.tp+acc.fp),
			Recall:             ratio(acc.tp, acc.tp+acc.fn),
			FPR:                ratio(acc.falsePositiveCases, acc.negativeCases),
			FNR:                ratio(acc.fn, acc.tp+acc.fn),
		}
		metrics.F1 = f1(metrics.Precision, metrics.Recall)
		report.Entities = append(report.Entities, metrics)
		report.Summary.TP += acc.tp
		report.Summary.FP += acc.fp
		report.Summary.FN += acc.fn
	}
	return report, nil
}

type entityAccumulator struct {
	tp                 int
	fp                 int
	fn                 int
	negativeCases      int
	falsePositiveCases int
}

func groupExpected(spans []ExpectedSpan) map[llmguard.EntityType][]spanKey {
	out := make(map[llmguard.EntityType][]spanKey)
	for _, span := range spans {
		out[span.Entity] = append(out[span.Entity], spanKeyFromExpected(span))
	}
	return out
}

func groupPredicted(findings []llmguard.Finding) map[llmguard.EntityType][]spanKey {
	out := make(map[llmguard.EntityType][]spanKey)
	for _, finding := range findings {
		out[finding.Entity] = append(out[finding.Entity], spanKeyFromFinding(finding))
	}
	return out
}

func compareSpans(expected, predicted []spanKey) (tp, fp, fn int) {
	expectedSet := make(map[spanKey]struct{}, len(expected))
	for _, span := range expected {
		expectedSet[span] = struct{}{}
	}
	predictedSet := make(map[spanKey]struct{}, len(predicted))
	for _, span := range predicted {
		predictedSet[span] = struct{}{}
	}
	for span := range predictedSet {
		if _, ok := expectedSet[span]; ok {
			tp++
		} else {
			fp++
		}
	}
	for span := range expectedSet {
		if _, ok := predictedSet[span]; !ok {
			fn++
		}
	}
	return tp, fp, fn
}

func ratio(numerator, denominator int) float64 {
	if denominator == 0 {
		return 0
	}
	return float64(numerator) / float64(denominator)
}

func f1(precision, recall float64) float64 {
	if precision+recall == 0 {
		return 0
	}
	return 2 * precision * recall / (precision + recall)
}
