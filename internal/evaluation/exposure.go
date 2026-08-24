package evaluation

import (
	"context"
	"sort"

	"github.com/muonsoft/llm-guard"
)

type labelExposureKey struct {
	sourceLabel  string
	mappedEntity string
	disposition  string
}

type ignoredKey struct {
	sourceLabel string
	reason      string
}

// EvaluateExposure runs the exposure profile on a normalized suite.
func EvaluateExposure(ctx context.Context, guard *llmguard.Guard, suite Suite) (ExposureReport, error) {
	byLabel := make(map[labelExposureKey]*LabelExposureMetrics)
	ignored := make(map[ignoredKey]int)
	var summary ExposureSummary

	for _, rec := range suite.Records {
		resolved, err := DetectResolve(ctx, guard, rec.Input)
		if err != nil {
			return ExposureReport{}, err
		}
		caseMetrics := exposureCaseMetrics(rec, resolved)
		summary.SensitiveBytes += caseMetrics.sensitiveBytes
		summary.CoveredSensitiveBytes += caseMetrics.coveredSensitive
		summary.LeakedSensitiveBytes += caseMetrics.leakedSensitive
		summary.OvermatchedBytes += caseMetrics.overmatched

		for key, m := range caseMetrics.byLabel {
			agg := byLabel[key]
			if agg == nil {
				agg = &LabelExposureMetrics{
					SourceLabel:  key.sourceLabel,
					MappedEntity: key.mappedEntity,
					Disposition:  key.disposition,
				}
				byLabel[key] = agg
			}
			agg.SpanCount += m.spanCount
			agg.FullyCoveredSpanCount += m.fullyCovered
			agg.SensitiveBytes += m.sensitiveBytes
			agg.CoveredSensitiveBytes += m.coveredSensitive
			agg.LeakedSensitiveBytes += m.leakedSensitive
			agg.OvermatchedBytes += m.overmatched
		}
		for key, count := range caseMetrics.ignored {
			ignored[key] += count
		}
	}

	summary.ByteCoverage = ratio(summary.CoveredSensitiveBytes, summary.SensitiveBytes)

	report := ExposureReport{
		Profile:        string(ProfileExposure),
		SuiteID:        suite.SuiteID,
		MappingVersion: suite.MappingVersion,
		SourceIDs:      append([]string(nil), suite.SourceIDs...),
		Cases:          len(suite.Records),
		Summary:        summary,
		Status:         StatusDiagnostic,
	}
	report.ByLabel = sortedLabelMetrics(byLabel)
	report.Ignored = sortedIgnoredCounts(ignored)
	return report, nil
}

type perLabelCaseMetrics struct {
	spanCount        int
	fullyCovered     int
	sensitiveBytes   int
	coveredSensitive int
	leakedSensitive  int
	overmatched      int
}

type caseExposureResult struct {
	sensitiveBytes   int
	coveredSensitive int
	leakedSensitive  int
	overmatched      int
	byLabel          map[labelExposureKey]perLabelCaseMetrics
	ignored          map[ignoredKey]int
}

func exposureCaseMetrics(rec SuiteRecord, resolved []llmguard.Finding) caseExposureResult {
	byLabel := make(map[labelExposureKey]perLabelCaseMetrics)
	ignored := make(map[ignoredKey]int)
	labelIntervals := make(map[labelExposureKey][]byteInterval)

	for _, ann := range rec.Annotations {
		switch ann.Disposition {
		case DispositionIgnored:
			ignored[ignoredKey{sourceLabel: ann.SourceLabel, reason: ann.Reason}]++
			continue
		case DispositionSupported, DispositionUnsupported:
			key := labelExposureKey{
				sourceLabel:  ann.SourceLabel,
				mappedEntity: ann.MappedEntity,
				disposition:  ann.Disposition,
			}
			iv := byteInterval{start: ann.Start, end: ann.End}
			labelIntervals[key] = append(labelIntervals[key], iv)
		}
	}

	var goldIntervals []byteInterval
	for key, intervals := range labelIntervals {
		merged := mergeIntervals(intervals)
		goldIntervals = append(goldIntervals, merged...)
		m := byLabel[key]
		m.spanCount = len(intervals)
		m.sensitiveBytes = unionByteCount(intervals)
		byLabel[key] = m
	}

	var predIntervals []byteInterval
	type predFinding struct {
		entity string
		iv     byteInterval
	}
	var predFindings []predFinding
	for _, f := range resolved {
		iv := byteInterval{start: f.Start, end: f.End}
		predIntervals = append(predIntervals, iv)
		predFindings = append(predFindings, predFinding{
			entity: string(f.Entity),
			iv:     iv,
		})
	}
	mergedGold := mergeIntervals(goldIntervals)
	mergedPred := mergeIntervals(predIntervals)

	sensitiveBytes := unionByteCount(goldIntervals)
	coveredSensitive := intersectionByteCount(goldIntervals, predIntervals)
	leakedSensitive := differenceByteCount(goldIntervals, predIntervals)
	overmatched := differenceByteCount(predIntervals, goldIntervals)

	for key, intervals := range labelIntervals {
		m := byLabel[key]
		mergedLabel := mergeIntervals(intervals)
		m.coveredSensitive = intersectionByteCount(mergedLabel, predIntervals)
		m.leakedSensitive = differenceByteCount(mergedLabel, predIntervals)
		for _, iv := range intervals {
			if spanFullyCovered(iv, mergedPred) {
				m.fullyCovered++
			}
		}
		byLabel[key] = m
	}

	for _, pf := range predFindings {
		extra := intervalDifference([]byteInterval{pf.iv}, mergedGold)
		if unionByteCount(extra) == 0 {
			continue
		}
		key := labelExposureKey{
			sourceLabel:  "",
			mappedEntity: pf.entity,
			disposition:  "",
		}
		m := byLabel[key]
		m.overmatched += unionByteCount(extra)
		byLabel[key] = m
	}

	_ = mergedGold

	return caseExposureResult{
		sensitiveBytes:   sensitiveBytes,
		coveredSensitive: coveredSensitive,
		leakedSensitive:  leakedSensitive,
		overmatched:      overmatched,
		byLabel:          byLabel,
		ignored:          ignored,
	}
}

