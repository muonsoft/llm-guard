package evaluation

import (
	"strings"
	"testing"
)

func TestValidateThresholdSet_WhenContractEmptyEntityBounds_ExpectRejection(t *testing.T) {
	t.Parallel()
	set := ThresholdSet{
		SchemaVersion: 1,
		ID:            "bad",
		Profiles: map[string]ProfileThreshold{
			"contract": {
				Status:   "gate",
				Entities: map[string]NumericBounds{"EMAIL": {}},
			},
		},
	}
	err := validateThresholdSet(set)
	if err == nil {
		t.Fatal("expected contract entities EMAIL empty bounds to be rejected")
	}
	if !containsAll(err.Error(), "contract", "entities[EMAIL]", "no bound fields") {
		t.Fatalf("error = %q", err.Error())
	}
}

func TestValidateThresholdSet_WhenContractEmptySourceBounds_ExpectRejection(t *testing.T) {
	t.Parallel()
	set := ThresholdSet{
		SchemaVersion: 1,
		ID:            "bad",
		Profiles: map[string]ProfileThreshold{
			"contract": {
				Status:  "gate",
				Sources: map[string]NumericBounds{"src": {}},
			},
		},
	}
	err := validateThresholdSet(set)
	if err == nil {
		t.Fatal("expected contract sources src empty bounds to be rejected")
	}
	if !containsAll(err.Error(), "contract", "sources[src]", "no bound fields") {
		t.Fatalf("error = %q", err.Error())
	}
}

func TestValidateThresholdSet_WhenContractMinByteCoverage_ExpectRejection(t *testing.T) {
	t.Parallel()
	cov := 0.5
	set := ThresholdSet{
		SchemaVersion: 1,
		ID:            "bad",
		Profiles: map[string]ProfileThreshold{
			"contract": {Status: "gate", MinByteCoverage: &cov},
		},
	}
	err := validateThresholdSet(set)
	if err == nil {
		t.Fatal("expected contract min_byte_coverage to be rejected")
	}
	if !containsAll(err.Error(), "contract", "min_byte_coverage") {
		t.Fatalf("error = %q", err.Error())
	}
}

func TestValidateThresholdSet_WhenExposureSourceMinRecall_ExpectRejection(t *testing.T) {
	t.Parallel()
	recall := 0.9
	set := ThresholdSet{
		SchemaVersion: 1,
		ID:            "bad",
		Profiles: map[string]ProfileThreshold{
			"exposure": {
				Status: "gate",
				Sources: map[string]NumericBounds{
					"src-a": {MinRecall: &recall},
				},
			},
		},
	}
	err := validateThresholdSet(set)
	if err == nil {
		t.Fatal("expected exposure source min_recall to be rejected")
	}
	if !containsAll(err.Error(), "exposure", "sources", "min_recall") {
		t.Fatalf("error = %q", err.Error())
	}
}

func TestValidateThresholdSet_WhenLifecycleMinPrecision_ExpectRejection(t *testing.T) {
	t.Parallel()
	precision := 0.9
	set := ThresholdSet{
		SchemaVersion: 1,
		ID:            "bad",
		Profiles: map[string]ProfileThreshold{
			"lifecycle": {Status: "gate", MinPrecision: &precision},
		},
	}
	err := validateThresholdSet(set)
	if err == nil {
		t.Fatal("expected lifecycle min_precision to be rejected")
	}
	if !containsAll(err.Error(), "lifecycle", "min_precision") {
		t.Fatalf("error = %q", err.Error())
	}
}

func TestValidateThresholdSet_WhenLifecycleEntities_ExpectRejection(t *testing.T) {
	t.Parallel()
	set := ThresholdSet{
		SchemaVersion: 1,
		ID:            "bad",
		Profiles: map[string]ProfileThreshold{
			"lifecycle": {
				Status:   "gate",
				Entities: map[string]NumericBounds{"EMAIL": {}},
			},
		},
	}
	err := validateThresholdSet(set)
	if err == nil {
		t.Fatal("expected lifecycle entities to be rejected")
	}
	if !containsAll(err.Error(), "lifecycle", "entities") {
		t.Fatalf("error = %q", err.Error())
	}
}

func TestValidateThresholdSet_WhenExposureEntities_ExpectRejection(t *testing.T) {
	t.Parallel()
	set := ThresholdSet{
		SchemaVersion: 1,
		ID:            "bad",
		Profiles: map[string]ProfileThreshold{
			"exposure": {
				Status:   "gate",
				Entities: map[string]NumericBounds{"EMAIL": {}},
			},
		},
	}
	err := validateThresholdSet(set)
	if err == nil {
		t.Fatal("expected exposure entities to be rejected")
	}
	if !containsAll(err.Error(), "exposure", "entities") {
		t.Fatalf("error = %q", err.Error())
	}
}

func containsAll(s string, parts ...string) bool {
	for _, part := range parts {
		if !strings.Contains(s, part) {
			return false
		}
	}
	return true
}
