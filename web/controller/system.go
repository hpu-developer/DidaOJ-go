package controller

import (
	foundationmodel "foundation/foundation-model"
	foundationservice "foundation/foundation-service"
	"github.com/gin-gonic/gin"
	metacontroller "meta/controller"
	"meta/error-code"
	metaerror "meta/meta-error"
	metapanic "meta/meta-panic"
	"meta/meta-response"
	metasystem "meta/meta-system"
	"time"
)

type SystemController struct {
	metacontroller.Controller
}

func (c *SystemController) GetStatus(ctx *gin.Context) {

	nowTime := time.Now()

	cpuUsage, err := metasystem.GetCpuUsage()
	if err != nil {
		metapanic.ProcessError(metaerror.Wrap(err, "get cpu usage failed"))
		return
	}
	memoryUsed, memoryTotal, err := metasystem.GetVirtualMemory()
	if err != nil {
		metapanic.ProcessError(metaerror.Wrap(err, "get virtual memory failed"))
		return
	}
	avgMessage, err := metasystem.GetAvgMessage()
	if err != nil {
		metapanic.ProcessError(metaerror.Wrap(err, "get avg message failed"))
		return
	}

	judgers, err := foundationservice.GetJudgerService().GetJudgerList(ctx)
	if err != nil {
		metapanic.ProcessError(metaerror.Wrap(err, "get judger list failed"))
		return
	}

	webStatus := foundationmodel.NewWebStatusBuilder().
		Name("DidaOJ").
		CpuUsage(cpuUsage).
		MemUsage(memoryUsed).
		MemTotal(memoryTotal).
		AvgMessage(avgMessage).
		UpdateTime(nowTime).
		Build()

	responseData := struct {
		Web    *foundationmodel.WebStatus `json:"web"`
		Judger []*foundationmodel.Judger  `json:"judger,omitempty"`
	}{
		Web:    webStatus,
		Judger: judgers,
	}

	metaresponse.NewResponse(ctx, metaerrorcode.Success, responseData)
}
