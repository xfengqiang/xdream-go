package web

import (
	"github.com/kataras/iris"
	"reflect"
	"strings"
	"github.com/kataras/iris/mvc"
	"github.com/kataras/iris/context"
)

type controllerCfg struct {
	Controller interface{}
	Handlers []context.Handler
}
var (
	App *iris.Application = iris.New()
	routerConfig map[string]controllerCfg = map[string]controllerCfg{}
)

//func init()  {
//	App.Use(recover.New())
//}

func InitRoutes()  {
	for path, c := range routerConfig {
		mvc.Configure(App, func(application *mvc.Application) {
			groupRoute := application.Party(path)
			if len(c.Handlers) > 0 {
				groupRoute.Router.Use(c.Handlers...)
			}
			groupRoute.Handle(c.Controller)
		})
	}
}

func RegisterController(controller interface{}, path string, handlers ... context.Handler)  {
	//如果没有指定path，则使用controller的名字作为请求路径
	if path == "" {
		valueType := reflect.Indirect(reflect.ValueOf(controller)).Type()
		className := strings.TrimSuffix(valueType.Name(), "Controller")
		path = "/"+strings.ToLower(className[0:1]) + className[1:]
	}
	routerConfig[path] = controllerCfg{
		Controller:controller,
		Handlers:handlers,
	}
}

func RegisterGet(path string, fn func(*Context))  {
	App.Get(path, WrapHandler(fn))
}

func RegisterPost(path string, fn func(*Context))  {
	App.Post(path, WrapHandler(fn))
}

