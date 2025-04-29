package controller

import (
	foundationerrorcode "foundation/error-code"
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
