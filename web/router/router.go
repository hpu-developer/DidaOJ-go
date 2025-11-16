package router

import (
	foundationerrorcode "foundation/error-code"
	foundationauth "foundation/foundation-auth"
	foundationservice "foundation/foundation-service"
	foundationrouter "foundation/router"
	metahttp "meta/meta-http"
	metapanic "meta/meta-panic"
	metaresponse "meta/meta-response"
	"web/controller"

	"github.com/gin-gonic/gin"
)

func AuthCheckMiddleware(auth foundationauth.AuthType) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		metahttp.AuthMiddlewareOptional(ctx)
		if ctx.IsAborted() {
			return
		}
		userId, err := foundationauth.GetUserIdFromContext(ctx)
		if err != nil {
			metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
			ctx.Abort()
			return
		}
		ok, err := foundationservice.GetUserService().CheckUserAuthByUserId(ctx, userId, auth)
		if err != nil {
			metapanic.ProcessError(err)
			metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
			ctx.Abort()
			return
		}
		if !ok {
			metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
			ctx.Abort()
			return
		}
		// 已经通过之前的调用过Next了，这里不需要再调一次了
	}
}

func RegisterRoutes(r *gin.Engine) {
	metahttp.AuthMiddleware = foundationrouter.TokenAuthMiddleware()
	metahttp.AuthMiddlewareOptional = foundationrouter.AuthMiddlewareOptional()

	metahttp.AutoRegisterRoute(r, "/", new(controller.HomeController), metahttp.AuthMiddlewareTypeNone)
	metahttp.AutoRegisterRoute(r, "/problem", new(controller.ProblemController), metahttp.AuthMiddlewareTypeOptional)
	metahttp.AutoRegisterRoute(r, "/judge", new(controller.JudgeController), metahttp.AuthMiddlewareTypeOptional)
	metahttp.AutoRegisterRoute(r, "/user", new(controller.UserController), metahttp.AuthMiddlewareTypeOptional)
	metahttp.AutoRegisterRoute(r, "/contest", new(controller.ContestController), metahttp.AuthMiddlewareTypeOptional)
	metahttp.AutoRegisterRoute(
		r,
		"/collection",
		new(controller.CollectionController),
		metahttp.AuthMiddlewareTypeOptional,
	)
	metahttp.AutoRegisterRoute(r, "/discuss", new(controller.DiscussController), metahttp.AuthMiddlewareTypeOptional)
	metahttp.AutoRegisterRoute(r, "/rank", new(controller.RankController), metahttp.AuthMiddlewareTypeNone)
	metahttp.AutoRegisterRoute(r, "/system", new(controller.SystemController), metahttp.AuthMiddlewareTypeOptional)
	metahttp.AutoRegisterRoute(r, "/run", new(controller.RunController), metahttp.AuthMiddlewareTypeRequire)

}
