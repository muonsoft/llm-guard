// External consumer fixture for release-check.sh consumer mode.
// Imports only the public module path; no internal packages.
package main

import (
	"context"
	"fmt"
	"os"

	llmguard "github.com/muonsoft/llm-guard"
)

func main() {
	guard, err := llmguard.New(
		llmguard.WithDetector(llmguard.NewEmailDetector()),
		llmguard.WithDetector(llmguard.NewPersonDetector()),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "new guard: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()
	prompt := "Contact a@b.co; signed by Иван Петров."

	findings, err := guard.Detect(ctx, prompt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "detect: %v\n", err)
		os.Exit(1)
	}
	if len(findings) == 0 {
		fmt.Fprintf(os.Stderr, "detect: expected findings\n")
		os.Exit(1)
	}

	masked, err := guard.Mask(ctx, prompt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "mask: %v\n", err)
		os.Exit(1)
	}

	restored, err := guard.Restore(ctx, masked.Text, masked.Tokens)
	if err != nil {
		fmt.Fprintf(os.Stderr, "restore: %v\n", err)
		os.Exit(1)
	}

	if restored != prompt {
		fmt.Fprintf(os.Stderr, "round-trip mismatch\n")
		os.Exit(1)
	}

	fmt.Println("ok")
}
