package controller

import (
	"fmt"
	foundationerrorcode "foundation/error-code"
	foundationauth "foundation/foundation-auth"
	foundationdao "foundation/foundation-dao"
	foundationenum "foundation/foundation-enum"
	foundationmodel "foundation/foundation-model"
	foundationrequest "foundation/foundation-request"
	foundationservice "foundation/foundation-service"
	foundationuser "foundation/foundation-user"
	foundationview "foundation/foundation-view"
	"io"
	cfturnstile "meta/cf-turnstile"
	metacontroller "meta/controller"
	metaerrorcode "meta/error-code"
	metaemail "meta/meta-email"
	metaerror "meta/meta-error"
	metamath "meta/meta-math"
	metapanic "meta/meta-panic"
	metaresponse "meta/meta-response"
	metastring "meta/meta-string"
	metatime "meta/meta-time"
	"net/http"
	"strconv"
	"strings"
	"time"
	"web/config"
	weberrorcode "web/error-code"
	"web/request"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	metacontroller.Controller
}

func (c *UserController) GetInfo(ctx *gin.Context) {
	username := ctx.Query("username")
	if username == "" {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	userInfo, err := foundationservice.GetUserService().GetInfoByUsername(ctx, username)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	if userInfo == nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.NotFound, nil)
		return
	}

	// 计算升级相关的经验值信息
	if userInfo.Experience >= 0 && userInfo.Level >= 0 {
		// 使用foundationuser包中的通用函数计算总经验
		// 计算当前等级升级所需总经验
		userInfo.ExperienceUpgrade = foundationuser.GetTotalExperienceForLevel(userInfo.Level + 1)
		// 计算当前等级段已积攒经验（当前总经验 - 上一等级升级所需总经验）
		userInfo.ExperienceCurrentLevel = userInfo.Experience - foundationuser.GetTotalExperienceForLevel(userInfo.Level)
		// 确保当前等级进度不为负数
		if userInfo.ExperienceCurrentLevel < 0 {
			userInfo.ExperienceCurrentLevel = 0
		}
	}
	acProblems, attemptProblems, err := foundationservice.GetJudgeService().GetUserAttemptProblems(ctx, userInfo.Id)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}

	newYear := metatime.GetCurrentYear()
	userStatic, err := foundationservice.GetJudgeService().GetUserJudgeJobCountStatics(ctx, userInfo.Id, newYear)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}

	responseData := struct {
		User           *foundationview.UserInfo               `json:"user"`
		ProblemsAc     []*foundationview.ProblemViewKey       `json:"problems_ac"`
		ProblemAttempt []*foundationview.ProblemViewKey       `json:"problems_attempt"`
		Statics        []*foundationview.JudgeJobCountStatics `json:"statics"`
	}{
		User:           userInfo,
		ProblemsAc:     acProblems,
		ProblemAttempt: attemptProblems,
		Statics:        userStatic,
	}
	metaresponse.NewResponse(ctx, metaerrorcode.Success, responseData)
}

func (c *UserController) GetModifyInfo(ctx *gin.Context) {
	userId, err := foundationauth.GetUserIdFromContext(ctx)
	if err != nil {
		metaresponse.NewResponse(ctx, weberrorcode.UserNeedLogin, nil)
		return
	}
	userInfo, err := foundationservice.GetUserService().GetModifyInfo(ctx, userId)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	metaresponse.NewResponse(ctx, metaerrorcode.Success, userInfo)
}

