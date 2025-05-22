package weberrorcode

import "meta/error-code"

const (
	UserNotMatch           metaerrorcode.ErrorCode = 100001
	RegisterMailSendFail   metaerrorcode.ErrorCode = 100002
	RegisterMailKeyError   metaerrorcode.ErrorCode = 100003
	RegisterUserFail       metaerrorcode.ErrorCode = 100004
	ForgetUserWithoutEmail metaerrorcode.ErrorCode = 100005
	ForgetUserMailKeyError metaerrorcode.ErrorCode = 100006
)
