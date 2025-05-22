package controller

import (
	"errors"
	"fmt"
	foundationerrorcode "foundation/error-code"
	foundationauth "foundation/foundation-auth"
	foundationmodel "foundation/foundation-model"
	foundationservice "foundation/foundation-service"
	foundationuser "foundation/foundation-user"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	metacontroller "meta/controller"
	"meta/error-code"
	metaemail "meta/meta-email"
	metamath "meta/meta-math"
	metaredis "meta/meta-redis"
	"meta/meta-response"
	metatime "meta/meta-time"
	"strconv"
	"time"
	"web/config"
	weberrorcode "web/error-code"
	"web/request"
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
	userInfo, err := foundationservice.GetUserService().GetInfo(ctx, username)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	if userInfo == nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.NotFound, nil)
		return
	}
	acProblems, err := foundationservice.GetJudgeService().GetUserAcProblemIds(ctx, userInfo.Id)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}

	responseData := struct {
		User       *foundationmodel.UserInfo `json:"user"`
		ProblemsAc []string                  `json:"problems_ac"`
	}{
		User:       userInfo,
		ProblemsAc: acProblems,
	}
	metaresponse.NewResponse(ctx, metaerrorcode.Success, responseData)
}

func (c *UserController) PostRegisterEmail(ctx *gin.Context) {
	var requestData struct {
		Email string `json:"email" binding:"required,email"`
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
	body := fmt.Sprintf("%s：\n您好！\n欢迎您使用DidaOJ，以下是您的邮箱验证码：\n\n%s\n\n本验证码用于您注册本系统的账号，请勿泄露给他人。\n请在10分钟之内使用本验证码，过期请重新申请。\n如有疑问，请联系管理员。\n\n祝好！\nDidaOJ团队",
		email, code)

	err = metaemail.SendEmail(config.GetConfig().Email.Email, config.GetConfig().Email.Password, config.GetConfig().Email.Host, config.GetConfig().Email.Port,
		requestData.Email, subject, body)
	if err != nil {
		metaresponse.NewResponse(ctx, weberrorcode.RegisterMailSendFail, nil)
		return
	}

	metaresponse.NewResponse(ctx, metaerrorcode.Success)
}

func (c *UserController) PostRegister(ctx *gin.Context) {
	var requestData struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
		Email    string `json:"email" binding:"required"`
		Key      string `json:"key" binding:"required"`
		Nickname string `json:"nickname"`
	}
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	if requestData.Username == "" || requestData.Password == "" || requestData.Email == "" {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	// 判断username是否>3并且<20
	if len(requestData.Username) < 3 || len(requestData.Username) > 20 {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	// 判断username是否仅包含字母数字下划线
	if !foundationuser.IsValidUsername(requestData.Username) {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	// 判断password是否>6并且<20
	if len(requestData.Password) < 6 || len(requestData.Password) > 20 {
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

	user := foundationmodel.NewUserBuilder().
		Username(requestData.Username).
		Password(passwordEncode).
		Email(requestData.Email).
		Nickname(requestData.Nickname).
		RegTime(metatime.GetTimeNow()).
		Build()

	err = foundationservice.GetUserService().InsertUser(ctx, user)
	if err != nil {
		metaresponse.NewResponse(ctx, weberrorcode.RegisterUserFail)
		return
	}

	metaresponse.NewResponse(ctx, metaerrorcode.Success)
}

func (c *UserController) PostLoginRefresh(ctx *gin.Context) {
	userId, err := foundationauth.GetUserIdFromContext(ctx)
	if err != nil {
		metaresponse.NewResponseError(ctx, err, nil)
		return
	}
	loginResponse, err := foundationservice.GetUserService().GetUserLoginResponse(ctx, userId)
	if err != nil {
		metaresponse.NewResponseError(ctx, err, nil)
		return
	}
	if loginResponse == nil {
		metaresponse.NewResponse(ctx, weberrorcode.UserNotMatch, nil)
		return
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
	loginResponse, err := foundationservice.GetUserService().Login(ctx, userLoginRequest.Username, userLoginRequest.Password)
	if err != nil {
		metaresponse.NewResponseError(ctx, err, nil)
		return
	}
	if loginResponse == nil {
		metaresponse.NewResponse(ctx, weberrorcode.UserNotMatch, nil)
		return
	}
	metaresponse.NewResponse(ctx, metaerrorcode.Success, loginResponse)
}
