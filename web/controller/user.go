package controller

import (
	"github.com/gin-gonic/gin"
	metacontroller "meta/controller"
	"meta/error-code"
	"meta/response"
)

type UserController struct {
	metacontroller.Controller
}

func (c *UserController) PostLogin(ctx *gin.Context) {

	response.NewResponse(ctx, metaerrorcode.Success)
}
