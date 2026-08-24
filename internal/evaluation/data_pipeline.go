package evaluation

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const fetchHTTPTimeout = 30 * time.Second

// FetchSource downloads and verifies a source artifact into cache.
func FetchSource(manifest SourceManifest, cacheDir string, client *http.Client) error {
	if client == nil {
		client = &http.Client{Timeout: fetchHTTPTimeout}
	}
	switch manifest.ID {
	case redMadRobotSourceID:
		return fetchRedMadRobot(manifest, cacheDir, client)
	case factRuEvalSourceID:
		return fetchFactRuEval(manifest, cacheDir)
	case "gitleaks-reviewed-fixtures":
		return verifyGitleaksFixtures(manifest)
	default:
		return fmt.Errorf("unsupported source %q for fetch", manifest.ID)
	}
}

func fetchRedMadRobot(manifest SourceManifest, cacheDir string, client *http.Client) error {
	destDir := filepath.Join(cacheDir, "sources", manifest.ID)
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return err
	}
	dest := filepath.Join(destDir, "test.csv")
	url := strings.TrimRight(manifest.CanonicalURL, "/") + "/resolve/" + manifest.Revision + "/test.csv"
	tmp := dest + ".tmp"
	if err := downloadVerifyAtomic(client, url, tmp, dest, manifest.DigestSHA256); err != nil {
		return err
	}
	return nil
}

func fetchFactRuEval(manifest SourceManifest, cacheDir string) error {
	dest := filepath.Join(cacheDir, "sources", manifest.ID)
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}
	if _, err := os.Stat(filepath.Join(dest, ".git")); err != nil {
		return fmt.Errorf("factrueval cache missing; clone %s at %s into %s", manifest.CanonicalURL, manifest.Revision, dest)
	}
	return verifyFactRuEvalDigest(dest, manifest.DigestSHA256)
}

func verifyGitleaksFixtures(manifest SourceManifest) error {
	canonical := []struct {
		rel   string
		local string
	}{
		{"cmd/generate/config/rules/privatekey.go", "privatekey.go"},
		{"testdata/repos/small/README.md", "README.md"},
		{"testdata/repos/small/api/api.go", "api.go"},
		{"testdata/repos/small/api/ignoreCommit.go", "ignoreCommit.go"},
		{"testdata/repos/small/api/ignoreGlobal.go", "ignoreGlobal.go"},
		{"testdata/repos/small/main.go", "main.go"},
	}
	root := filepath.Join("testdata", "evaluation", "generated", "gitleaks")
	var paths []string
	open := func(rel string) (io.ReadCloser, error) {
		for _, item := range canonical {
			if item.rel == rel {
				return os.Open(filepath.Join(root, item.local))
			}
		}
		return nil, fmt.Errorf("unknown path %q", rel)
	}
	for _, item := range canonical {
		paths = append(paths, item.rel)
	}
	return verifyCanonicalDigest(paths, manifest.DigestSHA256, open)
}

func downloadVerifyAtomic(client *http.Client, url, tmpPath, destPath, expectedDigest string) error {
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}
	if err := os.MkdirAll(filepath.Dir(tmpPath), 0o755); err != nil {
		return err
	}
	f, err := os.Create(tmpPath)
	if err != nil {
		return err
	}
	if _, err := io.Copy(f, resp.Body); err != nil {
		f.Close()
		os.Remove(tmpPath)
		return err
	}
	if err := f.Close(); err != nil {
		os.Remove(tmpPath)
		return err
	}
	if err := VerifyFileDigest(tmpPath, expectedDigest); err != nil {
		os.Remove(tmpPath)
		return err
	}
	if err := os.Rename(tmpPath, destPath); err != nil {
		os.Remove(tmpPath)
		return err
	}
	return nil
}

// NormalizeSource reads verified cache input and writes normalized JSONL.
func NormalizeSource(manifest SourceManifest, policy MappingPolicy, cacheDir, outPath string) error {
	if err := verifySourceBeforeNormalize(manifest, cacheDir); err != nil {
		return err
	}
	if policy.SourceID != manifest.ID && policy.SourceID != "" {
		return fmt.Errorf("mapping source_id %q does not match manifest id %q", policy.SourceID, manifest.ID)
	}
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return err
	}
	tmp := outPath + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	writer := bufio.NewWriter(f)
	emit := func(rec SuiteRecord) error {
		rec.SuiteID = manifest.ID + "-normalized"
		line, err := json.Marshal(rec)
		if err != nil {
			return err
		}
		if _, err := writer.Write(append(line, '\n')); err != nil {
			return err
		}
		return nil
	}
	var normErr error
	switch manifest.ID {
	case redMadRobotSourceID:
		csvPath := filepath.Join(cacheDir, "sources", manifest.ID, "test.csv")
		in, err := os.Open(csvPath)
		if err != nil {
			normErr = err
			break
		}
		normErr = NormalizeRedMadRobotCSV(in, policy, emit)
		in.Close()
	case factRuEvalSourceID:
		root := filepath.Join(cacheDir, "sources", manifest.ID)
		normErr = NormalizeFactRuEvalTree(root, policy, emit)
	case "gitleaks-reviewed-fixtures":
		normErr = NormalizeGitleaksFixtures(policy, emit)
	default:
		normErr = fmt.Errorf("unsupported source %q for normalize", manifest.ID)
	}
	if err := writer.Flush(); err != nil && normErr == nil {
		normErr = err
	}
	if err := f.Close(); err != nil && normErr == nil {
		normErr = err
	}
	if normErr != nil {
		os.Remove(tmp)
		return normErr
	}
	return os.Rename(tmp, outPath)
}

func verifySourceBeforeNormalize(manifest SourceManifest, cacheDir string) error {
	switch manifest.ID {
	case redMadRobotSourceID:
		csvPath := filepath.Join(cacheDir, "sources", manifest.ID, "test.csv")
		return VerifyFileDigest(csvPath, manifest.DigestSHA256)
	case factRuEvalSourceID:
		root := filepath.Join(cacheDir, "sources", manifest.ID)
		return verifyFactRuEvalDigest(root, manifest.DigestSHA256)
	case "gitleaks-reviewed-fixtures":
		return verifyGitleaksFixtures(manifest)
	default:
		return fmt.Errorf("unsupported source %q", manifest.ID)
	}
}

func verifyFactRuEvalDigest(root, expected string) error {
	var paths []string
	for _, split := range []string{"devset", "testset"} {
		splitRoot := filepath.Join(root, split)
		entries, err := os.ReadDir(splitRoot)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			name := entry.Name()
			if strings.HasSuffix(name, ".txt") || strings.HasSuffix(name, ".tokens") ||
				strings.HasSuffix(name, ".spans") || strings.HasSuffix(name, ".objects") {
				if split == "testset" && name == "list.txt" {
					continue
				}
				paths = append(paths, filepath.Join(split, name))
			}
		}
	}
	return verifyCanonicalDigest(paths, expected, func(rel string) (io.ReadCloser, error) {
		return os.Open(filepath.Join(root, rel))
	})
}

func verifyCanonicalDigest(paths []string, expected string, open func(string) (io.ReadCloser, error)) error {
	actual, err := canonicalPathDigest(paths, open)
	if err != nil {
		return err
	}
	if actual != expected {
		return fmt.Errorf("canonical digest mismatch")
	}
	return nil
}

// DefaultNormalizedPath returns cache normalized JSONL path for a manifest.
func DefaultNormalizedPath(cacheDir string, manifest SourceManifest) string {
	return filepath.Join(cacheDir, "normalized", manifest.ID+".jsonl")
}
