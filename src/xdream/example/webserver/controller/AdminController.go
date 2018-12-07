package controller

import (
	"xdream/web"
	"fmt"
	"xdream/logger"
	"github.com/kataras/iris/context"
)


func init()  {
	web.App.Get("/admin/log/level", func(ctx context.Context) {
		ctx.JSON(struct {
			web.RespInfo
			Data map[string]string `json:"data"`
		}{
			web.RespInfo{},
			logger.GetLevelStatus(),
		})
	})

	web.App.Post("/admin/log/level/reset/{key}", func(ctx context.Context) {
		key := ctx.Params().Get("key")
		ret := logger.ResetLevel(key)
		ctx.JSON(web.RespInfo{
			Code:0,
			Msg:fmt.Sprintln("reset ret", ret),
		})
	})

	web.App.Post("/admin/log/level/{key}/{level}", func(ctx context.Context) {
		key := ctx.Params().Get("key")
		level := ctx.Params().Get("level")

		if _, ok := logger.LevelMap[level]; ok || level=="disable"{
			ret := logger.SetLevel(key, level)
			ctx.JSON(web.RespInfo{
				Code:0,
				Msg:fmt.Sprintln("set ret",ret),
			})
		}else{
			ctx.JSON(web.RespInfo{
				Code:-1,
				Msg:fmt.Sprintf("wrong level="+level),
			})
		}
	})
}

