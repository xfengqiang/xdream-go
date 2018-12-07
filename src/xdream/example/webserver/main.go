package main

import (
	"github.com/kataras/iris"
	_ "xdream/example/webserver/controller"
	"xdream/web"
	"github.com/kataras/iris/context"
	"fmt"
	"strconv"
	"runtime"
	"xdream/example/webserver/middleware"
	"flag"
	"encoding/json"
	"io/ioutil"
	"xdream/xutil"
	"log"
	"xdream/logger"
)

func loadConfig(configPath string )  {
	content, err := ioutil.ReadFile(configPath)
	if err != nil {
		panic("read config error "+err.Error())
	}
	err = json.Unmarshal(content, &web.AppConfig)
	if err != nil  {
		panic("decode config file fail."+"see "+configPath+" " + xutil.JsonStringError(string(content), err))
	}

	log.Println(web.AppConfig)
}

func initLogger()  {
	logger.InitLogger(&web.AppConfig.AppLog)
	if lg, ok := web.AppConfig.CompLog["iris"]; ok {

		logComp := lg
		logComp.SetLevel(lg.Level) //这里需要初始化一下

		ls := logger.NewLevelSetter(lg.Level, func(level string) {
			// "disable"
			// "fatal"
			// "error"
			// "warn"
			// "info"
			// "debug"

			//同时设置logComp和web.App.Logger(),保证性能
			if level=="disable" {
				logComp.Enabled = false
			}else{
				logComp.Enabled = true
				lg.SetLevel(level)
			}
			web.App.Logger().SetLevel(level)
		})

		logger.RegisterLevelSetter("iris", ls)
		log.SetOutput(logComp)
	}
}
func main() {

	configPath := flag.String("c", "config.json", "path to config file")
	flag.Parse()

	//解析配置文件
	loadConfig(*configPath)
	//初始化日志组件
	initLogger()

	//自定义panic处理函数
	web.App.Use(myrecover())
	web.App.Use(middleware.GloabalHandler)

	//初始化路由配置，必须在app.Use后面执行，否则自定义panic处理函数不生效
	web.InitRoutes()
	web.App.Get("/panic", func(i context.Context) {
		panic("abc")
	})
	web.App.Run(iris.Addr("0.0.0.0:8080"), configure)
}

func getRequestLogs(ctx context.Context) string {
	var status, ip, method, path string
	status = strconv.Itoa(ctx.GetStatusCode())
	path = ctx.Path()
	method = ctx.Method()
	ip = ctx.RemoteAddr()
	// the date should be logged by iris' Logger, so we skip them
	return fmt.Sprintf("%v %s %s %s", status, path, method, ip)
}


// New returns a new recover middleware,
// it recovers from panics and logs
// the panic message to the application's logger "Warn" level.
func myrecover() context.Handler {
	return func(ctx context.Context) {
		defer func() {
			if err := recover(); err != nil {
				if ctx.IsStopped() {
					return
				}

				var stacktrace string
				for i := 1; ; i++ {
					_, f, l, got := runtime.Caller(i)
					if !got {
						break

					}

					stacktrace += fmt.Sprintf("%s:%d\n", f, l)
				}

				// when stack finishes
				logMessage := fmt.Sprintf("Recovered from a route's Handler('%s')\n", ctx.HandlerName())
				logMessage += fmt.Sprintf("At Request: %s\n", getRequestLogs(ctx))
				logMessage += fmt.Sprintf("Trace: %s\n", err)
				logMessage += fmt.Sprintf("\n%s", stacktrace)
				ctx.Application().Logger().Warn(logMessage)

				ctx.StatusCode(500)
				ctx.StopExecution()
			}
		}()

		ctx.Next()
	}
}


func configure(app *iris.Application) {
	app.Configure(
		iris.WithoutServerError(iris.ErrServerClosed),
	)
}
