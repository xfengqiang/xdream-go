package middleware

import (
	"github.com/kataras/iris/context"
	"time"
)

func PerActionHandler(ctx context.Context)  {
	startTime := time.Now()
	ctx.Application().Logger().Infof("[PerAction] request start url:%s at:%v",
		startTime)

	// for this specific request, then skip the whole cache
	bodyHandler := ctx.NextHandler()
	if bodyHandler == nil {
		return
	}
	bodyHandler(ctx)
	ctx.Application().Logger().Infof("[PerAction] request url:%s timeUsed:%v ",
		ctx.Request().RequestURI, time.Since(startTime))
}


func PerControllerHandler(ctx context.Context)  {
	startTime := time.Now()
	ctx.Application().Logger().Infof("[PerController] request start url:%s ",
		startTime)

	// for this specific request, then skip the whole cache
	bodyHandler := ctx.NextHandler()
	if bodyHandler == nil {
		return
	}
	bodyHandler(ctx)
	ctx.Application().Logger().Infof("[PerController] request url:%s timeUsed:%v ",
		ctx.Request().RequestURI, time.Since(startTime))
}

func GloabalHandler(ctx context.Context)  {

	startTime := time.Now()
	ctx.Application().Logger().Infof("[global ] request start url:%s ",
		startTime)
	ctx.Next()
	ctx.Application().Logger().Infof("[global ] request end url:%s timeUsed:%v ",
		ctx.Request().RequestURI, time.Since(startTime))
}
