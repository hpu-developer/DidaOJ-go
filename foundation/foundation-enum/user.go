package foundationenum

type UserGender int

const (
	UserGenderUnknown UserGender = 0
	UserGenderMale    UserGender = 1
	UserGenderFemale  UserGender = 2
)

func GetUserGender(gender string) UserGender {
	switch gender {
	case "male":
		return UserGenderMale
	case "female":
		return UserGenderFemale
	default:
		return UserGenderUnknown
	}
}
