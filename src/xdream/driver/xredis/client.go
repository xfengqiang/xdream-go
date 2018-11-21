package xredis

import (
    rds "github.com/gomodule/redigo/redis"
    "time"
)

type RedisConn struct {
    Conf RedisConfig
    rds.Conn
}

func (this *RedisConn) Do(commandName string, args ...interface{}) (reply interface{}, err error) {
    if commandName == "" {
        return 
    }
    var startTime time.Time = time.Now()
    if redisCallback != nil {
        startTime = time.Now()
    }

    defer func(){
        if redisCallback != nil {
            timeUsed := time.Now().Sub(startTime)
            redisCallback(this, timeUsed, err, commandName, args...)
        }
    }()
    reply, err = this.Conn.Do(commandName, args...)
    return
}