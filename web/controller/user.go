package controller

import (
	foundationerrorcode "foundation/error-code"
	foundationauth "foundation/foundation-auth"
	foundationmodel "foundation/foundation-model"
	foundationservice "foundation/foundation-service"
	"github.com/gin-gonic/gin"
	metacontroller "meta/controller"
	"meta/error-code"
	"meta/meta-response"
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
