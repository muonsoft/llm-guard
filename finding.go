package llmguard

// Finding describes a detected entity span without storing the matched value.
// Start and End are UTF-8 byte offsets into the input text passed to Detect.
type Finding struct {
	Entity     EntityType
	Start      int
	End        int
	Confidence float64
	Detector   string
}
