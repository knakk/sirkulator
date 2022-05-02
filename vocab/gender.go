package vocab

import "golang.org/x/text/language"

type Gender string

const (
	GenderUnknown Gender = "?"
	GenderMale    Gender = "m"
	GenderFemale  Gender = "f"
	GenderOther   Gender = "o"
)

func ParseGender(s string) Gender {
	switch s {
	case "m", "male", "mann":
		return GenderMale
	case "f", "female", "kvinne":
		return GenderFemale
	case "o":
		return GenderOther
	default:
		return GenderUnknown
	}
}

func (g Gender) Label(tag language.Tag) string {
	return "Gender.Label: TODO"
}
