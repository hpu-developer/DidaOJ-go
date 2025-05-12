package foundationjudge

type JudgeLanguage int

var (
	JudgeLanguageUnknown JudgeLanguage = -1
	JudgeLanguageC       JudgeLanguage = 0
	JudgeLanguageCpp     JudgeLanguage = 1
	JudgeLanguageJava    JudgeLanguage = 2
	JudgeLanguagePython  JudgeLanguage = 3
	JudgeLanguagePascal  JudgeLanguage = 4
	JudgeLanguageGolang  JudgeLanguage = 5
)
