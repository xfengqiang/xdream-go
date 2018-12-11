package user

import (
	"xdream/web"
	"xdream/example/webserver/data/api"
	"xdream/logger"
)



func GetUserInfo(ctx *web.Context)  {
	id := ctx.Params().Get("id")
	logger.SInfof(ctx.LogContext(), "request for user info id=%s", id)
	userInfo := api.UserInfo{}
	userInfo.User.Id = id
	userInfo.User.Name = "fank"
	ctx.JSON(userInfo)
}