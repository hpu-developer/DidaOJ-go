package controller

import (
	foundationerrorcode "foundation/error-code"
	foundationauth "foundation/foundation-auth"
	foundationservice "foundation/foundation-service"
	"github.com/gin-gonic/gin"
	metacontroller "meta/controller"
	"meta/error-code"
	"meta/response"
	weberrorcode "web/error-code"
	"web/request"
)

type UserController struct {
	metacontroller.Controller
}

func (c *UserController) PostLoginRefresh(ctx *gin.Context) {
	userId, err := foundationauth.GetUserIdFromContext(ctx)
	if err != nil {
		response.NewResponseError(ctx, err, nil)
		return
	}
	loginResponse, err := foundationservice.GetUserService().GetUserLoginResponse(ctx, userId)
	if err != nil {
		response.NewResponseError(ctx, err, nil)
		return
	}
	if loginResponse == nil {
		response.NewResponse(ctx, weberrorcode.WebErrorCodeUerNotMatch, nil)
		return
	}
	response.NewResponse(ctx, metaerrorcode.Success, loginResponse)
}

func (c *UserController) PostLogin(ctx *gin.Context) {
	var userLoginRequest request.UserLogin
	if err := ctx.ShouldBindJSON(&userLoginRequest); err != nil {
		response.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	if userLoginRequest.Username == "" || userLoginRequest.Password == "" {
		response.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	loginResponse, err := foundationservice.GetUserService().Login(ctx, userLoginRequest.Username, userLoginRequest.Password)
	if err != nil {
		response.NewResponseError(ctx, err, nil)
		return
	}
	if loginResponse == nil {
		response.NewResponse(ctx, weberrorcode.WebErrorCodeUerNotMatch, nil)
		return
	}
	response.NewResponse(ctx, metaerrorcode.Success, loginResponse)
}
