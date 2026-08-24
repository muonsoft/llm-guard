package evaluation

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"sort"
)

func canonicalPathDigest(paths []string, open func(string) (io.ReadCloser, error)) (string, error) {
	sorted := append([]string(nil), paths...)
	sort.Strings(sorted)
	h := sha256.New()
	for _, path := range sorted {
		rc, err := open(path)
		if err != nil {
			return "", fmt.Errorf("digest open %q: %w", path, err)
		}
		data, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return "", err
		}
		h.Write([]byte(path))
		h.Write([]byte{0})
		h.Write(data)
		h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// VerifyFileDigest compares one file against an expected SHA-256 hex digest.
func VerifyFileDigest(path string, expected string) error {
	rc, err := openFile(path)
	if err != nil {
		return err
	}
	defer rc.Close()
	data, err := io.ReadAll(rc)
	if err != nil {
		return err
	}
	sum := sha256.Sum256(data)
	actual := hex.EncodeToString(sum[:])
	if actual != expected {
		return fmt.Errorf("digest mismatch for %q", path)
	}
	return nil
}

func openFile(path string) (io.ReadCloser, error) {
	return openOSFile(path)
}
