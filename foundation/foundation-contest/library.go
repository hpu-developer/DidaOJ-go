package foundationcontest

// GetContestProblemIndexStr 根据1返回A，2返回B，3返回C等
func GetContestProblemIndexStr(index int) string {
	result := ""
	for index > 0 {
		index--
		result = string(rune('A'+(index%26))) + result
		index = index / 26
	}
	return result
}

// GetContestProblemIndex 根据A返回1，B返回2，C返回3等
func GetContestProblemIndex(indexStr string) int {
	// 判断是否仅包含大写字母
	for i := 0; i < len(indexStr); i++ {
		if indexStr[i] < 'A' || indexStr[i] > 'Z' {
			return -1
		}
	}
	result := 0
	for i := 0; i < len(indexStr); i++ {
		result = result*26 + int(indexStr[i]-'A') + 1
	}
	return result
}
