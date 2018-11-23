package web

import (
	"github.com/kataras/iris"
	"github.com/kataras/iris/mvc"
)

type  Controller struct {
	Ctx iris.Context
}

func (c *Controller)ObjectResponse(data interface{}) ( mvc.Result) {
	return mvc.Response{
		Object:data,
	}
}

func (c *Controller)WriteObject(obj interface{})  {
	c.Ctx.JSON(obj)
}
