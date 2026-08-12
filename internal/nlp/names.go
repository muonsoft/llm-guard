package nlp

import "strings"

type nameEntry struct {
	role  Role
	forms FormClass
}

var exactNameForms = map[string]nameEntry{
	"иван":        {RoleFirst, FormNomMasc},
	"ивану":       {RoleFirst, FormDatMasc},
	"иваном":      {RoleFirst, FormInstMasc},
	"мария":       {RoleFirst, FormNomFem},
	"марии":       {RoleFirst, FormDatFem | FormNomFem},
	"марией":      {RoleFirst, FormInstFem},
	"анна":        {RoleFirst, FormNomFem},
	"анне":        {RoleFirst, FormDatFem},
	"анной":       {RoleFirst, FormInstFem},
	"сергей":      {RoleFirst, FormNomMasc},
	"сергею":      {RoleFirst, FormDatMasc},
	"сергеем":     {RoleFirst, FormInstMasc},
	"сергеевич":   {RolePatronymic, FormNomMasc},
	"сергеевичу":  {RolePatronymic, FormDatMasc},
	"сергеевичем": {RolePatronymic, FormInstMasc},
	"сергеевна":   {RolePatronymic, FormNomFem},
	"сергеевне":   {RolePatronymic, FormDatFem},
	"сергеевной":  {RolePatronymic, FormInstFem},
	"петров":      {RoleSurname, FormNomMasc},
	"петрову":     {RoleSurname, FormDatMasc},
	"петровым":    {RoleSurname, FormInstMasc},
	"петрова":     {RoleSurname, FormNomFem},
	"петровой":    {RoleSurname, FormInstFem},
	"иванова":     {RoleSurname, FormNomFem},
	"иванову":     {RoleSurname, FormDatFem},
	"ивановой":    {RoleSurname, FormInstFem},
	"сидоров":     {RoleSurname, FormNomMasc},
	"сидорову":    {RoleSurname, FormDatMasc},
	"сидоровым":   {RoleSurname, FormInstMasc},
	"смирнов":     {RoleSurname, FormNomMasc},
	"смирнову":    {RoleSurname, FormDatMasc},
	"смирновым":   {RoleSurname, FormInstMasc},
	"смирнова":    {RoleSurname, FormNomFem},
}

func lookupNameRole(folded string) (Role, FormClass, bool) {
	if entry, ok := exactNameForms[folded]; ok {
		return entry.role, entry.forms, true
	}
	return classifyBySuffix(folded)
}

func classifyBySuffix(folded string) (Role, FormClass, bool) {
	if role, forms, ok := classifyPatronymicSuffix(folded); ok {
		return role, forms, true
	}
	if role, forms, ok := classifySurnameSuffix(folded); ok {
		return role, forms, true
	}
	return RoleNone, 0, false
}

func classifyPatronymicSuffix(folded string) (Role, FormClass, bool) {
	switch {
	case hasSuffix(folded, "ович", "евич", "ич"):
		return RolePatronymic, FormNomMasc, true
	case hasSuffix(folded, "овичу", "евичу", "ичу"):
		return RolePatronymic, FormDatMasc, true
	case hasSuffix(folded, "овичем", "евичем", "ичем"):
		return RolePatronymic, FormInstMasc, true
	case hasSuffix(folded, "овна", "евна", "ична"):
		return RolePatronymic, FormNomFem, true
	case hasSuffix(folded, "овне", "евне", "ичне"):
		return RolePatronymic, FormDatFem, true
	case hasSuffix(folded, "овной", "евной", "ичной"):
		return RolePatronymic, FormInstFem, true
	default:
		return RoleNone, 0, false
	}
}

func classifySurnameSuffix(folded string) (Role, FormClass, bool) {
	if len([]rune(folded)) < 4 {
		return RoleNone, 0, false
	}
	switch {
	case hasSuffix(folded, "ов", "ев", "ин", "ын"):
		return RoleSurname, FormNomMasc, true
	case hasSuffix(folded, "ова", "ева", "ина", "ына"):
		return RoleSurname, FormNomFem, true
	case hasSuffix(folded, "ову", "еву", "ину", "ыну"):
		return RoleSurname, FormDatMasc, true
	case hasSuffix(folded, "овой", "евой", "иной", "ыной"):
		return RoleSurname, FormDatFem | FormInstFem, true
	case hasSuffix(folded, "овым", "евым", "иным"):
		return RoleSurname, FormInstMasc, true
	default:
		return RoleNone, 0, false
	}
}

func hasSuffix(word string, suffixes ...string) bool {
	for _, suffix := range suffixes {
		if strings.HasSuffix(word, suffix) && len(word) > len(suffix)+1 {
			return true
		}
	}
	return false
}
