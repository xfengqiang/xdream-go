package controller

import (
	//"github.com/kataras/iris/mvc"
	"github.com/kataras/iris"
	"github.com/kataras/iris/mvc"
	"time"
)
type  BaseController struct {
	Ctx iris.Context
	StartTime time.Time
}

func (c *BaseController)ObjectResponse(data interface{}) ( mvc.Result) {
	return mvc.Response{
		Object:data,
	}
}

func (c *BaseController)WriteObject(obj interface{})  {
	c.Ctx.JSON(obj)
}

// BeforeActivation called once before the server start
// and before the controller's registration, here you can add
// dependencies, to this controller and only, that the main caller may skip.
func (c *BaseController) BeforeActivation(b mvc.BeforeActivation) {

}
