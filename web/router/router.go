package router

import (
	foundationrouter "foundation/router"
	"github.com/gin-gonic/gin"
	metahttp "meta/meta-http"
	"web/controller"
)

func RegisterRoutes(r *gin.Engine) {
	metahttp.AuthMiddleware = foundationrouter.TokenAuthMiddleware()

	metahttp.AutoRegisterRoute(r, "/", new(controller.HomeController), metahttp.AuthMiddlewareTypeNone)
	metahttp.AutoRegisterRoute(r, "/problem", new(controller.ProblemController), metahttp.AuthMiddlewareTypeNone)
	metahttp.AutoRegisterRoute(r, "/judge", new(controller.JudgeController), metahttp.AuthMiddlewareTypeNone)
}
