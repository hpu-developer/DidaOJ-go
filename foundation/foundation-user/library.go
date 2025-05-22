package foundationuser

import "unicode"

func IsValidUsername(username string) bool {
	if len(username) < 3 || len(username) > 20 {
		return false
	}
	for _, ch := range username {
		if !(unicode.IsLetter(ch) || unicode.IsUpper(ch) || unicode.IsNumber(ch) || ch == '_') {
			return false
		}
	}
	return true
}
