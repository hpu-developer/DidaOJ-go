package foundationoj

import "strings"

func GetOriginOjKey(oj string) string {
	oj = strings.ToLower(oj)
	switch oj {
	case "didaoj":
		return "didaoj"
	case "hdu":
		return "HDU"
	case "poj":
		return "POJ"
	case "nyoj":
		return "NYOJ"
	default:
		return ""
	}
}
