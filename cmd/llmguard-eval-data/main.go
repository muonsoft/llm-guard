package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/muonsoft/llm-guard/internal/evaluation"
)

func main() {
	action := flag.String("action", "", "action: fetch or normalize (required)")
	manifestPath := flag.String("manifest", "", "path to source manifest JSON (required)")
	mappingPath := flag.String("mapping", "", "path to mapping policy JSON (required for normalize)")
	cachePath := flag.String("cache", "", "explicit cache directory (required)")
	flag.Parse()

	if strings.TrimSpace(*action) == "" {
		fmt.Fprintln(os.Stderr, "missing required -action flag")
		os.Exit(2)
	}
	if strings.TrimSpace(*manifestPath) == "" {
		fmt.Fprintln(os.Stderr, "missing required -manifest flag")
		os.Exit(2)
	}
	if strings.TrimSpace(*cachePath) == "" {
		fmt.Fprintln(os.Stderr, "missing required -cache flag")
		os.Exit(2)
	}

	manifest, err := evaluation.LoadSourceManifest(*manifestPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load manifest: %v\n", err)
		os.Exit(1)
	}

	switch strings.ToLower(strings.TrimSpace(*action)) {
	case "fetch":
		if err := evaluation.FetchSource(manifest, *cachePath, nil); err != nil {
			fmt.Fprintf(os.Stderr, "fetch: %v\n", err)
			os.Exit(1)
		}
	case "normalize":
		if strings.TrimSpace(*mappingPath) == "" {
			fmt.Fprintln(os.Stderr, "missing required -mapping flag for normalize")
			os.Exit(2)
		}
		policy, err := evaluation.LoadMappingPolicy(*mappingPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "load mapping: %v\n", err)
			os.Exit(1)
		}
		outPath := evaluation.DefaultNormalizedPath(*cachePath, manifest)
		if err := evaluation.NormalizeSource(manifest, policy, *cachePath, outPath); err != nil {
			fmt.Fprintf(os.Stderr, "normalize: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "unsupported action %q\n", *action)
		os.Exit(2)
	}
}
