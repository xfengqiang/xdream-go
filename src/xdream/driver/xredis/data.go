package xredis


import (
    "xdream/ns"
    "fmt"
)

type RedisConfig struct {
    Auth struct{
        Name string `json:"string"`
        MaxIdle int `json:"max_idle"`
        Prefix string `json:"prefix"`

        IdleTimeout int `json:"idle_timeout"` //单位毫秒, 默认0
        ConnectTimeout int `json:"connect_timeout"` //单位毫秒，默认2000, 小于0时不生效
        ReadTimeout int `json:"read_timeout"` //单位毫秒，默认2000, 小于0时不生效
        WriteTimeout int `json:"write_timeout"`//单位毫秒，默认2000, 小于0时不生效

        SlaveName string `json:"slave_name"`
        //MultiWrite []string
    }
    Services ns.NsConfig

    ip string
    port int
}

func (c *RedisConfig)GetInfo() string  {
    return  fmt.Sprintf("name:%s ip:%s port:%d",c.Auth.Name, c.ip, c.port)
}




