package api

import "xdream/web"


type UserInfo struct {
	web.RespInfo
	User struct{
		Id int `json:"id"`
		Name string `json:"name"`
	} `json:"user"`
}
