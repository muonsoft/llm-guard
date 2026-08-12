package llmguard

import (
	"context"
	"regexp"
	"sort"
)

const apiKeyDetectorName = "secret_api_key"

// Provider token shapes pinned to snapshot date 2026-08-12; see docs/secret-patterns.md.
var (
	githubClassicPATPattern = regexp.MustCompile(`ghp_[A-Za-z0-9]{36}`)
	githubFinePATPattern    = regexp.MustCompile(`github_pat_[A-Za-z0-9]{22}_[A-Za-z0-9]{59}`)
	gitlabPATPattern        = regexp.MustCompile(`glpat-[A-Za-z0-9\-_]{20}`)
	awsAccessKeyPattern     = regexp.MustCompile(`(?:AKIA|ASIA)[A-Z0-9]{16}`)
	openAISkPattern         = regexp.MustCompile(`sk-[A-Za-z0-9]{20,64}`)
	openAIProjSkPattern     = regexp.MustCompile(`sk-proj-[A-Za-z0-9\-_]{20,128}`)
)

type apiKeyShape struct {
	pattern    *regexp.Regexp
	innerExtra func(rune) bool
}

var apiKeyShapes = []apiKeyShape{
	{pattern: githubClassicPATPattern, innerExtra: apiKeyInnerBoundary},
	{pattern: githubFinePATPattern, innerExtra: apiKeyInnerBoundary},
	{pattern: gitlabPATPattern, innerExtra: apiKeyInnerBoundary},
	{pattern: awsAccessKeyPattern, innerExtra: nil},
	{pattern: openAIProjSkPattern, innerExtra: apiKeyInnerBoundary},
	{pattern: openAISkPattern, innerExtra: apiKeyInnerBoundary},
}

type apiKeyDetector struct{}

// NewAPIKeyDetector returns an immutable built-in SECRET_API_KEY detector.
func NewAPIKeyDetector() Detector {
	return apiKeyDetector{}
}

func (apiKeyDetector) Name() string {
	return apiKeyDetectorName
}

func (apiKeyDetector) Detect(ctx context.Context, text string) ([]Finding, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var findings []Finding
	seen := make(map[spanKey]struct{})

	for _, shape := range apiKeyShapes {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		matches := shape.pattern.FindAllStringIndex(text, -1)
		for _, loc := range matches {
			start, end := loc[0], loc[1]
			if !asciiTokenBoundaryOK(text, start, end, shape.innerExtra) {
				continue
			}
			key := spanKey{start: start, end: end}
			if _, exists := seen[key]; exists {
				continue
			}
			seen[key] = struct{}{}
			findings = append(findings, Finding{
				Entity:     EntitySecretAPIKey,
				Start:      start,
				End:        end,
				Confidence: 0.9,
				Detector:   apiKeyDetectorName,
			})
		}
	}

	if len(findings) == 0 {
		return nil, nil
	}

	sort.SliceStable(findings, func(i, j int) bool {
		a, b := findings[i], findings[j]
		if a.Start != b.Start {
			return a.Start < b.Start
		}
		if a.End != b.End {
			return a.End < b.End
		}
		if a.Entity != b.Entity {
			return a.Entity < b.Entity
		}
		return a.Detector < b.Detector
	})

	return findings, nil
}

type spanKey struct {
	start int
	end   int
}

func apiKeyInnerBoundary(r rune) bool {
	return r == '-' || r == '_'
}
