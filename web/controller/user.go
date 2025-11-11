package controller

import (
	"errors"
	"fmt"
	foundationerrorcode "foundation/error-code"
	foundationauth "foundation/foundation-auth"
	foundationenum "foundation/foundation-enum"
	foundationmodel "foundation/foundation-model"
	"foundation/foundation-request"
	foundationservice "foundation/foundation-service"
	foundationuser "foundation/foundation-user"
	foundationview "foundation/foundation-view"
	"io"
	cfturnstile "meta/cf-turnstile"
	metacontroller "meta/controller"
	"meta/error-code"
	metaemail "meta/meta-email"
	metaerror "meta/meta-error"
	metamath "meta/meta-math"
	metapanic "meta/meta-panic"
	metaredis "meta/meta-redis"
	"meta/meta-response"
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
	"github.com/redis/go-redis/v9"
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
	redisClient := metaredis.GetSubsystem().GetClient()
	oldCodeKey := fmt.Sprintf("modify_email_old_key_%d", userId)
	storedOldCode, err := redisClient.Get(ctx, oldCodeKey).Result()
	if storedOldCode != requestDatastruct.EmailKey {
		metaresponse.NewResponse(ctx, weberrorcode.UserModifyOldEmailKeyError, nil)
		return
	}
	newCodeKey := fmt.Sprintf("modify_email_key_%d_%s", userId, requestDatastruct.NewEmail)
	storedNewCode, err := redisClient.Get(ctx, newCodeKey).Result()
	if storedNewCode != requestDatastruct.NewEmailKey {
		metaresponse.NewResponse(ctx, weberrorcode.UserModifyEmailKeyError, nil)
		return
	}
	// 删除所有的验证码
	if err := redisClient.Del(ctx, oldCodeKey).Err(); err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	if err := redisClient.Del(ctx, newCodeKey).Err(); err != nil {
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

	redisClient := metaredis.GetSubsystem().GetClient()
	flagKey := fmt.Sprintf("modify_email_dup_%d_%s", userId, email)
	codeKey := fmt.Sprintf("modify_email_key_%d_%s", userId, email)

	// 检查是否在 1 分钟内重复发送
	ok, err := redisClient.SetNX(ctx, flagKey, 1, time.Minute).Result()
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	if !ok {
		metaresponse.NewResponse(ctx, metaerrorcode.TooManyRequests, nil)
		return
	}

	// 生成验证码并存入 Redis，设置10分钟过期
	code := strconv.Itoa(metamath.GetRandomInt(100000, 999999))
	if err := redisClient.Set(ctx, codeKey, code, 10*time.Minute).Err(); err != nil {
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

	redisClient := metaredis.GetSubsystem().GetClient()
	flagKey := fmt.Sprintf("modify_email_old_dup_%d", userId)
	codeKey := fmt.Sprintf("modify_email_old_key_%d", userId)

	// 检查是否在 1 分钟内重复发送
	ok, err := redisClient.SetNX(ctx, flagKey, 1, time.Minute).Result()
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	if !ok {
		metaresponse.NewResponse(ctx, metaerrorcode.TooManyRequests, nil)
		return
	}

	// 生成验证码并存入 Redis，设置10分钟过期
	code := strconv.Itoa(metamath.GetRandomInt(100000, 999999))
	if err := redisClient.Set(ctx, codeKey, code, 10*time.Minute).Err(); err != nil {
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

	redisClient := metaredis.GetSubsystem().GetClient()
	if redisClient == nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	codeKey := fmt.Sprintf("modify_vjudge_%d", userId)

	if requestData.Approved {
		randomKey, err := redisClient.Get(ctx, codeKey).Result()
		if err != nil {
			metaresponse.NewResponse(ctx, weberrorcode.UserModifyVjudgeReload, nil)
			return
		}
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
		_, err = redisClient.Set(ctx, codeKey, randomString, 10*time.Minute).Result()
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

	redisClient := metaredis.GetSubsystem().GetClient()
	flagKey := fmt.Sprintf("register_email_dup_%s", email)
	codeKey := fmt.Sprintf("register_email_key_%s", email)

	// 检查是否在 1 分钟内重复发送
	ok, err := redisClient.SetNX(ctx, flagKey, 1, time.Minute).Result()
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	if !ok {
		metaresponse.NewResponse(ctx, metaerrorcode.TooManyRequests, nil)
		return
	}

	// 生成验证码并存入 Redis，设置10分钟过期
	code := strconv.Itoa(metamath.GetRandomInt(100000, 999999))
	if err := redisClient.Set(ctx, codeKey, code, 10*time.Minute).Err(); err != nil {
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

	codeKey := fmt.Sprintf("register_email_key_%s", email)
	redisClient := metaredis.GetSubsystem().GetClient()

	storedCode, err := redisClient.Get(ctx, codeKey).Result()
	if errors.Is(err, redis.Nil) {
		metaresponse.NewResponse(ctx, weberrorcode.RegisterMailKeyError, nil)
		return
	} else if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	if storedCode != requestData.Key {
		metaresponse.NewResponse(ctx, weberrorcode.RegisterMailKeyError, nil)
		return
	}

	// 删除验证码
	if err := redisClient.Del(ctx, codeKey).Err(); err != nil {
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

	redisClient := metaredis.GetSubsystem().GetClient()
	codeKey := fmt.Sprintf("forget_password_key_%s", username)

	// 生成验证码并存入 Redis，设置10分钟过期
	code := strconv.Itoa(metamath.GetRandomInt(100000, 999999))
	if err := redisClient.Set(ctx, codeKey, code, 10*time.Minute).Err(); err != nil {
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
	codeKey := fmt.Sprintf("forget_password_key_%s", username)
	redisClient := metaredis.GetSubsystem().GetClient()
	storedCode, err := redisClient.Get(ctx, codeKey).Result()
	if errors.Is(err, redis.Nil) {
		metaresponse.NewResponse(ctx, weberrorcode.ForgetUserMailKeyError, nil)
		return
	}
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	if storedCode != requestData.Key {
		metaresponse.NewResponse(ctx, weberrorcode.ForgetUserMailKeyError, nil)
		return
	}
	// 删除验证码
	if err := redisClient.Del(ctx, codeKey).Err(); err != nil {
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
