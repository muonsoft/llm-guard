package evaluation

import "github.com/muonsoft/llm-guard"

// SuiteSchemaVersion is the supported normalized suite schema version.
const SuiteSchemaVersion = 2

// Disposition values for suite v2 annotations.
const (
	DispositionSupported   = "supported"
	DispositionUnsupported = "unsupported"
	DispositionIgnored     = "ignored"
)

// SuiteRecord is one normalized evaluation case (schema v2).
type SuiteRecord struct {
	SchemaVersion    int               `json:"schema_version"`
	SuiteID          string            `json:"suite_id"`
	SourceID         string            `json:"source_id"`
	SourceRecordID   string            `json:"source_record_id"`
	MappingVersion   string            `json:"mapping_version"`
	Input            string            `json:"input"`
	InputSHA256      string            `json:"input_sha256"`
	Annotations      []SuiteAnnotation `json:"annotations"`
	DeclaredEntities []string          `json:"declared_entities,omitempty"`
	Lifecycle        *SuiteLifecycle   `json:"lifecycle,omitempty"`
}

// SuiteAnnotation is one source label span in a normalized case.
type SuiteAnnotation struct {
	SourceLabel  string `json:"source_label"`
	MappedEntity string `json:"mapped_entity"`
	Start        int    `json:"start"`
	End          int    `json:"end"`
	Disposition  string `json:"disposition"`
	Reason       string `json:"reason"`
}

// SuiteLifecycle holds optional lifecycle expectations for a case.
type SuiteLifecycle struct {
	ExpectedAction string `json:"expected_action"`
	ResponseRecipe string `json:"response_recipe"`
}

// Suite is a loaded normalized evaluation suite.
type Suite struct {
	Records        []SuiteRecord
	SuiteID        string
	MappingVersion string
	SourceIDs      []string
	Scope          []llmguard.EntityType
}
