package foundationcontest

func GetContestProblemIndexStr(index int) string {
	result := ""
	for index > 0 {
		index--
		result = string(rune('A'+(index%26))) + result
		index = index / 26
	}
	return result
}

func GetContestProblemIndex(indexStr string) int {
	result := 0
	for i := 0; i < len(indexStr); i++ {
		result = result*26 + int(indexStr[i]-'A') + 1
	}
	return result
}
