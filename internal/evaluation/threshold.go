package evaluation

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strings"
)

const thresholdSchemaVersion = 1

var allowedProfileNames = map[string]struct{}{
	"contract":  {},
	"exposure":  {},
	"lifecycle": {},
}

var allowedThresholdStatuses = map[string]struct{}{
	"gate":       {},
	"diagnostic": {},
}

var profileTopLevelBoundKeys = map[string]map[string]struct{}{
	"contract": {
		"max_fp":        {},
		"max_fn":        {},
		"min_precision": {},
		"min_recall":    {},
	},
	"exposure": {
		"min_byte_coverage":          {},
		"max_leaked_sensitive_bytes": {},
	},
	"lifecycle": {
		"max_fp": {},
		"max_fn": {},
	},
}

// ThresholdSet is a versioned profile threshold configuration.
type ThresholdSet struct {
	SchemaVersion int                         `json:"schema_version"`
	ID            string                      `json:"id"`
	Profiles      map[string]ProfileThreshold `json:"profiles"`
}

// ProfileThreshold defines gate/diagnostic bounds for one profile.
type ProfileThreshold struct {
	Status                  string                   `json:"status"`
	MaxFP                   *float64                 `json:"max_fp,omitempty"`
	MaxFN                   *float64                 `json:"max_fn,omitempty"`
	MinPrecision            *float64                 `json:"min_precision,omitempty"`
	MinRecall               *float64                 `json:"min_recall,omitempty"`
	MinByteCoverage         *float64                 `json:"min_byte_coverage,omitempty"`
	MaxLeakedSensitiveBytes *float64                 `json:"max_leaked_sensitive_bytes,omitempty"`
	Entities                map[string]NumericBounds `json:"entities,omitempty"`
	Sources                 map[string]NumericBounds `json:"sources,omitempty"`
}

// NumericBounds holds optional numeric thresholds.
type NumericBounds struct {
	MaxFP                   *float64 `json:"max_fp,omitempty"`
	MaxFN                   *float64 `json:"max_fn,omitempty"`
	MinPrecision            *float64 `json:"min_precision,omitempty"`
	MinRecall               *float64 `json:"min_recall,omitempty"`
	MinByteCoverage         *float64 `json:"min_byte_coverage,omitempty"`
	MaxLeakedSensitiveBytes *float64 `json:"max_leaked_sensitive_bytes,omitempty"`
}

func numericBoundsEmpty(b NumericBounds) bool {
	return b.MaxFP == nil && b.MaxFN == nil && b.MinPrecision == nil &&
		b.MinRecall == nil && b.MinByteCoverage == nil && b.MaxLeakedSensitiveBytes == nil
}

// LoadThresholdSet reads and validates a threshold JSON file.
func LoadThresholdSet(path string) (ThresholdSet, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return ThresholdSet{}, err
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return ThresholdSet{}, fmt.Errorf("thresholds: %w", err)
	}
	for key := range raw {
		if key != "schema_version" && key != "id" && key != "profiles" {
			return ThresholdSet{}, fmt.Errorf("thresholds: unknown field %q", key)
		}
	}
	dec := json.NewDecoder(strings.NewReader(string(data)))
	dec.DisallowUnknownFields()
	var set ThresholdSet
	if err := dec.Decode(&set); err != nil {
		return ThresholdSet{}, fmt.Errorf("thresholds: %w", err)
	}
	if err := validateThresholdSet(set); err != nil {
		return ThresholdSet{}, err
	}
	return set, nil
}

func validateThresholdSet(set ThresholdSet) error {
	if set.SchemaVersion != thresholdSchemaVersion {
		return fmt.Errorf("thresholds: unsupported schema_version %d", set.SchemaVersion)
	}
	if strings.TrimSpace(set.ID) == "" {
		return fmt.Errorf("thresholds: id is empty")
	}
	if set.Profiles == nil {
		return fmt.Errorf("thresholds: profiles is required")
	}
	for name, profile := range set.Profiles {
		if _, ok := allowedProfileNames[name]; !ok {
			return fmt.Errorf("thresholds: unknown profile %q", name)
		}
		if err := validateProfileThreshold(name, profile); err != nil {
			return err
		}
	}
	return nil
}

func validateProfileThreshold(name string, profile ProfileThreshold) error {
	if strings.TrimSpace(profile.Status) == "" {
		return fmt.Errorf("thresholds profile %q: status is empty", name)
	}
	if _, ok := allowedThresholdStatuses[profile.Status]; !ok {
		return fmt.Errorf("thresholds profile %q: unknown status %q", name, profile.Status)
	}
	if err := validateProfileTopLevelBounds(name, profile); err != nil {
		return err
	}
	if err := validateNumericBoundsKeys(profile, name); err != nil {
		return err
	}
	if name == "lifecycle" && len(profile.Sources) > 0 {
		return fmt.Errorf("thresholds profile %q: lifecycle sources bounds are not supported", name)
	}
	if name == "lifecycle" && len(profile.Entities) > 0 {
		return fmt.Errorf("thresholds profile %q: lifecycle entities bounds are not supported", name)
	}
	if name == "exposure" && len(profile.Entities) > 0 {
		return fmt.Errorf("thresholds profile %q: exposure entities bounds are not supported", name)
	}
	for entity, bounds := range profile.Entities {
		if strings.TrimSpace(entity) == "" {
			return fmt.Errorf("thresholds profile %q: entities contains empty key", name)
		}
		if err := validateNestedBounds(name, "entities", entity, bounds); err != nil {
			return err
		}
	}
	for source, bounds := range profile.Sources {
		if strings.TrimSpace(source) == "" {
			return fmt.Errorf("thresholds profile %q: sources contains empty key", name)
		}
		if err := validateNestedBounds(name, "sources", source, bounds); err != nil {
			return err
		}
	}
	return nil
}

