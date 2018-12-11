package controller

import (
	"xdream/web"
	"xdream/example/webserver/controller/user"
)

func init()  {
	web.RegisterGet("/user/{id:string}", user.GetUserInfo)
}
