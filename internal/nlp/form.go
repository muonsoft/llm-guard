package nlp

// Role identifies a lexical name component.
type Role uint8

const (
	RoleNone Role = iota
	RoleFirst
	RolePatronymic
	RoleSurname
)

// FormClass is a bounded compatible-form bitset for declined Russian names.
type FormClass uint16

const (
	FormNomMasc FormClass = 1 << iota
	FormDatMasc
	FormInstMasc
	FormNomFem
	FormDatFem
	FormInstFem
)

func (f FormClass) compatible(other FormClass) bool {
	if f == 0 || other == 0 {
		return true
	}
	return f&other != 0
}

func mergeForms(parts ...FormClass) FormClass {
	if len(parts) == 0 {
		return 0
	}
	merged := parts[0]
	for _, part := range parts[1:] {
		if merged == 0 || part == 0 {
			continue
		}
		merged &= part
		if merged == 0 {
			return 0
		}
	}
	return merged
}
