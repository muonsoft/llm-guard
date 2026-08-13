package detect_test

import (
	"strings"
	"testing"

	"github.com/muonsoft/llm-guard/internal/detect"
)

func TestValidateEmailMailbox_WhenAsciiForms_ExpectAccepted(t *testing.T) {
	t.Parallel()

	if !detect.ValidateEmailMailbox("a@b.co") {
		t.Fatal("expected valid mailbox")
	}
	if !detect.ValidateEmailMailbox(strings.Repeat("a", 50) + "@b.co") {
		t.Fatal("expected long local part valid")
	}
	if !detect.ValidateEmailMailbox("user@example.company") {
		t.Fatal("expected company suffix valid")
	}
}
