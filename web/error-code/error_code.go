package weberrorcode

import "meta/error-code"

const (
	UserNotMatch                    metaerrorcode.ErrorCode = 100001
	RegisterMailSendFail            metaerrorcode.ErrorCode = 100002
	RegisterMailKeyError            metaerrorcode.ErrorCode = 100003
	RegisterUserFail                metaerrorcode.ErrorCode = 100004
	ForgetUserWithoutEmail          metaerrorcode.ErrorCode = 100005
	ForgetUserMailKeyError          metaerrorcode.ErrorCode = 100006
	ProblemTitleDuplicate           metaerrorcode.ErrorCode = 100007
	ContestTitleDuplicate           metaerrorcode.ErrorCode = 100008
	CollectionTitleDuplicate        metaerrorcode.ErrorCode = 100009
	ContestNotFoundProblem          metaerrorcode.ErrorCode = 100010
	ContestTooManyProblem           metaerrorcode.ErrorCode = 100011
	ContestCannotStartTimeBeforeNow metaerrorcode.ErrorCode = 100012
	JudgeJobCannotApprove           metaerrorcode.ErrorCode = 100013
	ContestPostPasswordError        metaerrorcode.ErrorCode = 100014
)
