package evaluation

// ApplyContractThresholds updates contract report status from threshold set.
func ApplyContractThresholds(report *ContractReport, set ThresholdSet) {
	profile, ok := set.ProfileThresholdFor(string(ProfileContract))
	if !ok {
		report.Status = StatusDiagnostic
		return
	}
	report.ThresholdID = set.ID
	if profile.Status == "diagnostic" {
		report.Status = StatusDiagnostic
		return
	}
	if report.HasContractRegression() {
		report.Status = StatusFail
		return
	}
	if violatesContractBounds(report, profile) {
		report.Status = StatusFail
		return
	}
	report.Status = StatusPass
}

// ApplyExposureThresholds updates exposure report status from threshold set.
func ApplyExposureThresholds(report *ExposureReport, set ThresholdSet) {
	profile, ok := set.ProfileThresholdFor(string(ProfileExposure))
	if !ok {
		report.Status = StatusDiagnostic
		return
	}
	report.ThresholdID = set.ID
	if profile.Status == "diagnostic" {
		report.Status = StatusDiagnostic
		return
	}
	if violatesExposureBounds(report, profile) {
		report.Status = StatusFail
		return
	}
	report.Status = StatusPass
}

func violatesContractBounds(report *ContractReport, profile ProfileThreshold) bool {
	if profile.MaxFP != nil && float64(report.Summary.FP) > *profile.MaxFP {
		return true
	}
	if profile.MaxFN != nil && float64(report.Summary.FN) > *profile.MaxFN {
		return true
	}
	for _, row := range report.Entities {
		if bounds, ok := profile.Entities[string(row.Entity)]; ok {
			if violatesEntityContractBounds(row, bounds) {
				return true
			}
		}
	}
	return false
}

func violatesEntityContractBounds(row EntityMetrics, bounds NumericBounds) bool {
	if bounds.MaxFP != nil && float64(row.FP) > *bounds.MaxFP {
		return true
	}
	if bounds.MaxFN != nil && float64(row.FN) > *bounds.MaxFN {
		return true
	}
	if bounds.MinPrecision != nil && row.Precision < *bounds.MinPrecision {
		return true
	}
	if bounds.MinRecall != nil && row.Recall < *bounds.MinRecall {
		return true
	}
	return false
}

func violatesExposureBounds(report *ExposureReport, profile ProfileThreshold) bool {
	if profile.MaxLeakedSensitiveBytes != nil && float64(report.Summary.LeakedSensitiveBytes) > *profile.MaxLeakedSensitiveBytes {
		return true
	}
	if profile.MinByteCoverage != nil && report.Summary.ByteCoverage < *profile.MinByteCoverage {
		return true
	}
	return false
}

// ApplyLifecycleThresholds updates lifecycle report status from threshold set.
func ApplyLifecycleThresholds(report *LifecycleReport, set ThresholdSet) {
	profile, ok := set.ProfileThresholdFor(string(ProfileLifecycle))
	if !ok {
		report.Status = StatusDiagnostic
		return
	}
	report.ThresholdID = set.ID
	if profile.Status == "diagnostic" {
		report.Status = StatusDiagnostic
		return
	}
	if report.HasLifecycleRegression() {
		report.Status = StatusFail
		return
	}
	if violatesLifecycleBounds(report, profile) {
		report.Status = StatusFail
		return
	}
	report.Status = StatusPass
}

func violatesLifecycleBounds(report *LifecycleReport, profile ProfileThreshold) bool {
	if profile.MaxFN != nil && float64(len(report.Diagnostics)) > *profile.MaxFN {
		return true
	}
	if profile.MaxFP != nil && float64(report.Errors) > *profile.MaxFP {
		return true
	}
	return false
}

// LifecycleFailsGate returns true when lifecycle thresholds with gate status are violated.
func LifecycleFailsGate(report LifecycleReport, set ThresholdSet) bool {
	profile, ok := set.ProfileThresholdFor(string(ProfileLifecycle))
	if !ok || profile.Status != "gate" {
		return false
	}
	return report.HasLifecycleRegression() || violatesLifecycleBounds(&report, profile)
}

// ExposureFailsGate returns true when exposure thresholds with gate status are violated.
func ExposureFailsGate(report ExposureReport, set ThresholdSet) bool {
	profile, ok := set.ProfileThresholdFor(string(ProfileExposure))
	if !ok || profile.Status != "gate" {
		return false
	}
	return violatesExposureBounds(&report, profile)
}

// FormatSafeMarker is used in tests to verify formatters never leak input substrings.
const FormatSafeMarker = "PEB_SAFE_MARKER_XYZ"
