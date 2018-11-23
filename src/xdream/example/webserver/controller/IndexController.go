package controller

import (
	"github.com/kataras/iris/mvc"
	"xdream/example/webserver/data/api"
	"xdream/web"
	"fmt"
	"github.com/kataras/iris"
	"xdream/example/webserver/middleware"
)

type IndexController struct {
	BaseController
	cnt int
}

func init()  {
	web.RegisterController(new(IndexController), "", middleware.PerControllerHandler)
}

// BeforeActivation called once before the server start
// and before the controller's registration, here you can add
// dependencies, to this controller and only, that the main caller may skip.
func (c *IndexController) AfterActivation(b mvc.AfterActivation) {
	// select the route based on the method name you want to
	// modify.
	index := b.GetRoute("GetHello")
	// just prepend the handler(s) as middleware(s) you want to use.
	// or append for "done" handlers.
	index.Handlers = append([]iris.Handler{middleware.PerActionHandler}, index.Handlers...)

	fmt.Println("============AfterActivation ")
	//log.Printf("request url:%s timeUsed:%v \n", c.Ctx.Request().RequestURI, time.Since(c.StartTime))
}


func (c *IndexController)GetHello() ( mvc.Result) {
	c.Ctx.Application().Logger().Infof("action execute")
	res := api.UserInfo{}
	res.RespInfo.Code = 100
	res.RespInfo.Msg = fmt.Sprintf("ok:%d", c.cnt)
	res.User.Id = 1
	res.User.Name = "fank xu"
	c.cnt++
	return c.ObjectResponse(res)
}

func (c *IndexController)GetPanic()  {
	panic("abc")
}