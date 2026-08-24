package evaluation

import (
	"strings"
	"testing"

	"github.com/muonsoft/llm-guard"
)

func contactEmailFindings() []llmguard.Finding {
	return []llmguard.Finding{
		{Entity: llmguard.EntityEmail, Start: 8, End: 14},
	}
}

func TestInferPlaceholderMap_WhenAlignedMask_ExpectTokenValue(t *testing.T) {
	t.Parallel()
	input := "Contact a@b.co"
	masked := "Contact {{LLMG_0123456789abcdef0123456789abcdef_0001}}"
	inferred, err := inferPlaceholderMap(input, masked, contactEmailFindings())
	if err != nil {
		t.Fatalf("inferPlaceholderMap: %v", err)
	}
	token := "{{LLMG_0123456789abcdef0123456789abcdef_0001}}"
	if inferred[token] != "a@b.co" {
		t.Fatalf("inferred value = %q, want email placeholder value", inferred[token])
	}
}

func TestInferPlaceholderMap_WhenCyrillicFIOAndEmail_ExpectSpanAlignedValues(t *testing.T) {
	t.Parallel()
	input := "Иван Петров a@b.co"
	ns := strings.Repeat("a", 32)
	t1 := "{{LLMG_" + ns + "_0001}}"
	t2 := "{{LLMG_" + ns + "_0002}}"
	masked := t1 + " " + t2
	findings := []llmguard.Finding{
		{Entity: llmguard.EntityPerson, Start: 0, End: 21},
		{Entity: llmguard.EntityEmail, Start: 22, End: 28},
	}
	inferred, err := inferPlaceholderMap(input, masked, findings)
	if err != nil {
		t.Fatalf("inferPlaceholderMap: %v", err)
	}
	if inferred[t1] != "Иван Петров" {
		t.Fatalf("T1 = %q, want full person span", inferred[t1])
	}
	if inferred[t2] != "a@b.co" {
		t.Fatalf("T2 = %q, want email span", inferred[t2])
	}
}

func TestInferPlaceholderMap_WhenFindingsReverseSliceOrder_ExpectStartOrderMapping(t *testing.T) {
	t.Parallel()
	input := "Иван Петров a@b.co"
	ns := strings.Repeat("b", 32)
	t1 := "{{LLMG_" + ns + "_0001}}"
	t2 := "{{LLMG_" + ns + "_0002}}"
	masked := t1 + " " + t2
	findings := []llmguard.Finding{
		{Entity: llmguard.EntityEmail, Start: 22, End: 28},
		{Entity: llmguard.EntityPerson, Start: 0, End: 21},
	}
	inferred, err := inferPlaceholderMap(input, masked, findings)
	if err != nil {
		t.Fatalf("inferPlaceholderMap: %v", err)
	}
	if inferred[t1] != "Иван Петров" {
		t.Fatalf("T1 = %q, want person span by start order", inferred[t1])
	}
	if inferred[t2] != "a@b.co" {
		t.Fatalf("T2 = %q, want email span by start order", inferred[t2])
	}
}

func TestExpectedRecipeRestore_WhenIdentityMasked_ExpectOriginalInput(t *testing.T) {
	t.Parallel()
	input := "Contact a@b.co"
	masked := "Contact {{LLMG_0123456789abcdef0123456789abcdef_0001}}"
	inferred, err := inferPlaceholderMap(input, masked, contactEmailFindings())
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
	inferred, err := inferPlaceholderMap(input, masked, contactEmailFindings())
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
	inferred, err := inferPlaceholderMap(input, masked, contactEmailFindings())
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
