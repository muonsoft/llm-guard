package evaluation

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

func buildSuiteRecord(sourceID, recordID, mappingVersion, text string, annotations []SuiteAnnotation) SuiteRecord {
	sum := sha256.Sum256([]byte(text))
	return SuiteRecord{
		SchemaVersion:  SuiteSchemaVersion,
		SuiteID:        sourceID + "-normalized",
		SourceID:       sourceID,
		SourceRecordID: recordID,
		MappingVersion: mappingVersion,
		Input:          text,
		InputSHA256:    hex.EncodeToString(sum[:]),
		Annotations:    annotations,
	}
}

func ignoredRecord(sourceID, recordID, mappingVersion, text, reason string) SuiteRecord {
	rec := buildSuiteRecord(sourceID, recordID, mappingVersion, text, nil)
	if len(text) == 0 {
		return rec
	}
	rec.Annotations = []SuiteAnnotation{{
		SourceLabel: "_record",
		Start:       0,
		End:         len(text),
		Disposition: DispositionIgnored,
		Reason:      reason,
	}}
	return rec
}

func suiteRecordSHA256(text string) string {
	sum := sha256.Sum256([]byte(text))
	return hex.EncodeToString(sum[:])
}

func validateRecordInputSHA256(rec *SuiteRecord) error {
	expected := suiteRecordSHA256(rec.Input)
	if rec.InputSHA256 != expected {
		return fmt.Errorf("record %q: input_sha256 mismatch", rec.SourceRecordID)
	}
	return nil
}
