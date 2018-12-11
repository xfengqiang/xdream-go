package web

import (
	"github.com/kataras/iris"
	"sync"
	"xdream/logger"
)

// Context is our custom context.
// Let's implement a context which will give us access
// to the client's Session with a trivial `ctx.Session()` call.
type Context struct {
	iris.Context
	logCtx *logger.Context
}

// Session returns the current client's session.
func (ctx *Context) LogContext() *logger.Context {
	// this help us if we call `Session()` multiple times in the same handler
	if ctx.logCtx == nil {
		reqid := ctx.GetHeader("X-REQ-ID")
		logid := ctx.GetHeader("X-LOG-ID")
		refid := ctx.GetHeader("X-REF-ID")

		//TODO ID 自动生成
		ctx.logCtx = &logger.Context{
			RefId:refid,
			LogId:logid,
			ReqId:reqid,
		}
	}

	return ctx.logCtx
}


var contextPool = sync.Pool{New: func() interface{} {
	return &Context{}
}}

func acquire(original iris.Context) *Context {
	ctx := contextPool.Get().(*Context)
	ctx.Context = original // set the context to the original one in order to have access to iris's implementation.
	ctx.logCtx = nil      // reset the session
	return ctx
}

func release(ctx *Context) {
	contextPool.Put(ctx)
}

// Handler will convert our handler of func(*Context) to an iris Handler,
// in order to be compatible with the HTTP API.
func WrapHandler(h func(*Context)) iris.Handler {
	return func(original iris.Context) {
		ctx := acquire(original)

		h(ctx)
		release(ctx)
	}
}

