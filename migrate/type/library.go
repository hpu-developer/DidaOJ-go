package migratetype

import (
	foundationjudge "foundation/foundation-judge"
)

func GetJudgeLanguageByCodeOJ(language int) foundationjudge.JudgeLanguage {
	switch language {
	case 0:
		return foundationjudge.JudgeLanguageC
	case 1:
		return foundationjudge.JudgeLanguageCpp
	case 2:
		return foundationjudge.JudgeLanguagePascal
	case 3:
		return foundationjudge.JudgeLanguageJava
	case 6:
		return foundationjudge.JudgeLanguagePython
	case 10:
		return foundationjudge.JudgeLanguageCpp
	case 14:
		return foundationjudge.JudgeLanguageCpp
	case 16:
		return foundationjudge.JudgeLanguageCpp
	default:
		return foundationjudge.JudgeLanguageUnknown
	}
}

func GetJudgeLanguageByVhoj(language string) foundationjudge.JudgeLanguage {
	switch language {
	case "C":
		return foundationjudge.JudgeLanguageC
	case "CPP":
		return foundationjudge.JudgeLanguageCpp
	case "JAVA":
		return foundationjudge.JudgeLanguageJava
	case "PYTHON":
		return foundationjudge.JudgeLanguagePython
	default:
		return foundationjudge.JudgeLanguageUnknown
	}
}

func GetJudgeStatusByCodeOJ(status int) foundationjudge.JudgeStatus {
	switch status {
	case 0:
		return foundationjudge.JudgeStatusInit
	case 1:
		return foundationjudge.JudgeStatusRejudge
	case 2:
		return foundationjudge.JudgeStatusCompiling
	case 3:
		return foundationjudge.JudgeStatusRunning
	case 4:
		return foundationjudge.JudgeStatusAC
	case 5:
		return foundationjudge.JudgeStatusPE
	case 6:
		return foundationjudge.JudgeStatusWA
	case 7:
		return foundationjudge.JudgeStatusTLE
	case 8:
		return foundationjudge.JudgeStatusMLE
	case 9:
		return foundationjudge.JudgeStatusOLE
	case 10:
		return foundationjudge.JudgeStatusRE
	case 11:
		return foundationjudge.JudgeStatusCE
	case 12:
		return foundationjudge.JudgeStatusCLE
	case 13:
		return foundationjudge.JudgeStatusJudgeFail
	default:
		return foundationjudge.JudgeStatusUnknown
	}
}

func GetJudgeStatusByVhoj(status string) foundationjudge.JudgeStatus {
	switch status {
	case "PENDING":
		return foundationjudge.JudgeStatusInit
	case "SUBMITTED":
		fallthrough
	case "QUEUEING":
		return foundationjudge.JudgeStatusQueuing
	case "COMPILING":
		return foundationjudge.JudgeStatusCompiling
	case "JUDGING":
		return foundationjudge.JudgeStatusRunning
	case "AC":
		return foundationjudge.JudgeStatusAC
	case "PE":
		return foundationjudge.JudgeStatusPE
	case "WA":
		return foundationjudge.JudgeStatusWA
	case "TLE":
		return foundationjudge.JudgeStatusTLE
	case "MLE":
		return foundationjudge.JudgeStatusMLE
	case "OLE":
		return foundationjudge.JudgeStatusOLE
	case "RE":
		return foundationjudge.JudgeStatusRE
	case "CE":
		return foundationjudge.JudgeStatusCE
	case "CLE":
		return foundationjudge.JudgeStatusCLE
	case "SUBMIT_FAILED_TEMP":
		fallthrough
	case "SUBMIT_FAILED_PERM":
		return foundationjudge.JudgeStatusSubmitFail
	case "FAILED_OTHER":
		return foundationjudge.JudgeStatusJudgeFail
	default:
		return foundationjudge.JudgeStatusUnknown
	}
}

func GetJudgeLanguageByEOJ(language string) foundationjudge.JudgeLanguage {
	switch language {
	case "C":
		fallthrough
	case "C With O2":
		return foundationjudge.JudgeLanguageC
	case "C++":
		fallthrough
	case "C++ 17":
		fallthrough
	case "C++ 17 With O2":
		fallthrough
	case "C++ 20 With O2":
		fallthrough
	case "C++ With O2":
		fallthrough
	case "C++ 20":
		return foundationjudge.JudgeLanguageCpp
	case "PyPy3":
		fallthrough
	case "Python2":
		fallthrough
	case "Python3":
		return foundationjudge.JudgeLanguagePython
	case "Java":
		return foundationjudge.JudgeLanguageJava
	}
	return foundationjudge.JudgeLanguageUnknown
}
