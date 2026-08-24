package evaluation

import (
	"io"

	"github.com/muonsoft/llm-guard"
)

type repeatingReader struct {
	buf []byte
	off int
}

func newRepeatingReader(seed []byte) *repeatingReader {
	if len(seed) == 0 {
		seed = []byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x01}
	}
	return &repeatingReader{buf: append([]byte(nil), seed...)}
}

func (r *repeatingReader) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = r.buf[r.off%len(r.buf)]
		r.off++
	}
	return len(p), nil
}

// NewDeterministicGuard builds MVP guard with a repeating entropy reader for tests.
func NewDeterministicGuard(source io.Reader) (*llmguard.Guard, error) {
	if source == nil {
		source = newRepeatingReader(nil)
	}
	return llmguard.New(
		llmguard.WithDetector(llmguard.NewPersonDetector()),
		llmguard.WithDetector(llmguard.NewAddressDetector()),
		llmguard.WithDetector(llmguard.NewEmailDetector()),
		llmguard.WithDetector(llmguard.NewPhoneDetector()),
		llmguard.WithDetector(llmguard.NewIPDetector()),
		llmguard.WithDetector(llmguard.NewURLDetector()),
		llmguard.WithDetector(llmguard.NewINNDetector()),
		llmguard.WithDetector(llmguard.NewSNILSDetector()),
		llmguard.WithDetector(llmguard.NewPassportDetector()),
		llmguard.WithDetector(llmguard.NewBankCardDetector()),
		llmguard.WithDetector(llmguard.NewBankAccountDetector()),
		llmguard.WithDetector(llmguard.NewDateOfBirthDetector()),
		llmguard.WithDetector(llmguard.NewJWTDetector()),
		llmguard.WithDetector(llmguard.NewPEMPrivateKeyDetector()),
		llmguard.WithDetector(llmguard.NewAPIKeyDetector()),
		llmguard.WithDetector(llmguard.NewDSNDetector()),
		llmguard.WithRandomSource(source),
	)
}
