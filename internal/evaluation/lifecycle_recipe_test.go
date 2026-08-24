package evaluation

import "testing"

func TestApplyResponseRecipe_WhenMutatePlaceholder_ExpectDigitChanged(t *testing.T) {
	t.Parallel()
	masked := "Contact {{LLMG_0123456789abcdef0123456789abcdef_0001}} please"
	out := applyResponseRecipe(masked, "mutate_placeholder")
	if out == masked {
		t.Fatal("expected mutated placeholder")
	}
	if out == masked[:len(masked)-1]+"}" {
		t.Fatal("must not mutate closing brace")
	}
	loc := placeholderPattern.FindStringIndex(out)
	if loc == nil {
		t.Fatal("expected placeholder to remain")
	}
	token := out[loc[0]:loc[1]]
	if token == "{{LLMG_0123456789abcdef0123456789abcdef_0001}}" {
		t.Fatalf("expected digit mutation in token %q", token)
	}
}

func TestApplyResponseRecipe_WhenDeletePlaceholder_ExpectSingleRemoval(t *testing.T) {
	t.Parallel()
	masked := "a {{LLMG_0123456789abcdef0123456789abcdef_0001}} b {{LLMG_0123456789abcdef0123456789abcdef_0002}}"
	out := applyResponseRecipe(masked, "delete_placeholder")
	if !placeholderPattern.MatchString(out) {
		t.Fatal("expected one placeholder to remain")
	}
	if len(placeholderPattern.FindAllStringIndex(out, -1)) != 1 {
		t.Fatalf("expected exactly one placeholder, got %q", out)
	}
}
