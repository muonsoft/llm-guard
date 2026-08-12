package nlp_test

import (
	"testing"

	"github.com/muonsoft/llm-guard/internal/nlp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNameForms_WhenMandatoryDeclensions_ExpectRoles(t *testing.T) {
	t.Parallel()

	cases := []struct {
		word string
		role nlp.Role
	}{
		{"Ивану", nlp.RoleFirst},
		{"Сергеевичу", nlp.RolePatronymic},
		{"Петрову", nlp.RoleSurname},
		{"Иваном", nlp.RoleFirst},
		{"Сергеевичем", nlp.RolePatronymic},
		{"Петровым", nlp.RoleSurname},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.word, func(t *testing.T) {
			t.Parallel()
			tokens := nlp.Tokenize(tc.word)
			require.Len(t, tokens, 1)
			assert.Equal(t, tc.role, tokens[0].Role)
		})
	}
}
