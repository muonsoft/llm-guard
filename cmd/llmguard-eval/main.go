package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/muonsoft/llm-guard/internal/evaluation"
)

func main() {
	corpus := flag.String("corpus", "", "path to versioned JSONL evaluation corpus")
	format := flag.String("format", "markdown", "report format: markdown or json")
	failOnRegression := flag.Bool("fail-on-regression", false, "exit non-zero when coverage is incomplete or FP/FN > 0")
	flag.Parse()

	if strings.TrimSpace(*corpus) == "" {
		fmt.Fprintln(os.Stderr, "missing required -corpus flag")
		os.Exit(2)
	}

	cases, err := evaluation.LoadCases(*corpus)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load corpus: %v\n", err)
		os.Exit(1)
	}

	guard, err := evaluation.NewMVPGuard()
	if err != nil {
		fmt.Fprintf(os.Stderr, "build guard: %v\n", err)
		os.Exit(1)
	}

	report, err := evaluation.Evaluate(context.Background(), guard, cases)
	if err != nil {
		fmt.Fprintf(os.Stderr, "evaluate: %v\n", err)
		os.Exit(1)
	}
	report.CorpusPath = *corpus

	switch strings.ToLower(strings.TrimSpace(*format)) {
	case "markdown", "md":
		fmt.Print(evaluation.FormatMarkdown(report))
	case "json":
		out, err := evaluation.FormatJSON(report)
		if err != nil {
			fmt.Fprintf(os.Stderr, "format json: %v\n", err)
			os.Exit(1)
		}
		fmt.Print(string(out))
	default:
		fmt.Fprintf(os.Stderr, "unsupported format %q\n", *format)
		os.Exit(2)
	}

	if *failOnRegression && report.HasRegression() {
		os.Exit(1)
	}
}
