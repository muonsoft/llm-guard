package evaluation

import (
	"strings"
	"testing"
)

func TestInferPlaceholderMap_WhenAlignedMask_ExpectTokenValue(t *testing.T) {
	t.Parallel()
	input := "Contact a@b.co"
	masked := "Contact {{LLMG_0123456789abcdef0123456789abcdef_0001}}"
	inferred, err := inferPlaceholderMap(input, masked)
	if err != nil {
		t.Fatalf("inferPlaceholderMap: %v", err)
	}
	token := "{{LLMG_0123456789abcdef0123456789abcdef_0001}}"
	if inferred[token] != "a@b.co" {
		t.Fatalf("inferred value = %q, want email placeholder value", inferred[token])
	}
}

func TestExpectedRecipeRestore_WhenIdentityMasked_ExpectOriginalInput(t *testing.T) {
	t.Parallel()
	input := "Contact a@b.co"
	masked := "Contact {{LLMG_0123456789abcdef0123456789abcdef_0001}}"
	inferred, err := inferPlaceholderMap(input, masked)
	if err != nil {
		t.Fatalf("inferPlaceholderMap: %v", err)
	}
	if got := expectedRecipeRestore(masked, inferred); got != input {
		t.Fatalf("expected restore = %q, want original input", got)
	}
}

func TestExpectedRecipeRestore_WhenMutateRecipe_ExpectMutatedTokenNotWrapped(t *testing.T) {
	t.Parallel()
	input := "Contact a@b.co"
	masked := "Contact {{LLMG_0123456789abcdef0123456789abcdef_0001}}"
	inferred, err := inferPlaceholderMap(input, masked)
	if err != nil {
		t.Fatalf("inferPlaceholderMap: %v", err)
	}
	recipe := applyResponseRecipe(masked, "mutate_placeholder")
	expected := expectedRecipeRestore(recipe, inferred)
	if strings.HasPrefix(expected, "junk ") || strings.HasSuffix(expected, " junk") {
		t.Fatalf("expected must not be junk-wrapped, got %q", expected)
	}
	if !placeholderPattern.MatchString(expected) {
		t.Fatal("mutate recipe expected restore must retain mutated placeholder token")
	}
	if expected == input {
		t.Fatal("mutate recipe expected restore must not equal original input")
	}
}

func TestExpectedRecipeRestore_WhenJunkWrapped_ExpectNotEqualToExpected(t *testing.T) {
	t.Parallel()
	input := "Contact a@b.co"
	masked := "Contact {{LLMG_0123456789abcdef0123456789abcdef_0001}}"
	inferred, err := inferPlaceholderMap(input, masked)
	if err != nil {
		t.Fatalf("inferPlaceholderMap: %v", err)
	}
	expected := expectedRecipeRestore(masked, inferred)
	junkWrapped := "junk " + expected + " junk"
	if junkWrapped == expected {
		t.Fatal("junk wrapper must change the string")
	}
}

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
