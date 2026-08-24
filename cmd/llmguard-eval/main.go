package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/muonsoft/llm-guard"
	"github.com/muonsoft/llm-guard/internal/evaluation"
)

func main() {
	corpus := flag.String("corpus", "", "path to versioned JSONL evaluation corpus (schema v1)")
	suitePath := flag.String("suite", "", "path to normalized JSONL evaluation suite (schema v2)")
	profile := flag.String("profile", "", "evaluation profile: contract, exposure, or lifecycle")
	thresholdsPath := flag.String("thresholds", "", "path to optional threshold set JSON")
	format := flag.String("format", "markdown", "report format: markdown or json")
	failOnRegression := flag.Bool("fail-on-regression", false, "exit non-zero when coverage is incomplete or FP/FN > 0")
	flag.Parse()

	hasCorpus := strings.TrimSpace(*corpus) != ""
	hasSuite := strings.TrimSpace(*suitePath) != ""

	if hasCorpus && hasSuite {
		fmt.Fprintln(os.Stderr, "flags -corpus and -suite are mutually exclusive")
		os.Exit(2)
	}
	if !hasCorpus && !hasSuite {
		fmt.Fprintln(os.Stderr, "missing required -corpus or -suite flag")
		os.Exit(2)
	}
	if hasSuite && strings.TrimSpace(*profile) == "" {
		fmt.Fprintln(os.Stderr, "missing required -profile flag with -suite")
		os.Exit(2)
	}
	if hasCorpus {
		if strings.TrimSpace(*profile) != "" {
			fmt.Fprintln(os.Stderr, "flag -profile is only valid with -suite")
			os.Exit(2)
		}
		if strings.TrimSpace(*thresholdsPath) != "" {
			fmt.Fprintln(os.Stderr, "flag -thresholds is only valid with -suite")
			os.Exit(2)
		}
	}

	guard, err := evaluation.NewMVPGuard()
	if err != nil {
		fmt.Fprintf(os.Stderr, "build guard: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()
	reportFormat := strings.ToLower(strings.TrimSpace(*format))

	if hasCorpus {
		os.Exit(runV1(ctx, guard, *corpus, reportFormat, *failOnRegression))
	}
	os.Exit(runV2(ctx, guard, *suitePath, strings.ToLower(strings.TrimSpace(*profile)), *thresholdsPath, reportFormat, *failOnRegression))
}

func runV1(ctx context.Context, guard *llmguard.Guard, corpusPath, reportFormat string, failOnRegression bool) int {
	cases, err := evaluation.LoadCases(corpusPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load corpus: %v\n", err)
		return 1
	}

	report, err := evaluation.Evaluate(ctx, guard, cases)
	if err != nil {
		fmt.Fprintf(os.Stderr, "evaluate: %v\n", err)
		return 1
	}
	report.CorpusPath = corpusPath

	if code, err := writeV1Report(report, reportFormat); code != 0 {
		return code
	} else if err != nil {
		fmt.Fprintf(os.Stderr, "format: %v\n", err)
		return 1
	}
	if failOnRegression && report.HasRegression() {
		return 1
	}
	return 0
}

func runV2(ctx context.Context, guard *llmguard.Guard, suitePath, profile, thresholdsPath, reportFormat string, failOnRegression bool) int {
	suite, err := evaluation.LoadSuite(suitePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load suite: %v\n", err)
		return 1
	}

	var thresholds evaluation.ThresholdSet
	if strings.TrimSpace(thresholdsPath) != "" {
		thresholds, err = evaluation.LoadThresholdSet(thresholdsPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "load thresholds: %v\n", err)
			return 1
		}
	}

	switch profile {
	case "contract":
		return runContract(ctx, guard, suite, thresholds, reportFormat, failOnRegression)
	case "exposure":
		return runExposure(ctx, guard, suite, thresholds, reportFormat, failOnRegression)
	case "lifecycle":
		return runLifecycle(ctx, guard, suite, thresholds, reportFormat, failOnRegression)
	default:
		fmt.Fprintf(os.Stderr, "unsupported profile %q\n", profile)
		return 2
	}
}

func runContract(ctx context.Context, guard *llmguard.Guard, suite evaluation.Suite, thresholds evaluation.ThresholdSet, reportFormat string, failOnRegression bool) int {
	report, err := evaluation.EvaluateContract(ctx, guard, suite)
	if err != nil {
		fmt.Fprintf(os.Stderr, "evaluate: %v\n", err)
		return 1
	}
	if thresholds.ID != "" {
		evaluation.ApplyContractThresholds(&report, thresholds)
	}
	if code, err := writeContractReport(report, reportFormat); code != 0 {
		return code
	} else if err != nil {
		fmt.Fprintf(os.Stderr, "format: %v\n", err)
		return 1
	}
	if failOnRegression && report.HasContractRegression() {
		return 1
	}
	if thresholds.ID != "" && report.Status == evaluation.StatusFail {
		return 1
	}
	return 0
}

func runExposure(ctx context.Context, guard *llmguard.Guard, suite evaluation.Suite, thresholds evaluation.ThresholdSet, reportFormat string, failOnRegression bool) int {
	report, err := evaluation.EvaluateExposure(ctx, guard, suite)
	if err != nil {
		fmt.Fprintf(os.Stderr, "evaluate: %v\n", err)
		return 1
	}
	if thresholds.ID != "" {
		evaluation.ApplyExposureThresholds(&report, thresholds)
	}
	if code, err := writeExposureReport(report, reportFormat); code != 0 {
		return code
	} else if err != nil {
		fmt.Fprintf(os.Stderr, "format: %v\n", err)
		return 1
	}
	if failOnRegression && thresholds.ID != "" && evaluation.ExposureFailsGate(report, thresholds) {
		return 1
	}
	if thresholds.ID != "" && report.Status == evaluation.StatusFail {
		return 1
	}
	return 0
}

func writeV1Report(report evaluation.Report, reportFormat string) (int, error) {
	switch reportFormat {
	case "markdown", "md":
		fmt.Print(evaluation.FormatMarkdown(report))
	case "json":
		out, err := evaluation.FormatJSON(report)
		if err != nil {
			return 0, err
		}
		fmt.Print(string(out))
	default:
		fmt.Fprintf(os.Stderr, "unsupported format %q\n", reportFormat)
		return 2, nil
	}
	return 0, nil
}

func writeContractReport(report evaluation.ContractReport, reportFormat string) (int, error) {
	switch reportFormat {
	case "markdown", "md":
		fmt.Print(evaluation.FormatContractMarkdown(report))
	case "json":
		out, err := evaluation.FormatContractJSON(report)
		if err != nil {
			return 0, err
		}
		fmt.Print(string(out))
	default:
		fmt.Fprintf(os.Stderr, "unsupported format %q\n", reportFormat)
		return 2, nil
	}
	return 0, nil
}

func writeExposureReport(report evaluation.ExposureReport, reportFormat string) (int, error) {
	switch reportFormat {
	case "markdown", "md":
		fmt.Print(evaluation.FormatExposureMarkdown(report))
	case "json":
		out, err := evaluation.FormatExposureJSON(report)
		if err != nil {
			return 0, err
		}
		fmt.Print(string(out))
	default:
		fmt.Fprintf(os.Stderr, "unsupported format %q\n", reportFormat)
		return 2, nil
	}
	return 0, nil
}

func runLifecycle(ctx context.Context, guard *llmguard.Guard, suite evaluation.Suite, thresholds evaluation.ThresholdSet, reportFormat string, failOnRegression bool) int {
	report, err := evaluation.EvaluateLifecycle(ctx, guard, suite)
	if err != nil {
		fmt.Fprintf(os.Stderr, "evaluate: %v\n", err)
		return 1
	}
	if thresholds.ID != "" {
		evaluation.ApplyLifecycleThresholds(&report, thresholds)
	}
	if code, err := writeLifecycleReport(report, reportFormat); code != 0 {
		return code
	} else if err != nil {
		fmt.Fprintf(os.Stderr, "format: %v\n", err)
		return 1
	}
	if failOnRegression && report.HasLifecycleRegression() {
		return 1
	}
	if thresholds.ID != "" && report.Status == evaluation.StatusFail {
		return 1
	}
	return 0
}

func writeLifecycleReport(report evaluation.LifecycleReport, reportFormat string) (int, error) {
	switch reportFormat {
	case "markdown", "md":
		fmt.Print(evaluation.FormatLifecycleMarkdown(report))
	case "json":
		out, err := evaluation.FormatLifecycleJSON(report)
		if err != nil {
			return 0, err
		}
		fmt.Print(string(out))
	default:
		fmt.Fprintf(os.Stderr, "unsupported format %q\n", reportFormat)
		return 2, nil
	}
	return 0, nil
}
