package web

import (
	"xdream/logger"
	"encoding/json"
)

var AppConfig Config

type Config struct {
	AppLog  logger.LogConfig`json:"app_log"`
	CompLog map[string]*logger.CompLogger `json:"comp_log"`
}

func (cfg Config)String() string  {
	v, _ := json.MarshalIndent(cfg, "", "\t")
	return  string(v)
}