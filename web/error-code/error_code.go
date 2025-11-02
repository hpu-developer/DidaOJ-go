package weberrorcode

import "meta/error-code"

const (
	UserNotMatch                         metaerrorcode.ErrorCode = 100001
	MailSendFail                         metaerrorcode.ErrorCode = 100002
	RegisterMailKeyError                 metaerrorcode.ErrorCode = 100003
	RegisterUserFail                     metaerrorcode.ErrorCode = 100004
	ForgetUserWithoutEmail               metaerrorcode.ErrorCode = 100005
	ForgetUserMailKeyError               metaerrorcode.ErrorCode = 100006
	ProblemTitleDuplicate                metaerrorcode.ErrorCode = 100007
	ContestTitleDuplicate                metaerrorcode.ErrorCode = 100008
	CollectionTitleDuplicate             metaerrorcode.ErrorCode = 100009
	ContestNotFoundProblem               metaerrorcode.ErrorCode = 100010
	ContestTooManyProblem                metaerrorcode.ErrorCode = 100011
	ContestCannotStartTimeBeforeNow      metaerrorcode.ErrorCode = 100012
	JudgeJobCannotApprove                metaerrorcode.ErrorCode = 100013
	ContestPostPasswordError             metaerrorcode.ErrorCode = 100014
	ProblemNotFound                      metaerrorcode.ErrorCode = 100015
	ProblemDailyAlreadyExists            metaerrorcode.ErrorCode = 100016
	ProblemDailyProblemAlreadyExists     metaerrorcode.ErrorCode = 100017
	ProblemJudgeDataMustZip              metaerrorcode.ErrorCode = 100018
	ProblemJudgeDataCannotDir            metaerrorcode.ErrorCode = 100019
	ProblemJudgeDataRuleYamlFail         metaerrorcode.ErrorCode = 100020
	ProblemJudgeDataSpjLanguageNotValid  metaerrorcode.ErrorCode = 100021
	ProblemJudgeDataSpjContentNotValid   metaerrorcode.ErrorCode = 100022
	ProblemJudgeDataSpjCompileFail       metaerrorcode.ErrorCode = 100023
	ProblemJudgeDataWithoutTask          metaerrorcode.ErrorCode = 100024
	ProblemJudgeDataTaskCountTooMany1000 metaerrorcode.ErrorCode = 100025
	ProblemJudgeDataTaskLoadFail         metaerrorcode.ErrorCode = 100026
	ProblemJudgeDataProcessWrapLineFail  metaerrorcode.ErrorCode = 100027
	ProblemJudgeDataProcessMd5Fail       metaerrorcode.ErrorCode = 100028
	ProblemJudgeDataSubmitFail           metaerrorcode.ErrorCode = 100029
	JudgeApproveCannotOriginOj           metaerrorcode.ErrorCode = 100030
	ProblemJudgeDataTooLarge20MB         metaerrorcode.ErrorCode = 100031
	JudgeListTooManySkip                 metaerrorcode.ErrorCode = 100032
	ProblemJudgeDataHasNotValid          metaerrorcode.ErrorCode = 100033
	UserNeedLogin                        metaerrorcode.ErrorCode = 100034
	ProblemCrawlCannotOriginOj           metaerrorcode.ErrorCode = 100035
	UserModifyVjudgeReload               metaerrorcode.ErrorCode = 100036
	UserModifyVjudgeCannotGet            metaerrorcode.ErrorCode = 100037
	UserModifyVjudgeVerifyFail           metaerrorcode.ErrorCode = 100038
	UserModifyOldEmailKeyError           metaerrorcode.ErrorCode = 100039
	UserModifyEmailKeyError              metaerrorcode.ErrorCode = 100040
)
