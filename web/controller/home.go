package controller

import (
	"github.com/gin-gonic/gin"
	metacontroller "meta/controller"
	"meta/error-code"
	metaconfig "meta/meta-config"
	"meta/meta-response"
	"meta/meta-time"
	webconfig "web/config"
)

type HomeController struct {
	metacontroller.Controller
}

type HomeData struct {
	Port    int32  `json:"port"`
	Time    string `json:"time"`
	Powered string `json:"powered"`
	Module  string `json:"module"`
	Name    string `json:"name"`
}

func (c *HomeController) Get(ctx *gin.Context) {
	////获取当前时间
	metaresponse.NewResponse(
		ctx, metaerrorcode.Success, HomeData{
			Port:    webconfig.GetHttpPort(),
			Time:    metatime.GetTimeNowString(),
			Powered: "Golang",
			Module:  metaconfig.GetModuleName(),
			Name:    metaconfig.GetNodeName(),
		},
	)
}
