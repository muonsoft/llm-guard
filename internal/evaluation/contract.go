package evaluation

import (
	"context"
	"sort"

	"github.com/muonsoft/llm-guard"
)

// EvaluateContract runs the contract profile on a normalized suite.
func EvaluateContract(ctx context.Context, guard *llmguard.Guard, suite Suite) (ContractReport, error) {
	scope := entityScopeOrder(suite.Scope)
	stats := make(map[llmguard.EntityType]*entityAccumulator, len(scope))
	for _, entity := range scope {
		stats[entity] = &entityAccumulator{}
	}
	var overlapPairs int
	var failures []ContractFailureDiagnostic
	sourceSummaries := make(sourceContractSummaries)

	for _, rec := range suite.Records {
		resolved, err := DetectResolve(ctx, guard, rec.Input)
		if err != nil {
			return ContractReport{}, err
		}
		recScope := recordScope(rec, suite.Scope)
		goldByEntity := contractGoldSpans(rec, recScope)
		predictedByEntity := predictedSpansByEntity(resolved)
		overlapPairs += countOverlappingPairs(goldByEntity, predictedByEntity)
		failures = append(failures, contractFailureDiagnostics(rec, goldByEntity, predictedByEntity, recScope)...)

		var recTP, recFP, recFN int
		for _, entity := range scope {
			if !entityInScope(entity, recScope) {
				continue
			}
			acc := stats[entity]
			expected := goldByEntity[entity]
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
			recTP += tp
			recFP += fp
			recFN += fn
		}
		src := sourceSummaries[rec.SourceID]
		src.TP += recTP
		src.FP += recFP
		src.FN += recFN
		sourceSummaries[rec.SourceID] = src
	}

	report := ContractReport{
		Profile:         string(ProfileContract),
		SuiteID:         suite.SuiteID,
		MappingVersion:  suite.MappingVersion,
		SourceIDs:       append([]string(nil), suite.SourceIDs...),
		Cases:           len(suite.Records),
		sourceSummaries: sourceSummaries,
		Diagnostics: ContractDiagnostics{
			OverlappingPairs: overlapPairs,
			Failures:         failures,
		},
	}
	for _, entity := range scope {
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
	if report.HasContractRegression() {
		report.Status = StatusFail
	} else {
		report.Status = StatusPass
	}
	return report, nil
}

func contractGoldSpans(rec SuiteRecord, recScope map[llmguard.EntityType]struct{}) map[llmguard.EntityType][]spanKey {
	gold := make(map[llmguard.EntityType][]spanKey)
	for _, ann := range rec.Annotations {
		if ann.Disposition != DispositionSupported {
			continue
		}
		entity := llmguard.EntityType(ann.MappedEntity)
		if !entityInScope(entity, recScope) {
			continue
		}
		gold[entity] = append(gold[entity], spanKey{entity: entity, start: ann.Start, end: ann.End})
	}
	return gold
}

func predictedSpansByEntity(findings []llmguard.Finding) map[llmguard.EntityType][]spanKey {
	return groupPredicted(findings)
}

func countOverlappingPairs(goldByEntity map[llmguard.EntityType][]spanKey, predictedByEntity map[llmguard.EntityType][]spanKey) int {
	if predictedByEntity == nil {
		return 0
	}
	total := 0
	for entity, goldSpans := range goldByEntity {
		predicted := predictedByEntity[entity]
		for _, g := range goldSpans {
			gi := byteInterval{start: g.start, end: g.end}
			for _, p := range predicted {
				pi := byteInterval{start: p.start, end: p.end}
				if intervalsOverlap(gi, pi) && !intervalsEqual(gi, pi) {
					total++
				}
			}
		}
	}
	return total
}

func contractFailureDiagnostics(
	rec SuiteRecord,
	goldByEntity map[llmguard.EntityType][]spanKey,
	predictedByEntity map[llmguard.EntityType][]spanKey,
	recScope map[llmguard.EntityType]struct{},
) []ContractFailureDiagnostic {
	labelBySpan := make(map[spanKey]string)
	for _, ann := range rec.Annotations {
		if ann.Disposition != DispositionSupported {
			continue
		}
		entity := llmguard.EntityType(ann.MappedEntity)
		if !entityInScope(entity, recScope) {
			continue
		}
		key := spanKey{entity: entity, start: ann.Start, end: ann.End}
		labelBySpan[key] = ann.SourceLabel
	}

	var failures []ContractFailureDiagnostic
	entities := make(map[llmguard.EntityType]struct{})
	for entity := range goldByEntity {
		if entityInScope(entity, recScope) {
			entities[entity] = struct{}{}
		}
	}
	for entity := range predictedByEntity {
		if entityInScope(entity, recScope) {
			entities[entity] = struct{}{}
		}
	}
	for entity := range entities {
		goldSpans := goldByEntity[entity]
		predicted := predictedByEntity[entity]
		goldSet := make(map[spanKey]struct{}, len(goldSpans))
		for _, g := range goldSpans {
			goldSet[g] = struct{}{}
		}
		predSet := make(map[spanKey]struct{}, len(predicted))
		for _, p := range predicted {
			predSet[p] = struct{}{}
		}
		for span := range predSet {
			if _, ok := goldSet[span]; ok {
				continue
			}
			ps, pe := span.start, span.end
			failures = append(failures, ContractFailureDiagnostic{
				SourceRecordID: rec.SourceRecordID,
				SourceLabel:    labelBySpan[span],
				Entity:         string(entity),
				Kind:           "fp",
				PredictedStart: &ps,
				PredictedEnd:   &pe,
				InputSHA256:    rec.InputSHA256,
			})
		}
		for span := range goldSet {
			if _, ok := predSet[span]; ok {
				continue
			}
			gs, ge := span.start, span.end
			failures = append(failures, ContractFailureDiagnostic{
				SourceRecordID: rec.SourceRecordID,
				SourceLabel:    labelBySpan[span],
				Entity:         string(entity),
				Kind:           "fn",
				GoldStart:      &gs,
				GoldEnd:        &ge,
				InputSHA256:    rec.InputSHA256,
			})
		}
	}
	sort.Slice(failures, func(i, j int) bool {
		if failures[i].SourceRecordID != failures[j].SourceRecordID {
			return failures[i].SourceRecordID < failures[j].SourceRecordID
		}
		if failures[i].Entity != failures[j].Entity {
			return failures[i].Entity < failures[j].Entity
		}
		return failures[i].Kind < failures[j].Kind
	})
	return failures
}
