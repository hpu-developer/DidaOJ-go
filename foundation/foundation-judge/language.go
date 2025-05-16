package foundationjudge

type JudgeLanguage int

const (
	JudgeLanguageUnknown JudgeLanguage = -1
	JudgeLanguageC       JudgeLanguage = 0
	JudgeLanguageCpp     JudgeLanguage = 1
	JudgeLanguageJava    JudgeLanguage = 2
	JudgeLanguagePython  JudgeLanguage = 3
	JudgeLanguagePascal  JudgeLanguage = 4
	JudgeLanguageGolang  JudgeLanguage = 5
	JudgeLanguageMax     JudgeLanguage = iota
)

func IsValidJudgeLanguage(language int) bool {
	return language > int(JudgeLanguageUnknown) && language < int(JudgeLanguageMax)
}

func GetLanguageByKey(language string) JudgeLanguage {
	switch language {
	case "c":
		return JudgeLanguageC
	case "cpp":
		return JudgeLanguageCpp
	case "java":
		return JudgeLanguageJava
	case "python":
		return JudgeLanguagePython
	case "golang":
		return JudgeLanguageGolang
	default:
		return JudgeLanguageUnknown
	}
}