func sortedLabelMetrics(byLabel map[labelExposureKey]*LabelExposureMetrics) []LabelExposureMetrics {
	keys := make([]labelExposureKey, 0, len(byLabel))
	for k := range byLabel {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].sourceLabel != keys[j].sourceLabel {
			return keys[i].sourceLabel < keys[j].sourceLabel
		}
		if keys[i].mappedEntity != keys[j].mappedEntity {
			return keys[i].mappedEntity < keys[j].mappedEntity
		}
		return keys[i].disposition < keys[j].disposition
	})
	out := make([]LabelExposureMetrics, 0, len(keys))
	for _, k := range keys {
		out = append(out, *byLabel[k])
	}
	return out
}

func sortedIgnoredCounts(ignored map[ignoredKey]int) []IgnoredCount {
	keys := make([]ignoredKey, 0, len(ignored))
	for k := range ignored {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].sourceLabel != keys[j].sourceLabel {
			return keys[i].sourceLabel < keys[j].sourceLabel
		}
		return keys[i].reason < keys[j].reason
	})
	out := make([]IgnoredCount, 0, len(keys))
	for _, k := range keys {
		out = append(out, IgnoredCount{
			SourceLabel: k.sourceLabel,
			Reason:      k.reason,
			Count:       ignored[k],
		})
	}
	return out
}

// ExposureCaseMetricsForTest exposes per-case exposure metrics for unit tests.
func ExposureCaseMetricsForTest(rec SuiteRecord, resolved []llmguard.Finding) (
	sensitive, covered, leaked, overmatched int,
	byLabel map[string]PerLabelCaseMetricsForTest,
) {
	result := exposureCaseMetrics(rec, resolved)
	out := make(map[string]PerLabelCaseMetricsForTest, len(result.byLabel))
	for k, m := range result.byLabel {
		out[labelMetricsKey(k)] = PerLabelCaseMetricsForTest{
			SpanCount:        m.spanCount,
			FullyCovered:     m.fullyCovered,
			SensitiveBytes:   m.sensitiveBytes,
			CoveredSensitive: m.coveredSensitive,
			LeakedSensitive:  m.leakedSensitive,
			Overmatched:      m.overmatched,
		}
	}
	return result.sensitiveBytes, result.coveredSensitive, result.leakedSensitive, result.overmatched, out
}

func labelMetricsKey(k labelExposureKey) string {
	return k.sourceLabel + "\x00" + k.mappedEntity + "\x00" + k.disposition
}

// PerLabelCaseMetricsForTest exposes per-label case metrics for tests.
type PerLabelCaseMetricsForTest struct {
	SpanCount        int
	FullyCovered     int
	SensitiveBytes   int
	CoveredSensitive int
	LeakedSensitive  int
	Overmatched      int
}

// LabelMetricsLookupKey builds a lookup key for ExposureCaseMetricsForTest results.
func LabelMetricsLookupKey(sourceLabel, mappedEntity, disposition string) string {
	return sourceLabel + "\x00" + mappedEntity + "\x00" + disposition
}

// ByteIntervalForTest is a test helper interval.
type ByteIntervalForTest struct {
	Start int
	End   int
}

// ExposureByteMetricsForTest computes exposure byte metrics from interval slices.
func ExposureByteMetricsForTest(gold, pred []ByteIntervalForTest) (sensitive, covered, leaked, overmatched int) {
	toIV := func(in []ByteIntervalForTest) []byteInterval {
		out := make([]byteInterval, len(in))
		for i, iv := range in {
			out[i] = byteInterval{start: iv.Start, end: iv.End}
		}
		return out
	}
	g := toIV(gold)
	p := toIV(pred)
	return unionByteCount(g), intersectionByteCount(g, p), differenceByteCount(g, p), differenceByteCount(p, g)
}

// MergeIntervalsForTest exposes interval merging for tests.
func MergeIntervalsForTest(intervals []ByteIntervalForTest) int {
	ivs := make([]byteInterval, len(intervals))
	for i, iv := range intervals {
		ivs[i] = byteInterval{start: iv.Start, end: iv.End}
	}
	return unionByteCount(ivs)
}
