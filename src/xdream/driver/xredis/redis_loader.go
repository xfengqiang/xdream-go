package xredis

import (
	rds "github.com/gomodule/redigo/redis"
	"time"
)

//加载数据库配置
func InitWithConfig(configs map[string]RedisConfig) (err error) {
	for name, cfg := range configs {
		cfg.Auth.Name = name

		if cfg.Auth.ReadTimeout == 0 {
			cfg.Auth.ReadTimeout = 2000
		}
		if cfg.Auth.WriteTimeout == 0 {
			cfg.Auth.WriteTimeout = 2000
		}
		if cfg.Auth.ConnectTimeout == 0 {
			cfg.Auth.ConnectTimeout = 2000
		}

		configMap[name] = cfg

		pool := &rds.Pool{
			MaxIdle: cfg.Auth.MaxIdle,
			IdleTimeout: time.Duration(cfg.Auth.IdleTimeout) * time.Millisecond,
			Dial: func () (rds.Conn, error) { return CreateRedisConn(name) },
		}

		redisDBPool[name] = pool
	}
	return
}