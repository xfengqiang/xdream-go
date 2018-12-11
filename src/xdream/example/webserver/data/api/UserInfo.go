package api

import "xdream/web"


type UserInfo struct {
	web.RespInfo
	User struct{
		Id string `json:"id"`
		Name string `json:"name"`
	} `json:"user"`
}