func validateNumericBoundsKeys(profile ProfileThreshold, name string) error {
	allowed, ok := profileTopLevelBoundKeys[name]
	if !ok {
		return fmt.Errorf("thresholds: unknown profile %q", name)
	}
	raw, err := json.Marshal(profile)
	if err != nil {
		return err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return err
	}
	for key := range fields {
		switch key {
		case "status", "entities", "sources":
			continue
		}
		if _, ok := allowed[key]; !ok {
			return fmt.Errorf("thresholds profile %q: unknown field %q", name, key)
		}
	}
	return nil
}

func validateNestedBounds(profileName, section, key string, bounds NumericBounds) error {
	allowed, ok := profileTopLevelBoundKeys[profileName]
	if !ok {
		return fmt.Errorf("thresholds: unknown profile %q", profileName)
	}
	if numericBoundsEmpty(bounds) {
		return fmt.Errorf("thresholds profile %q %s[%s]: has no bound fields", profileName, section, key)
	}
	if profileName == "lifecycle" {
		return fmt.Errorf("thresholds profile %q %s[%q]: nested bounds are not supported", profileName, section, key)
	}
	if profileName == "exposure" && section == "entities" {
		return fmt.Errorf("thresholds profile %q %s[%q]: nested bounds are not supported", profileName, section, key)
	}
	path := fmt.Sprintf(" %s[%s]", section, key)
	if err := validateNumericBoundsValues(profileName, path, bounds); err != nil {
		return err
	}
	raw, err := json.Marshal(bounds)
	if err != nil {
		return err
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		return err
	}
	for k := range m {
		if _, ok := allowed[k]; !ok {
			return fmt.Errorf("thresholds profile %q %s[%q]: unknown field %q", profileName, section, key, k)
		}
	}
	return nil
}

func validateProfileTopLevelBounds(name string, profile ProfileThreshold) error {
	bounds := NumericBounds{
		MaxFP:                   profile.MaxFP,
		MaxFN:                   profile.MaxFN,
		MinPrecision:            profile.MinPrecision,
		MinRecall:               profile.MinRecall,
		MinByteCoverage:         profile.MinByteCoverage,
		MaxLeakedSensitiveBytes: profile.MaxLeakedSensitiveBytes,
	}
	return validateNumericBoundsValues(name, "", bounds)
}

func validateNumericBoundsValues(profileName, path string, bounds NumericBounds) error {
	if err := validateRateBound(profileName, path, "min_precision", bounds.MinPrecision); err != nil {
		return err
	}
	if err := validateRateBound(profileName, path, "min_recall", bounds.MinRecall); err != nil {
		return err
	}
	if err := validateRateBound(profileName, path, "min_byte_coverage", bounds.MinByteCoverage); err != nil {
		return err
	}
	if err := validateNonNegativeBound(profileName, path, "max_fp", bounds.MaxFP); err != nil {
		return err
	}
	if err := validateNonNegativeBound(profileName, path, "max_fn", bounds.MaxFN); err != nil {
		return err
	}
	if err := validateNonNegativeBound(profileName, path, "max_leaked_sensitive_bytes", bounds.MaxLeakedSensitiveBytes); err != nil {
		return err
	}
	return nil
}

func validateRateBound(profileName, path, field string, v *float64) error {
	if v == nil {
		return nil
	}
	if math.IsNaN(*v) || math.IsInf(*v, 0) || *v < 0 || *v > 1 {
		return fmt.Errorf(`thresholds profile %q%s: %s must be in [0,1]`, profileName, path, field)
	}
	return nil
}

func validateNonNegativeBound(profileName, path, field string, v *float64) error {
	if v == nil {
		return nil
	}
	if math.IsNaN(*v) || math.IsInf(*v, 0) || *v < 0 {
		return fmt.Errorf(`thresholds profile %q%s: %s must be >= 0`, profileName, path, field)
	}
	return nil
}

// ProfileStatus returns the configured status for a profile, defaulting to diagnostic.
func (set ThresholdSet) ProfileStatus(profile string) string {
	if p, ok := set.Profiles[profile]; ok && p.Status != "" {
		return p.Status
	}
	return "diagnostic"
}

// ProfileThresholdFor returns thresholds for a profile if present.
func (set ThresholdSet) ProfileThresholdFor(profile string) (ProfileThreshold, bool) {
	p, ok := set.Profiles[profile]
	return p, ok
}
