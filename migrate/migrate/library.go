package migrate

import (
	foundationjudge "foundation/foundation-judge"
)

func GetJudgeLanguageByCodeOJ(language int) foundationjudge.JudgeLanguage {
	switch language {
	case 0:
		return foundationjudge.JudgeLanguageC
	case 1:
		return foundationjudge.JudgeLanguageCpp

	case 3:
		return foundationjudge.JudgeLanguageJava
	case 6:
		return foundationjudge.JudgeLanguagePython
	default:
		return foundationjudge.JudgeLanguageUnknown
	}
}

func GetJudgeStatusByCodeOJ(status int) foundationjudge.JudgeStatus {
	return foundationjudge.JudgeStatus(status)
}
