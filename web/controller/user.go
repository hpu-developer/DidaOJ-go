package controller

import (
	"fmt"
	foundationerrorcode "foundation/error-code"
	foundationauth "foundation/foundation-auth"
	foundationmodel "foundation/foundation-model"
	foundationservice "foundation/foundation-service"
	"github.com/gin-gonic/gin"
	metacontroller "meta/controller"
	"meta/error-code"
	metaemail "meta/meta-email"
	metamath "meta/meta-math"
	"meta/meta-response"
	"strconv"
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
	if requestData.Email == "" {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	// 判断email是否格式有效

	confirmKey := strconv.Itoa(metamath.GetRandomInt(100000, 999999))

	subject := fmt.Sprintf("[DidaOJ] - 邮件验证码")
	body := fmt.Sprintf("%s：\n您好！\n欢迎您使用DidaOJ，以下是您的邮箱验证码：\n\n%s\n\n本验证码用于您注册本系统的账号，请勿泄露给他人。\n请在10分钟之内使用本验证码，过期请重新申请。\n如有疑问，请联系管理员。\n\n祝好！\nDidaOJ团队", requestData.Email, confirmKey)

	err := metaemail.SendEmail(config.GetConfig().Email.Email, config.GetConfig().Email.Password, config.GetConfig().Email.Host, config.GetConfig().Email.Port,
		requestData.Email, subject, body)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}

	metaresponse.NewResponse(ctx, metaerrorcode.Success, nil)
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