func (c *UserController) PostModify(ctx *gin.Context) {
	userId, err := foundationauth.GetUserIdFromContext(ctx)
	if err != nil {
		metaresponse.NewResponse(ctx, weberrorcode.UserNeedLogin, nil)
		return
	}
	var requestData foundationrequest.UserModifyInfo
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	nickname := requestData.Nickname
	if len(nickname) < 1 || len(nickname) > 30 {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	slogan := requestData.Slogan
	if len(slogan) > 100 {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	err = foundationservice.GetUserService().UpdateUserInfo(ctx, userId, &requestData, metatime.GetTimeNow())
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	metaresponse.NewResponse(ctx, metaerrorcode.Success)
}

func (c *UserController) PostModifyPassword(ctx *gin.Context) {
	userId, err := foundationauth.GetUserIdFromContext(ctx)
	if err != nil {
		metaresponse.NewResponse(ctx, weberrorcode.UserNeedLogin, nil)
		return
	}
	var requestData foundationrequest.UserModifyPassword
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	if !foundationuser.IsValidPassword(requestData.Password) {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	err = foundationservice.GetUserService().UpdateUserPassword(ctx, userId, &requestData, metatime.GetTimeNow())
	if err != nil {
		metaresponse.NewResponseError(ctx, err)
		return
	}
	metaresponse.NewResponse(ctx, metaerrorcode.Success)
}

func (c *UserController) PostModifyEmail(ctx *gin.Context) {
	var requestDatastruct struct {
		EmailKey    string `json:"email_key" binding:"required"`
		NewEmail    string `json:"new_email" binding:"required"`
		NewEmailKey string `json:"new_email_key" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&requestDatastruct); err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	userId, err := foundationauth.GetUserIdFromContext(ctx)
	if err != nil {
		metaresponse.NewResponse(ctx, weberrorcode.UserNeedLogin, nil)
		return
	}
	// 判断验证码是否正确
	kvStoreDao := foundationdao.GetKVStoreDao()
	oldCodeKey := fmt.Sprintf("modify_email_old_key_%d", userId)
	storedOldCode, err := kvStoreDao.GetValue(ctx, oldCodeKey)
	if storedOldCode == nil || string(*storedOldCode) != requestDatastruct.EmailKey {
		metaresponse.NewResponse(ctx, weberrorcode.UserModifyOldEmailKeyError, nil)
		return
	}
	newCodeKey := fmt.Sprintf("modify_email_key_%d_%s", userId, requestDatastruct.NewEmail)
	storedNewCode, err := kvStoreDao.GetValue(ctx, newCodeKey)
	if storedNewCode == nil || string(*storedNewCode) != requestDatastruct.NewEmailKey {
		metaresponse.NewResponse(ctx, weberrorcode.UserModifyEmailKeyError, nil)
		return
	}
	// 删除所有的验证码
	if err := kvStoreDao.DeleteValue(ctx, oldCodeKey); err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	if err := kvStoreDao.DeleteValue(ctx, newCodeKey); err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	err = foundationservice.GetUserService().UpdateUserEmail(
		ctx,
		userId,
		requestDatastruct.NewEmail,
		metatime.GetTimeNow(),
	)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	metaresponse.NewResponse(ctx, metaerrorcode.Success)
}

func (c *UserController) PostModifyEmailKey(ctx *gin.Context) {
	var requestData struct {
		Email string `json:"email" binding:"required"`
		Token string `json:"token" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}

	userId, err := foundationauth.GetUserIdFromContext(ctx)
	if err != nil {
		metaresponse.NewResponse(ctx, weberrorcode.UserNeedLogin, nil)
		return
	}
	email := requestData.Email

	if !metaemail.IsEmailValid(email) {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}

	isTurnstileValid, err := cfturnstile.IsTurnstileTokenValid(ctx, config.GetConfig().CfTurnstile, requestData.Token)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	if !isTurnstileValid {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}

	kvStoreDao := foundationdao.GetKVStoreDao()
	flagKey := fmt.Sprintf("modify_email_dup_%d_%s", userId, email)
	codeKey := fmt.Sprintf("modify_email_key_%d_%s", userId, email)

	// 检查是否在 1 分钟内重复发送
	ok, err := kvStoreDao.SetNXValue(ctx, flagKey, []byte("1"), time.Minute)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	if !ok {
		metaresponse.NewResponse(ctx, metaerrorcode.TooManyRequests, nil)
		return
	}

	// 生成验证码并存入 KV 存储，设置10分钟过期
	code := strconv.Itoa(metamath.GetRandomInt(100000, 999999))
	if err := kvStoreDao.SetValue(ctx, codeKey, []byte(code), 10*time.Minute); err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}

	subject := fmt.Sprintf("[DidaOJ] - 邮件验证码")
	body := fmt.Sprintf(
		"%s：\n\n您好！\n欢迎您使用DidaOJ，以下是您的邮箱验证码：\n\n%s\n\n本验证码用于您修改本系统绑定的邮箱，请勿泄露给他人。\n请在10分钟之内使用本验证码，过期请重新申请。\n如有疑问，请联系管理员。\n\n祝好！\nDidaOJ团队\nhttps://oj.didapipa.com",
		email, code,
	)

	err = metaemail.SendEmail(
		"DidaOJ",
		config.GetConfig().Email.Email,
		config.GetConfig().Email.Password,
		config.GetConfig().Email.Host,
		config.GetConfig().Email.Port,
		email,
		subject,
		body,
	)
	if err != nil {
		metaresponse.NewResponse(ctx, weberrorcode.MailSendFail, nil)
		return
	}

	metaresponse.NewResponse(ctx, metaerrorcode.Success)
}

func (c *UserController) PostModifyEmailKeyOld(ctx *gin.Context) {
	var requestData struct {
		Token string `json:"token" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}

	userId, err := foundationauth.GetUserIdFromContext(ctx)
	if err != nil {
		metaresponse.NewResponse(ctx, weberrorcode.UserNeedLogin, nil)
		return
	}
	emailPtr, err := foundationservice.GetUserService().GetEmail(ctx, userId)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	if emailPtr == nil || *emailPtr == "" {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	email := *emailPtr

	if !metaemail.IsEmailValid(email) {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}

	isTurnstileValid, err := cfturnstile.IsTurnstileTokenValid(ctx, config.GetConfig().CfTurnstile, requestData.Token)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	if !isTurnstileValid {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}

	kvStoreDao := foundationdao.GetKVStoreDao()
	flagKey := fmt.Sprintf("modify_email_old_dup_%d", userId)
	codeKey := fmt.Sprintf("modify_email_old_key_%d", userId)

	// 检查是否在 1 分钟内重复发送
	ok, err := kvStoreDao.SetNXValue(ctx, flagKey, []byte("1"), time.Minute)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	if !ok {
		metaresponse.NewResponse(ctx, metaerrorcode.TooManyRequests, nil)
		return
	}

	// 生成验证码并存入 KV 存储，设置10分钟过期
	code := strconv.Itoa(metamath.GetRandomInt(100000, 999999))
	if err := kvStoreDao.SetValue(ctx, codeKey, []byte(code), 10*time.Minute); err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}

	subject := fmt.Sprintf("[DidaOJ] - 邮件验证码")
	body := fmt.Sprintf(
		"%s：\n\n您好！\n欢迎您使用DidaOJ，以下是您的邮箱验证码：\n\n%s\n\n本验证码用于您修改本系统绑定的邮箱，请勿泄露给他人。\n请在10分钟之内使用本验证码，过期请重新申请。\n如有疑问，请联系管理员。\n\n祝好！\nDidaOJ团队\nhttps://oj.didapipa.com",
		email, code,
	)

	err = metaemail.SendEmail(
		"DidaOJ",
		config.GetConfig().Email.Email,
		config.GetConfig().Email.Password,
		config.GetConfig().Email.Host,
		config.GetConfig().Email.Port,
		email,
		subject,
		body,
	)
	if err != nil {
		metaresponse.NewResponse(ctx, weberrorcode.MailSendFail, nil)
		return
	}

	metaresponse.NewResponse(ctx, metaerrorcode.Success)
}

func (c *UserController) PostModifyVjudge(ctx *gin.Context) {
	userId, err := foundationauth.GetUserIdFromContext(ctx)
	if err != nil {
		metaresponse.NewResponse(ctx, weberrorcode.UserNeedLogin, nil)
		return
	}
	var requestData foundationrequest.UserModifyVjudge
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	username := requestData.Username
	if len(username) > 30 {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}

	if username == "" {
		err = foundationservice.GetUserService().UpdateUserVjudgeUsername(ctx, userId, username, metatime.GetTimeNow())
		if err != nil {
			metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
			return
		}
		metaresponse.NewResponse(ctx, metaerrorcode.Success)
		return
	}

	kvStoreDao := foundationdao.GetKVStoreDao()
	codeKey := fmt.Sprintf("modify_vjudge_%d", userId)

	if requestData.Approved {
		randomKeyBytes, err := kvStoreDao.GetValue(ctx, codeKey)
		if err != nil {
			metaresponse.NewResponse(ctx, weberrorcode.UserModifyVjudgeReload, nil)
			return
		}
		if randomKeyBytes == nil {
			metaresponse.NewResponse(ctx, weberrorcode.UserModifyVjudgeReload, nil)
			return
		}
		randomKey := string(*randomKeyBytes)
		if randomKey == "" {
			metaresponse.NewResponse(ctx, weberrorcode.UserModifyVjudgeReload, nil)
			return
		}
		vjudgeUrl := fmt.Sprintf("https://vjudge.net/user/%s", username)
		// 请求页面信息是否包含randomKey
		response, err := http.Get(vjudgeUrl)
		if err != nil {
			metaresponse.NewResponse(ctx, weberrorcode.UserModifyVjudgeCannotGet, nil)
			return
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				metapanic.ProcessError(metaerror.Wrap(err, "defer http body close failed"))
				return
			}
		}(response.Body)
		body, err := io.ReadAll(response.Body)
		if err != nil {
			metaresponse.NewResponse(ctx, weberrorcode.UserModifyVjudgeReload, nil)
			return
		}
		if !strings.Contains(string(body), randomKey) {
			metaresponse.NewResponse(ctx, weberrorcode.UserModifyVjudgeVerifyFail, nil)
			return
		}
		err = foundationservice.GetUserService().UpdateUserVjudgeUsername(ctx, userId, username, metatime.GetTimeNow())
		if err != nil {
			metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
			return
		}
		metaresponse.NewResponse(ctx, metaerrorcode.Success)
	} else {
		randomString := metastring.GetRandomString(16)
		err = kvStoreDao.SetValue(ctx, codeKey, []byte(randomString), 10*time.Minute)
		if err != nil {
			metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
			return
		}
		metaresponse.NewResponse(ctx, metaerrorcode.Success, randomString)
	}
}

func (c *UserController) PostAccountInfos(ctx *gin.Context) {
	var requestData struct {
		Users []int `json:"users" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	userAccountInfos, err := foundationservice.GetUserService().GetUserAccountInfos(ctx, requestData.Users)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	responseData := struct {
		Users []*foundationview.UserAccountInfo `json:"users"`
	}{
		Users: userAccountInfos,
	}
	metaresponse.NewResponse(ctx, metaerrorcode.Success, responseData)
}

func (c *UserController) PostParse(ctx *gin.Context) {
	var requestData struct {
		Users []string `json:"users" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	usernameList := requestData.Users
	userAccountInfos, err := foundationservice.GetUserService().GetUserAccountInfoByUsernames(ctx, usernameList)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	responseData := struct {
		Users []*foundationview.UserAccountInfo `json:"users"`
	}{
		Users: userAccountInfos,
	}
	metaresponse.NewResponse(ctx, metaerrorcode.Success, responseData)
}

func (c *UserController) PostRegisterEmail(ctx *gin.Context) {
	var requestData struct {
		Token string `json:"token" binding:"required"`
		Email string `json:"email" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	email := requestData.Email
	if email == "" {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	if !metaemail.IsEmailValid(email) {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}

	isTurnstileValid, err := cfturnstile.IsTurnstileTokenValid(ctx, config.GetConfig().CfTurnstile, requestData.Token)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	if !isTurnstileValid {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}

	kvStoreDao := foundationdao.GetKVStoreDao()
	flagKey := fmt.Sprintf("register_email_dup_%s", email)
	codeKey := fmt.Sprintf("register_email_key_%s", email)

	// 检查是否在 1 分钟内重复发送
	ok, err := kvStoreDao.SetNXValue(ctx, flagKey, []byte("1"), time.Minute)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	if !ok {
		metaresponse.NewResponse(ctx, metaerrorcode.TooManyRequests, nil)
		return
	}

	// 生成验证码并存入 KV 存储，设置10分钟过期
	code := strconv.Itoa(metamath.GetRandomInt(100000, 999999))
	if err := kvStoreDao.SetValue(ctx, codeKey, []byte(code), 10*time.Minute); err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}

	subject := fmt.Sprintf("[DidaOJ] - 邮件验证码")
	body := fmt.Sprintf(
		"%s：\n\n您好！\n欢迎您使用DidaOJ，以下是您的邮箱验证码：\n\n%s\n\n本验证码用于您注册本系统的账号，请勿泄露给他人。\n请在10分钟之内使用本验证码，过期请重新申请。\n如有疑问，请联系管理员。\n\n祝好！\nDidaOJ团队\nhttps://oj.didapipa.com",
		email, code,
	)

	err = metaemail.SendEmail(
		"DidaOJ",
		config.GetConfig().Email.Email,
		config.GetConfig().Email.Password,
		config.GetConfig().Email.Host,
		config.GetConfig().Email.Port,
		requestData.Email,
		subject,
		body,
	)
	if err != nil {
		metaresponse.NewResponse(ctx, weberrorcode.MailSendFail, nil)
		return
	}

	metaresponse.NewResponse(ctx, metaerrorcode.Success)
}

func (c *UserController) PostRegister(ctx *gin.Context) {
	var requestData struct {
		Username     string `json:"username" binding:"required"`
		Password     string `json:"password" binding:"required"`
		Nickname     string `json:"nickname" binding:"required"`
		RealName     string `json:"real_name,omitempty"`
		Gender       string `json:"gender,omitempty"`
		Organization string `json:"organization,omitempty"`
		Email        string `json:"email" binding:"required"`
		Key          string `json:"key" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	if requestData.Username == "" || requestData.Password == "" || requestData.Nickname == "" || requestData.Email == "" {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	// 判断username是否仅包含字母数字下划线
	if !foundationuser.IsValidUsername(requestData.Username) {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	// 判断password是否>6并且<20
	if !foundationuser.IsValidPassword(requestData.Password) {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	nickname := requestData.Nickname
	if len(nickname) < 1 || len(nickname) > 30 {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	realName := requestData.RealName
	if len(realName) > 20 {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	organization := requestData.Organization
	if len(organization) > 30 {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}

	email := requestData.Email
	if email == "" {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	if !metaemail.IsEmailValid(email) {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}

	kvStoreDao := foundationdao.GetKVStoreDao()
	codeKey := fmt.Sprintf("register_email_key_%s", email)

	// 获取存储的验证码
	storedCodeRaw, err := kvStoreDao.GetValue(ctx, codeKey)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	if storedCodeRaw == nil || string(*storedCodeRaw) != requestData.Key {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}

	// 删除验证码
	if err := kvStoreDao.DeleteValue(ctx, codeKey); err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}

	passwordEncode, err := foundationservice.GetUserService().GeneratePasswordEncode(requestData.Password)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}

	nowTime := metatime.GetTimeNow()

	gender := foundationenum.GetUserGender(requestData.Gender)

	user := foundationmodel.NewUserBuilder().
		Username(requestData.Username).
		Password(passwordEncode).
		Email(requestData.Email).
		Nickname(requestData.Nickname).
		Gender(gender).
		RealName(&realName).
		Organization(&organization).
		InsertTime(nowTime).
		ModifyTime(nowTime).
		Level(1).      // 默认初始等级为1
		Experience(0). // 默认初始经验为0
		Build()

	err = foundationservice.GetUserService().InsertUser(ctx, user)
	if err != nil {
		metaresponse.NewResponse(ctx, weberrorcode.RegisterUserFail)
		return
	}

	metaresponse.NewResponse(ctx, metaerrorcode.Success)
}

func (c *UserController) PostForget(ctx *gin.Context) {
	var requestData struct {
		Token    string `json:"token" binding:"required"`
		Username string `json:"username" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	username := requestData.Username
	if !foundationuser.IsValidUsername(requestData.Username) {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	isTurnstileValid, err := cfturnstile.IsTurnstileTokenValid(ctx, config.GetConfig().CfTurnstile, requestData.Token)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	if !isTurnstileValid {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	userEmail, err := foundationservice.GetUserService().GetEmailByUsername(ctx, username)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	if userEmail == nil || !metaemail.IsEmailValid(*userEmail) {
		metaresponse.NewResponse(ctx, weberrorcode.ForgetUserWithoutEmail, nil)
		return
	}

	kvStoreDao := foundationdao.GetKVStoreDao()
	codeKey := fmt.Sprintf("forget_password_key_%s", username)

	// 生成验证码并存入 KV 存储，设置10分钟过期
	code := strconv.Itoa(metamath.GetRandomInt(100000, 999999))
	if err := kvStoreDao.SetValue(ctx, codeKey, []byte(code), 10*time.Minute); err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}

	subject := fmt.Sprintf("[DidaOJ] - 邮件验证码")
	body := fmt.Sprintf(
		"%s：\n\n您好！\n欢迎您使用DidaOJ，以下是您的邮箱验证码：\n\n%s\n\n本验证码用于重置本系统的账号，请勿泄露给他人。\n请在10分钟之内使用本验证码，过期请重新申请。\n如有疑问，请联系管理员。\n\n祝好！\nDidaOJ团队\nhttps://oj.didapipa.com",
		*userEmail, code,
	)

	err = metaemail.SendEmail(
		"DidaOJ",
		config.GetConfig().Email.Email,
		config.GetConfig().Email.Password,
		config.GetConfig().Email.Host,
		config.GetConfig().Email.Port,
		*userEmail,
		subject,
		body,
	)
	if err != nil {
		metaresponse.NewResponse(ctx, weberrorcode.MailSendFail, nil)
		return
	}

	email := metaemail.MaskEmail(*userEmail)
	metaresponse.NewResponse(ctx, metaerrorcode.Success, email)
}

func (c *UserController) PostPasswordForget(ctx *gin.Context) {
	var requestData struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
		Key      string `json:"key" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	username := requestData.Username
	if !foundationuser.IsValidUsername(requestData.Username) {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	if !foundationuser.IsValidPassword(requestData.Password) {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	kvStoreDao := foundationdao.GetKVStoreDao()
	codeKey := fmt.Sprintf("forget_password_key_%s", username)
	storedCodeRaw, err := kvStoreDao.GetValue(ctx, codeKey)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	if storedCodeRaw == nil || string(*storedCodeRaw) != requestData.Key {
		metaresponse.NewResponse(ctx, weberrorcode.ForgetUserMailKeyError, nil)
		return
	}
	// 删除验证码
	if err := kvStoreDao.DeleteValue(ctx, codeKey); err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	passwordEncode, err := foundationservice.GetUserService().GeneratePasswordEncode(requestData.Password)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	nowTime := metatime.GetTimeNow()
	err = foundationservice.GetUserService().UpdatePassword(ctx, requestData.Username, passwordEncode, nowTime)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	metaresponse.NewResponse(ctx, metaerrorcode.Success)
}

func (c *UserController) PostLoginRefresh(ctx *gin.Context) {
	userId, err := foundationauth.GetUserIdFromContext(ctx)
	if err != nil {
		metaresponse.NewResponse(ctx, weberrorcode.UserNeedLogin, nil)
		return
	}
	nowTime := metatime.GetTimeNow()
	loginResponse, err := foundationservice.GetUserService().GetUserLoginResponse(ctx, userId, nowTime)
	if err != nil {
		metaresponse.NewResponseError(ctx, err, nil)
		return
	}
	if loginResponse == nil {
		metaresponse.NewResponse(ctx, weberrorcode.UserNotMatch, nil)
		return
	}
	// 记录登录日志
	err = foundationservice.GetUserService().PostLoginLog(
		ctx,
		loginResponse.Id,
		nowTime,
		ctx.ClientIP(),
		ctx.Request.UserAgent(),
	)
	if err != nil {
		metapanic.ProcessError(err)
	}

	metaresponse.NewResponse(ctx, metaerrorcode.Success, loginResponse)
}

// GetCheckinToday 获取今日签到人数和用户签到状态
func (c *UserController) GetCheckinToday(ctx *gin.Context) {
	// 获取当前日期，格式为"2006-01-02"
	nowTime := metatime.GetTimeNow()
	today := nowTime.Format("2006-01-02")

	// 默认签到状态为false
	checkIn := false

	// 尝试获取用户ID并检查签到状态
	userId, err := foundationauth.GetUserIdFromContext(ctx)
	if err == nil {
		checkIn, err = foundationservice.GetUserService().IsUserCheckedIn(ctx, userId, today)
		if err != nil {
			metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
			return
		}
	}

	count := -1
	if checkIn {
		// 查询今日签到人数
		count, err = foundationservice.GetUserService().GetCheckinCount(ctx, today)
		if err != nil {
			metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
			return
		}
	}

	// 返回新的结构 {count: 人数, check_in: 是否签到}
	responseData := struct {
		Count   int  `json:"count"`
		CheckIn bool `json:"check_in"`
	}{
		Count:   count,
		CheckIn: checkIn,
	}

	metaresponse.NewResponse(ctx, metaerrorcode.Success, responseData)
}

func (c *UserController) PostCheckin(ctx *gin.Context) {
	userId, err := foundationauth.GetUserIdFromContext(ctx)
	if err != nil {
		metaresponse.NewResponse(ctx, weberrorcode.UserNeedLogin, nil)
		return
	}
	nowTime := metatime.GetTimeNow()
	award, err := foundationservice.GetUserService().AddExperienceForCheckIn(ctx, userId, nowTime)
	if err != nil {
		metaresponse.NewResponseError(ctx, err, nil)
		return
	}
	if award == nil {
		metaresponse.NewResponse(ctx, weberrorcode.UserCheckinAlreadyDone, true)
		return
	}
	metaresponse.NewResponse(ctx, metaerrorcode.Success, award)
}

func (c *UserController) PostLogin(ctx *gin.Context) {
	var userLoginRequest request.UserLogin
	if err := ctx.ShouldBindJSON(&userLoginRequest); err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	if userLoginRequest.Username == "" || userLoginRequest.Password == "" {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}

	nowTime := metatime.GetTimeNow()

	loginResponse, err := foundationservice.GetUserService().Login(
		ctx,
		userLoginRequest.Username,
		userLoginRequest.Password,
		nowTime,
	)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	if loginResponse == nil {
		metaresponse.NewResponse(ctx, weberrorcode.UserNotMatch, nil)
		return
	}

	err = foundationservice.GetUserService().PostLoginLog(
		ctx,
		loginResponse.Id,
		nowTime,
		ctx.ClientIP(),
		ctx.Request.UserAgent(),
	)
	if err != nil {
		metapanic.ProcessError(err)
	}

	metaresponse.NewResponse(ctx, metaerrorcode.Success, loginResponse)
}
