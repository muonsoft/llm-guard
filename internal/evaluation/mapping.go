package evaluation

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/muonsoft/llm-guard"
)

const mappingSchemaVersion = 1

// PersonComposition configures PERSON mapping from source components.
type PersonComposition struct {
	Components    []string `json:"components"`
	MinComponents int      `json:"min_components"`
}

// AddressComposition configures ADDRESS mapping requirements.
type AddressComposition struct {
	Required []string `json:"required"`
}

// MappingPolicy is a versioned source-label to product-entity mapping.
type MappingPolicy struct {
	SchemaVersion     int                 `json:"schema_version"`
	Version           string              `json:"version"`
	SourceID          string              `json:"source_id"`
	Direct            map[string]string   `json:"direct"`
	Person            *PersonComposition  `json:"person,omitempty"`
	Address           *AddressComposition `json:"address,omitempty"`
	UnmappedSensitive []string            `json:"unmapped_sensitive,omitempty"`
	IgnoreReasons     []string            `json:"ignore_reasons,omitempty"`
}

// LoadMappingPolicy reads and validates a mapping policy JSON file.
func LoadMappingPolicy(path string) (MappingPolicy, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return MappingPolicy{}, err
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return MappingPolicy{}, fmt.Errorf("mapping: %w", err)
	}
	for key := range raw {
		if !isAllowedMappingKey(key) {
			return MappingPolicy{}, fmt.Errorf("mapping: unknown field %q", key)
		}
	}
	dec := json.NewDecoder(strings.NewReader(string(data)))
	dec.DisallowUnknownFields()
	var policy MappingPolicy
	if err := dec.Decode(&policy); err != nil {
		return MappingPolicy{}, fmt.Errorf("mapping: %w", err)
	}
	if err := validateMappingPolicy(policy); err != nil {
		return MappingPolicy{}, err
	}
	return policy, nil
}

func isAllowedMappingKey(key string) bool {
	switch key {
	case "schema_version", "version", "source_id", "direct", "person", "address",
		"unmapped_sensitive", "ignore_reasons":
		return true
	default:
		return false
	}
}

func validateMappingPolicy(m MappingPolicy) error {
	if m.SchemaVersion != mappingSchemaVersion {
		return fmt.Errorf("mapping: unsupported schema_version %d", m.SchemaVersion)
	}
	if strings.TrimSpace(m.Version) == "" {
		return fmt.Errorf("mapping: version is empty")
	}
	if strings.TrimSpace(m.SourceID) == "" {
		return fmt.Errorf("mapping: source_id is empty")
	}
	if m.Direct == nil {
		return fmt.Errorf("mapping: direct is required")
	}
	for label, entity := range m.Direct {
		if strings.TrimSpace(label) == "" {
			return fmt.Errorf("mapping: direct contains empty source label")
		}
		if err := validateProductEntity(entity, "mapping direct"); err != nil {
			return err
		}
	}
	if m.Person != nil {
		if len(m.Person.Components) == 0 {
			return fmt.Errorf("mapping: person.components is empty")
		}
		for _, c := range m.Person.Components {
			if strings.TrimSpace(c) == "" {
				return fmt.Errorf("mapping: person.components contains empty value")
			}
		}
		if m.Person.MinComponents < 1 {
			return fmt.Errorf("mapping: person.min_components must be >= 1")
		}
	}
	if m.Address != nil {
		if len(m.Address.Required) == 0 {
			return fmt.Errorf("mapping: address.required is empty")
		}
		for _, c := range m.Address.Required {
			if strings.TrimSpace(c) == "" {
				return fmt.Errorf("mapping: address.required contains empty value")
			}
		}
	}
	return nil
}

func validateProductEntity(entity, context string) error {
	if entity == "" {
		return nil
	}
	if _, ok := mvpEntitySet[llmguard.EntityType(entity)]; !ok {
		return fmt.Errorf("%s: unknown product entity %q", context, entity)
	}
	return nil
}
