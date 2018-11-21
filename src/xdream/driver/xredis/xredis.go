package xredis

import (
	"errors"
	"fmt"
	"time"
	rds "github.com/gomodule/redigo/redis"
	"xdream/ns"
)

const (
	DNS_TYPE_IP = "ip"
)

var ErrNil error = rds.ErrNil

var (
	configMap map[string]RedisConfig = map[string]RedisConfig{}
	redisDBPool map[string]*rds.Pool =  map[string]*rds.Pool{}
	routeConfig ns.NsRoute //默认路由规则
	redisCallback RedisCallBack //redis 调用回调方法
)



type RedisCallBack func(c *RedisConn, timeused time.Duration, err error, cmd string, args ...interface{})
type RedisKeyGenerator func(c *RedisConfig, fmt string, args ...interface{}) string


var keyGenderator RedisKeyGenerator = func(c *RedisConfig, fmtStr string, args ...interface{}) string {

	if c != nil && c.Auth.Prefix != "" {
		return fmt.Sprintf(c.Auth.Prefix+":"+fmtStr, args...)
	}
	return fmt.Sprintf(fmtStr, args...)
}

func RegisterCallBack(callback RedisCallBack) {
	redisCallback = callback
}

func RegisterKeyGenerator(gen RedisKeyGenerator) {
	keyGenderator = gen
}

func SetRouteConfig(config ns.NsRoute)  {
	routeConfig = config
}

func GetRedis(name string, info interface{}) (conn rds.Conn, err error) {
	p, ok := redisDBPool[name]
	if !ok {
		err = errors.New("no such redis db")
		return
	}

	conn = p.Get()
	//conn.SetInfo(info)
	return
}


func GetRedisV2(name string, info interface{}, master bool) (conn rds.Conn, err error) {
	// slave 支持，去掉

	return  GetRedis(name, info)

	conf, ok := configMap[name]
	if(!ok) {
		err = errors.New(fmt.Sprintf("no such redis db:%s", name))
	}

	if !master && conf.Auth.SlaveName!= "" {
		name = conf.Auth.SlaveName
	}

	//fmt.Println("use redis", name)
	p, ok := redisDBPool[name]
	if !ok {
		err = errors.New(fmt.Sprintf("no such redis db:%s", name))
		return
	}

	conn = p.Get()
	//conn.SetInfo(info)
	return
}

func CreateRedisConn(name string) (con rds.Conn, err error) {
	cfg, ok := configMap[name]
	if !ok {
		err = errors.New(fmt.Sprintf("Reids not configed. [%s]", name))
	}

	nameservice := ns.NameService{}
	instances, e := nameservice.GetServices(cfg.Services, routeConfig)
	if e != nil {
		err = fmt.Errorf("get  host for redis [%s] failed.%s",  name, e.Error())
		return
	}

	if len(instances) == 0 {
		err = fmt.Errorf("no  host fund redis [%s] ",  name)
		return
	}

	ip, port := instances[0].Ip, instances[0].Port

	var options []rds.DialOption
	if cfg.Auth.ConnectTimeout > 0 {
		options = append(options, rds.DialConnectTimeout(time.Duration(cfg.Auth.ConnectTimeout)*time.Millisecond))
	}
	if cfg.Auth.ReadTimeout > 0 {
		options = append(options, rds.DialConnectTimeout(time.Duration(cfg.Auth.ReadTimeout)*time.Millisecond))
	}
	if cfg.Auth.WriteTimeout > 0 {
		options = append(options, rds.DialConnectTimeout(time.Duration(cfg.Auth.WriteTimeout)*time.Millisecond))
	}

	c, err := rds.Dial("tcp", fmt.Sprintf("%v:%v", ip, port), options...)

	if err != nil {
		return
	}

	cfg.ip = ip
	cfg.port = port

	con = &RedisConn{
		Conf: cfg,
		Conn: c,
	}
	return
}
