package evaluation

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

const manifestSchemaVersion = 1

var allowedDistributionModes = map[string]struct{}{
	"manifest-only":     {},
	"cache-only":        {},
	"committed-derived": {},
}

// SourceManifest pins an external evaluation source.
type SourceManifest struct {
	SchemaVersion  int    `json:"schema_version"`
	ID             string `json:"id"`
	CanonicalURL   string `json:"canonical_url"`
	Revision       string `json:"revision"`
	DigestSHA256   string `json:"digest_sha256"`
	License        string `json:"license"`
	Attribution    string `json:"attribution"`
	Distribution   string `json:"distribution"`
	AdapterVersion string `json:"adapter_version"`
	SnapshotDate   string `json:"snapshot_date,omitempty"`
}

// LoadSourceManifest reads and validates a source manifest JSON file.
func LoadSourceManifest(path string) (SourceManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return SourceManifest{}, err
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return SourceManifest{}, fmt.Errorf("manifest: %w", err)
	}
	for key := range raw {
		if !isAllowedManifestKey(key) {
			return SourceManifest{}, fmt.Errorf("manifest: unknown field %q", key)
		}
	}
	dec := json.NewDecoder(strings.NewReader(string(data)))
	dec.DisallowUnknownFields()
	var manifest SourceManifest
	if err := dec.Decode(&manifest); err != nil {
		return SourceManifest{}, fmt.Errorf("manifest: %w", err)
	}
	if err := validateSourceManifest(manifest); err != nil {
		return SourceManifest{}, err
	}
	return manifest, nil
}

func isAllowedManifestKey(key string) bool {
	switch key {
	case "schema_version", "id", "canonical_url", "revision", "digest_sha256",
		"license", "attribution", "distribution", "adapter_version", "snapshot_date":
		return true
	default:
		return false
	}
}

func validateSourceManifest(m SourceManifest) error {
	if m.SchemaVersion != manifestSchemaVersion {
		return fmt.Errorf("manifest: unsupported schema_version %d", m.SchemaVersion)
	}
	if strings.TrimSpace(m.ID) == "" {
		return fmt.Errorf("manifest: id is empty")
	}
	if strings.TrimSpace(m.CanonicalURL) == "" {
		return fmt.Errorf("manifest: canonical_url is empty")
	}
	if strings.TrimSpace(m.Revision) == "" {
		return fmt.Errorf("manifest: revision is empty")
	}
	if strings.TrimSpace(m.DigestSHA256) == "" {
		return fmt.Errorf("manifest: digest_sha256 is empty")
	}
	if strings.TrimSpace(m.License) == "" {
		return fmt.Errorf("manifest: license is empty")
	}
	if strings.TrimSpace(m.Attribution) == "" {
		return fmt.Errorf("manifest: attribution is empty")
	}
	if strings.TrimSpace(m.Distribution) == "" {
		return fmt.Errorf("manifest: distribution is empty")
	}
	if _, ok := allowedDistributionModes[m.Distribution]; !ok {
		return fmt.Errorf("manifest: unknown distribution %q", m.Distribution)
	}
	if strings.TrimSpace(m.AdapterVersion) == "" {
		return fmt.Errorf("manifest: adapter_version is empty")
	}
	return nil
}
