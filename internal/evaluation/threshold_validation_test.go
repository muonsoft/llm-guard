package evaluation

import (
	"math"
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

func TestValidateThresholdSet_WhenContractMinPrecisionNegative_ExpectRejection(t *testing.T) {
	t.Parallel()
	v := -1.0
	set := ThresholdSet{
		SchemaVersion: 1,
		ID:            "bad",
		Profiles: map[string]ProfileThreshold{
			"contract": {Status: "gate", MinPrecision: &v},
		},
	}
	err := validateThresholdSet(set)
	if err == nil {
		t.Fatal("expected contract min_precision -1 to be rejected")
	}
	if !containsAll(err.Error(), `profile "contract"`, "min_precision", "[0,1]") {
		t.Fatalf("error = %q", err.Error())
	}
}

func TestValidateThresholdSet_WhenContractMinRecallAboveOne_ExpectRejection(t *testing.T) {
	t.Parallel()
	v := 1.1
	set := ThresholdSet{
		SchemaVersion: 1,
		ID:            "bad",
		Profiles: map[string]ProfileThreshold{
			"contract": {Status: "gate", MinRecall: &v},
		},
	}
	err := validateThresholdSet(set)
	if err == nil {
		t.Fatal("expected contract min_recall 1.1 to be rejected")
	}
	if !containsAll(err.Error(), `profile "contract"`, "min_recall", "[0,1]") {
		t.Fatalf("error = %q", err.Error())
	}
}

func TestValidateThresholdSet_WhenContractMaxFPNegative_ExpectRejection(t *testing.T) {
	t.Parallel()
	v := -1.0
	set := ThresholdSet{
		SchemaVersion: 1,
		ID:            "bad",
		Profiles: map[string]ProfileThreshold{
			"contract": {Status: "gate", MaxFP: &v},
		},
	}
	err := validateThresholdSet(set)
	if err == nil {
		t.Fatal("expected contract max_fp -1 to be rejected")
	}
	if !containsAll(err.Error(), `profile "contract"`, "max_fp", ">= 0") {
		t.Fatalf("error = %q", err.Error())
	}
}

func TestValidateThresholdSet_WhenExposureMinByteCoverageNegative_ExpectRejection(t *testing.T) {
	t.Parallel()
	v := -0.1
	set := ThresholdSet{
		SchemaVersion: 1,
		ID:            "bad",
		Profiles: map[string]ProfileThreshold{
			"exposure": {Status: "gate", MinByteCoverage: &v},
		},
	}
	err := validateThresholdSet(set)
	if err == nil {
		t.Fatal("expected exposure min_byte_coverage -0.1 to be rejected")
	}
	if !containsAll(err.Error(), `profile "exposure"`, "min_byte_coverage", "[0,1]") {
		t.Fatalf("error = %q", err.Error())
	}
}

func TestValidateThresholdSet_WhenExposureMaxLeakedSensitiveBytesNegative_ExpectRejection(t *testing.T) {
	t.Parallel()
	v := -1.0
	set := ThresholdSet{
		SchemaVersion: 1,
		ID:            "bad",
		Profiles: map[string]ProfileThreshold{
			"exposure": {Status: "gate", MaxLeakedSensitiveBytes: &v},
		},
	}
	err := validateThresholdSet(set)
	if err == nil {
		t.Fatal("expected exposure max_leaked_sensitive_bytes -1 to be rejected")
	}
	if !containsAll(err.Error(), `profile "exposure"`, "max_leaked_sensitive_bytes", ">= 0") {
		t.Fatalf("error = %q", err.Error())
	}
}

func TestValidateThresholdSet_WhenContractEntityMinPrecisionNegative_ExpectRejection(t *testing.T) {
	t.Parallel()
	v := -1.0
	set := ThresholdSet{
		SchemaVersion: 1,
		ID:            "bad",
		Profiles: map[string]ProfileThreshold{
			"contract": {
				Status:   "gate",
				Entities: map[string]NumericBounds{"EMAIL": {MinPrecision: &v}},
			},
		},
	}
	err := validateThresholdSet(set)
	if err == nil {
		t.Fatal("expected nested contract entities EMAIL min_precision -1 to be rejected")
	}
	if !containsAll(err.Error(), `profile "contract"`, "entities[EMAIL]", "min_precision", "[0,1]") {
		t.Fatalf("error = %q", err.Error())
	}
}

func TestValidateProfileThreshold_WhenContractMinPrecisionInf_ExpectRejection(t *testing.T) {
	t.Parallel()
	v := math.Inf(1)
	profile := ProfileThreshold{Status: "gate", MinPrecision: &v}
	err := validateProfileThreshold("contract", profile)
	if err == nil {
		t.Fatal("expected contract min_precision +Inf to be rejected")
	}
	if !containsAll(err.Error(), `profile "contract"`, "min_precision", "[0,1]") {
		t.Fatalf("error = %q", err.Error())
	}
}

func TestValidateThresholdSet_WhenContractMinPrecisionZeroAndOne_ExpectAccept(t *testing.T) {
	t.Parallel()
	zero := 0.0
	one := 1.0
	for _, v := range []*float64{&zero, &one} {
		set := ThresholdSet{
			SchemaVersion: 1,
			ID:            "ok",
			Profiles: map[string]ProfileThreshold{
				"contract": {Status: "gate", MinPrecision: v},
			},
		}
		if err := validateThresholdSet(set); err != nil {
			t.Fatalf("min_precision %v: %v", *v, err)
		}
	}
}

func TestValidateThresholdSet_WhenContractMaxFPZero_ExpectAccept(t *testing.T) {
	t.Parallel()
	v := 0.0
	set := ThresholdSet{
		SchemaVersion: 1,
		ID:            "ok",
		Profiles: map[string]ProfileThreshold{
			"contract": {Status: "gate", MaxFP: &v},
		},
	}
	if err := validateThresholdSet(set); err != nil {
		t.Fatalf("max_fp 0: %v", err)
	}
}
