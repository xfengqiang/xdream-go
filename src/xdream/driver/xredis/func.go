package xredis

import (
    rds "github.com/gomodule/redigo/redis"
    "time"
    "fmt"
)
func Str2Interface(keys []string) []interface{}{
    var ret []interface{} = []interface{}{}
    for _, k := range keys {
        ret = append(ret, k)
    }
    return ret
}

func buildKey(dbname string, keyfmt string, args ...interface{}) string{
    if conf, ok :=configMap[dbname]; ok {
        return keyGenderator(&conf, keyfmt, args...)
    }
    return keyGenderator(nil, keyfmt, args...)
}

func simpleKey(dbname string, keyfmt string) string{
    if conf, ok :=configMap[dbname]; ok {
        return keyGenderator(&conf, keyfmt)
    }
    return keyGenderator(nil, keyfmt)
}

func simpleKeys(dbname string, keys ...interface{}) []interface{}{
    ret := make([]interface{}, 0, len(keys))
    conf, ok :=configMap[dbname]
    for _, k := range keys {
        if ok {
            ret =append(ret, keyGenderator(&conf, fmt.Sprintf("%v", k)))
        }else{
            ret =append(ret, keyGenderator(nil, fmt.Sprintf("%v", k)))
        }
    }
    return ret
}

func buildMultiKeys(dbname string, key string, args ...interface{}) []string{
    ret := []string{}
    conf, _ :=configMap[dbname];
    for _, v := range args {
        ret = append(ret, keyGenderator(&conf, key, v))
    }
    return ret
}

func Expire(dbname string, info interface{}, key string,  seconds int) (err error) {
    redis, err := GetRedisV2(dbname, info, true)
    if err != nil {
        return
    }
    defer redis.Close()
    key = simpleKey(dbname, key)
    _, err = rds.Int(redis.Do("EXPIRE", key, seconds))
    return
}

func Set(dbname string, info interface{},key string, value string, options ...interface{}) (err error){
    redis, err := GetRedisV2(dbname, info, true)
    if err != nil {
        return
    }
    defer redis.Close()
    key = simpleKey(dbname, key)
    params := []interface{}{key, value}
    if options != nil {
        if exptime, ok := options[0].(time.Duration); ok {
            params = append(params, "px", int64(exptime/time.Millisecond))
        }
    }

    _, err = redis.Do("SET", params...)
    return
}

//value 为整形或者字符串，不能为struct或map
func Setex(dbname string, info interface{}, key string, seconds int, value interface{}) (err error) {
    redis, err := GetRedisV2(dbname, info, true)
    if err != nil {
        return
    }
    defer redis.Close()
    key = simpleKey(dbname, key)
    _, err = redis.Do("SETEX", key, seconds, value)
    return
}

func Setnx(dbname string, info interface{},key string, value interface{}) (ret int, err error){
    redis, err := GetRedisV2(dbname, info, true)
    if err != nil {
        return
    }
    defer redis.Close()
    key = simpleKey(dbname, key)
    return rds.Int(redis.Do("SETNX", key, value))
}

func Get(dbname string,info interface{},key string) (string, error) {
    redis, err := GetRedisV2(dbname, info, false)
    if err != nil {
        return "", err
    }
    defer redis.Close()
    key = simpleKey(dbname, key)
    return rds.String(redis.Do("GET", key))
}

func Ttl(dbname string, info interface{}, key string) (int64, error) {
    redis, err := GetRedisV2(dbname, info, false)
    if err != nil {
        return 0, err
    }
    defer redis.Close()
    key = simpleKey(dbname, key)
    return rds.Int64(redis.Do("TTL", key))
}

func Key(fmt string, args...interface{}) string {
    return keyGenderator(nil, fmt, args...)
}

func Del(dbname string, info interface{}, keys ...interface{}) (err error) {
    redis, err := GetRedisV2(dbname, info, true)
    if err != nil {
        return
    }
    defer redis.Close()
    keys = simpleKeys(dbname, keys...)
    _, err =  redis.Do("DEL", keys...)
    return
}

//===================================================
func Incr(dbname string, info interface{},key string) (int int, err error){
    redis, err := GetRedisV2(dbname, info, true)
    if err != nil {
        return
    }
    defer redis.Close()
    key = simpleKey(dbname, key)
    return  rds.Int(redis.Do("INCR", key))
}

func IncrBy(dbname string, info interface{},key string, v int) (int int, err error){
    redis, err := GetRedisV2(dbname, info, true)
    if err != nil {
        return
    }
    defer redis.Close()
    key = simpleKey(dbname, key)
    return rds.Int(redis.Do("INCRBY", key, v))
}

func Decr(dbname string, info interface{}, key string) (int int, err error){
    redis, err := GetRedisV2(dbname, info, true)
    if err != nil {
        return
    }
    defer redis.Close()
    key = simpleKey(dbname, key)
    return rds.Int(redis.Do("DECR", key))
}

func DecrBy(dbname string, info interface{},key string, v int) (int int, err error){
    redis, err := GetRedisV2(dbname, info, true)
    if err != nil {
        return
    }
    defer redis.Close()
    key = simpleKey(dbname, key)
    return rds.Int(redis.Do("DECRBY", key, v))
}

//===================================================
func LPush(dbname string, info interface{},key string, values ...interface{}) (err error){
    redis, err := GetRedisV2(dbname, info, true)
    if err != nil {
        return
    }
    defer redis.Close()
    key = simpleKey(dbname, key)
    err = execWrite(redis, "LPUSH", key, values...)
    return
}

func RPush(dbname string, info interface{},key string, values ...interface{}) (err error){
    redis, err := GetRedisV2(dbname, info, true)
    if err != nil {
        return
    }
    defer redis.Close()
    key = simpleKey(dbname, key)
    err = execWrite(redis, "RPUSH", key, values...)
    return
}

func LPushx(dbname string, info interface{},key string, values ...interface{}) (err error){
    redis, err := GetRedisV2(dbname, info, true)
    if err != nil {
        return
    }
    defer redis.Close()
    key = simpleKey(dbname, key)
    err = execWrite(redis, "LPUSHX", key, values...)
    return
}

func RPushx(dbname string, info interface{},key string, values ...interface{}) (err error){
    redis, err := GetRedisV2(dbname, info, true)
    if err != nil {
        return
    }
    defer redis.Close()
    key = simpleKey(dbname, key)
    err = execWrite(redis, "RPUSHX", key, values...)
    return
}

func LPop(dbname string, info interface{},key string) (ret string, err error){
    redis, err := GetRedisV2(dbname, info, true)
    if err != nil {
        return
    }
    defer redis.Close()
    key = simpleKey(dbname, key)
    return rds.String(redis.Do("LPOP", key))
}

func RPop(dbname string, info interface{},key string) (ret string, err error){
    redis, err := GetRedisV2(dbname, info, true)
    if err != nil {
        return
    }
    defer redis.Close()
    key = simpleKey(dbname, key)
    return rds.String(redis.Do("RPOP", key))
}

func BLPop(dbname string, info interface{},key string) (ret string, err error){
    redis, err := GetRedisV2(dbname, info, true)
    if err != nil {
        return
    }
    defer redis.Close()
    key = simpleKey(dbname, key)
    return rds.String(redis.Do("BLPOP", key))
}

func BRPop(dbname string, info interface{},key string) (ret string, err error){
    redis, err := GetRedisV2(dbname, info, true)
    if err != nil {
        return
    }
    defer redis.Close()
    key = simpleKey(dbname, key)
    return rds.String(redis.Do("BRPOP", key))
}

func LLen(dbname string, info interface{},key string) (ret int, err error){
    redis, err := GetRedisV2(dbname, info, false)
    if err != nil {
        return
    }
    defer redis.Close()
    key = simpleKey(dbname, key)
    return rds.Int(redis.Do("LLEN", key))
}

func LIndex(dbname string, info interface{},key string, idx int) (ret string, err error){
    redis, err := GetRedisV2(dbname, info, false)
    if err != nil {
        return
    }
    defer redis.Close()
    key = simpleKey(dbname, key)
    params := []interface{}{key, idx}
    return rds.String(redis.Do("LINDEX", params...))
}

func LSet(dbname string, info interface{},key string, idx int, value interface{}) (ret error){
    redis, err := GetRedisV2(dbname, info, true)
    if err != nil {
        return
    }
    defer redis.Close()
    key = simpleKey(dbname, key)
    return execWrite(redis, "LSET", key, idx, value)
}

func LRange(dbname string, info interface{},key string, start int, end int) (ret []string,err error){
    redis, err := GetRedisV2(dbname, info, false)
    if err != nil {
        return
    }
    defer redis.Close()
    key = simpleKey(dbname, key)
    params := []interface{}{key, start, end}
    return rds.Strings(redis.Do("LRANGE", params...))
}

func LTrim(dbname string, info interface{},key string, start int, end int) (ret []string, err error){
    redis, err := GetRedisV2(dbname, info, true)
    if err != nil {
        return
    }
    defer redis.Close()
    key = simpleKey(dbname, key)
    params := []interface{}{key, start, end}
    return  rds.Strings(redis.Do("LTRIM", params...))
}

func LRem(dbname string, info interface{},key string, start int, count int) (ret []string,err error){
    redis, err := GetRedisV2(dbname, info, true)
    if err != nil {
        return
    }
    defer redis.Close()
    key = simpleKey(dbname, key)
    params := []interface{}{key, start, count}
    return rds.Strings(redis.Do("LREM", params...))
}

//====================================================================
//set
func SAdd(dbname string, info interface{},key string, values...interface{}) (v int64, err error){
    redis, err := GetRedisV2(dbname, info, true)
    if err != nil {
        return
    }
    defer redis.Close()
    key = simpleKey(dbname, key)
    reply, err := execWrite2(redis,"SADD", key, values...)
	if err != nil {
		return 
	}
	v = reply.(int64)
	return 
}

func SRem(dbname string, info interface{},key string, members...interface{}) (err error){
    redis, err := GetRedisV2(dbname, info, true)
    if err != nil {
        return
    }
    defer redis.Close()
    key = simpleKey(dbname, key)
    return execWrite(redis,"SREM", key, members...)
}

func SIsMember(dbname string, info interface{},key string, member interface{}) (ret bool, err error){
    redis, err := GetRedisV2(dbname, info, false)
    if err != nil {
        return
    }
    defer redis.Close()
    key = simpleKey(dbname, key)
    params := []interface{}{key, member}
    return rds.Bool(redis.Do("SISMEMBER", params...))
}

func SCard(dbname string, info interface{},key string) (ret int, err error){
    redis, err := GetRedisV2(dbname, info, false)
    if err != nil {
        return
    }
    defer redis.Close()
    key = simpleKey(dbname, key)
    return rds.Int(redis.Do("SCARD", key))
}

func SPop(dbname string, info interface{},key string) (ret string, err error){
    redis, err := GetRedisV2(dbname, info, true)
    if err != nil {
        return
    }
    defer redis.Close()
    key = simpleKey(dbname, key)
    return rds.String(redis.Do("SPOP", key))
}

func SRandMember(dbname string, info interface{},key string, arg ...interface{}) (ret string, err error){
    redis, err := GetRedisV2(dbname, info, false)
    if err != nil {
        return
    }
    defer redis.Close()
	key = simpleKey(dbname, key)
	as := []interface{}{key}
	if len(arg) > 0 {
		as = append(as, arg...)
	}
    return rds.String(redis.Do("SRANDMEMBER",  as...))
}

func SInter(dbname string, info interface{},keys ...interface{}) (ret []string, err error){
    redis, err := GetRedisV2(dbname, info, false)
    if err != nil {
        return
    }
    defer redis.Close()
    keys = simpleKeys(dbname, keys...)
    return rds.Strings(redis.Do("SINTER", keys...))
}

func SUnion(dbname string, info interface{},keys ...interface{}) (ret []string, err error){
    redis, err := GetRedisV2(dbname, info, false)
    if err != nil {
        return
    }
    defer redis.Close()
    keys = simpleKeys(dbname, keys...)
    return rds.Strings(redis.Do("SUNION", keys...))
}

func SDiff(dbname string, info interface{},keys ...interface{}) (ret []string, err error){
    redis, err := GetRedisV2(dbname, info, false)
    if err != nil {
        return
    }
    defer redis.Close()
    keys = simpleKeys(dbname, keys...)
    return rds.Strings(redis.Do("SDIFF", keys...))
}

func SMembers(dbname string, info interface{},key string) (ret []string, err error){
    redis, err := GetRedisV2(dbname, info, false)
    if err != nil {
        return
    }
    defer redis.Close()
    key = simpleKey(dbname, key)
    return rds.Strings(redis.Do("SMEMBERS", key))
}

func SScan(dbname string, info interface{},key string,startId, count int ) (nextId int, ret []string, err error){
	redis, err := GetRedisV2(dbname, info, false)
	if err != nil {
		return
	}
	defer redis.Close()
	key = simpleKey(dbname, key)
	values, err := rds.Values(redis.Do("SSCAN", key, startId, "COUNT", count))
	if(err != nil ) {
		return 
	}
	nextId, _ = rds.Int(values[0], err)
	ret, _ = rds.Strings(values[1], err)
	return 
}


//====================================================================
//zset
func ZAdd(dbname string, info interface{},key string, values...interface{}) (err error){
    redis, err := GetRedisV2(dbname, info, true)
    if err != nil {
        return
    }
    defer redis.Close()
    key = simpleKey(dbname, key)
    return execWrite(redis, "ZADD", key, values...)
}

func ZRem(dbname string, info interface{},key string, members...interface{}) (err error){
    redis, err := GetRedisV2(dbname, info, true)
    if err != nil {
        return
    }
    defer redis.Close()
    key = simpleKey(dbname, key)
    return execWrite(redis, "ZREM", key, members...)
}

func ZRange(dbname string, info interface{},key string, start interface{}, end interface{}) (ret []string, err error) {
    redis, err := GetRedisV2(dbname, info, false)
    if err != nil {
        return
    }
    defer redis.Close()
    key = simpleKey(dbname, key)
    params := []interface{}{key, start, end}
    return rds.Strings(redis.Do("ZRANGE", params...))
}

func ZRangeWithScore(dbname string, info interface{},key string, start interface{}, end interface{}) (ret []string, err error) {
	redis, err := GetRedisV2(dbname, info, false)
	if err != nil {
		return
	}
	defer redis.Close()
	key = simpleKey(dbname, key)
	params := []interface{}{key, start, end, "WITHSCORES"}
	return rds.Strings(redis.Do("ZRANGE", params...))
}

func ZRevRange(dbname string, info interface{},key string, start interface{}, end interface{}) (ret []string, err error) {
    redis, err := GetRedisV2(dbname, info, false)
    if err != nil {
        return
    }
    defer redis.Close()
    key = simpleKey(dbname, key)
    return rds.Strings(redis.Do("ZREVRANGE", key, start, end))
}

func ZRevRangeWithScore(dbname string, info interface{},key string, start interface{}, end interface{}) (ret []string, err error) {
	redis, err := GetRedisV2(dbname, info, false)
	if err != nil {
		return
	}
	defer redis.Close()
	key = simpleKey(dbname, key)
	return rds.Strings(redis.Do("ZREVRANGE", key, start, end, "WITHSCORES"))
}

func ZRangeByScore(dbname string, info interface{},key string, start interface{}, end interface{}, withScore bool) (ret interface{}, err error) {
    redis, err := GetRedisV2(dbname, info, false)
    if err != nil {
        return
    }
    defer redis.Close()
    key = simpleKey(dbname, key)
    return zRangeByScore(redis, "ZRANGEBYSCORE", key, start, end, withScore)
}

func ZRevRangeByScore(dbname string, info interface{},key string, start interface{}, end interface{}, withScore bool) (ret interface{}, err error) {
    redis, err := GetRedisV2(dbname, info, false)
    if err != nil {
        return
    }
    defer redis.Close()
    key = simpleKey(dbname, key)
    return zRangeByScore(redis, "ZREVRANGEBYSCORE", key, start, end, withScore)
}


func ZRevRangeByScorePage(dbname string, info interface{},key string, start interface{}, end interface{},  offset int64, count int) (ret []string, err error) {
	redis, err := GetRedisV2(dbname, info, false)
	if err != nil {
		return
	}
	defer redis.Close()
	key = simpleKey(dbname, key)

	ret , err =  rds.Strings(redis.Do("ZREVRANGEBYSCORE", key, start, end, "WITHSCORES", "LIMIT", offset, count))

	return
}

func ZRangeByScorePage(dbname string, info interface{},key string, start interface{}, end interface{},  offset int64, count int) (ret []string, err error) {
	redis, err := GetRedisV2(dbname, info, false)
	if err != nil {
		return
	}
	defer redis.Close()
	key = simpleKey(dbname, key)

	ret , err =  rds.Strings(redis.Do("ZRANGEBYSCORE", key, start, end, "WITHSCORES", "LIMIT", offset, count))

	return
}

func zRangeByScore(redis rds.Conn, cmd string, key string, start interface{}, end interface{}, withScore bool) (interface{}, error) {
    if !withScore {
        ret, err := rds.Strings(redis.Do(cmd, key, start, end))
        return ret, err
    }

    values, err :=  rds.Strings(redis.Do(cmd, key, start, end, "WITHSCORES"))
    if err != nil {
        return nil, err
    }

    ret := map[string]string{}
    for i:=0; i<len(values); i+=2{
        ret[values[i]] = values[i+1]
    }
    return ret, err
}
func ZCard(dbname string, info interface{},key string) (ret int, err error) {
    redis, err := GetRedisV2(dbname, info, false)
    if err != nil {
        return
    }
    defer redis.Close()
    key = simpleKey(dbname, key)
    return rds.Int(redis.Do("ZCARD",key))
}

func ZCount(dbname string, info interface{},key string, start, end interface{}) (ret int, err error) {
    redis, err := GetRedisV2(dbname, info, false)
    if err != nil {
        return
    }
    defer redis.Close()
    key = simpleKey(dbname, key)
    return rds.Int(redis.Do("ZCOUNT", key, start, end))
}

func ZRemRangeByScore(dbname string, info interface{},key string, start, end interface{}) (err error){
    redis, err :=  GetRedisV2(dbname, info, true)
    if err != nil {
        return
    }
    defer redis.Close()
    key = simpleKey(dbname, key)
    params := []interface{}{start, end}
    return execWrite(redis,"ZREMRANGEBYSCORE", key, params...)
}

func ZRemRangeByRank(dbname string, info interface{},key string, start, end int) (err error){
    redis, err :=  GetRedisV2(dbname, info, true)
    if err != nil {
        return
    }
    defer redis.Close()
    key = simpleKey(dbname, key)
    params := []interface{}{start, end}
    return  execWrite(redis, "ZREMRANGEBYRANK", key, params...)
}

func ZScore(dbname string, info interface{},key string, member string) (ret int, err error){
    redis, err :=  GetRedisV2(dbname, info, false)
    if err != nil {
        return
    }
    defer redis.Close()
    key = simpleKey(dbname, key)
    return rds.Int(redis.Do("ZSCORE", key, member))
}

func ZRank(dbname string, info interface{},key string) (ret int, err error){
    redis, err :=  GetRedisV2(dbname, info, false)
    if err != nil {
        return
    }
    defer redis.Close()
    key = simpleKey(dbname, key)
    return rds.Int(redis.Do("ZRANK", key))
}

func ZRevRank(dbname string, info interface{},key string) (ret int, err error){
    redis, err :=  GetRedisV2(dbname, info, false)
    if err != nil {
        return
    }
    defer redis.Close()
    key = simpleKey(dbname, key)
    return rds.Int(redis.Do("ZREVRANK", key))
}

func ZIncrBy(dbname string, info interface{},key string, value float64, member interface{}) (v float64,err error){
    redis, err :=  GetRedisV2(dbname, info, true)
    if err != nil {
        return
    }
    defer redis.Close()
    key = simpleKey(dbname, key)
    params := []interface{}{value, member}
    return rds.Float64(execWrite2(redis, "ZINCRBY", key, params...))
}


//func (this *RedisClient)ZUinon(dbname string, info interface{},key string, value int, member interface{}) ([]string, error){
//  return []string, nil
//}
//func (this *RedisClient)ZInter(dbname string, info interface{},key string, value int, member interface{}) ([]string, error){
//    return []string, nil
//}
//func (this *RedisClient)ZScan(dbname string, info interface{},key string, value int, member interface{}) ([]string, error){
//    return []string, nil
//}

//=======================================================================================
//hash
func HSet(dbname string, info interface{},key string, hashKey string, value interface{}) (err error){
    redis, err :=  GetRedisV2(dbname, info, true)
    if err != nil {
        return
    }
    defer redis.Close()
    key = simpleKey(dbname, key)
    return execWrite(redis, "HSET", key, hashKey, value)
}

func HSetNx(dbname string, info interface{},key string, hashKey string, value interface{}) (err error){
    redis, err :=  GetRedisV2(dbname, info, true)
    if err != nil {
        return
    }
    defer redis.Close()
    key = simpleKey(dbname, key)
    return execWrite(redis, "HSETNX", key, hashKey, value)
}

func HLen(dbname string, info interface{},key string) (ret int, err error){
    redis, err :=  GetRedisV2(dbname, info, false)
    if err != nil {
        return
    }
    defer redis.Close()
    key = simpleKey(dbname, key)
    return rds.Int(redis.Do("HLEN", key))
}

func HDel(dbname string, info interface{},key string, hashkeys... string) (err error) {
    redis, err :=  GetRedisV2(dbname, info, true)
    if err != nil {
        return
    }
    defer redis.Close()
    key = simpleKey(dbname, key)
    params := make([]interface{}, 0, len(hashkeys))
    for _, k := range hashkeys{
        params = append(params, k)
    }
    return execWrite(redis, "HDel", key, params...)
}

func HKeys(dbname string, info interface{},key string) (ret []string, err error){
    redis, err :=  GetRedisV2(dbname, info, false)
    if err != nil {
        return
    }
    defer redis.Close()
    key = simpleKey(dbname, key)
    return rds.Strings(redis.Do("HKEYS", key))
}

func HVals(dbname string, info interface{},key string) (ret []string, err error){
    redis, err :=  GetRedisV2(dbname, info, false)
    if err != nil {
        return
    }
    defer redis.Close()
    key = simpleKey(dbname, key)
    return rds.Strings(redis.Do("HVALS", key))
}

func HGet(dbname string, info interface{},key string, hashKey string) (ret string, err error){
    redis, err :=  GetRedisV2(dbname, info, false)
    if err != nil {
        return
    }
    defer redis.Close()
    key = simpleKey(dbname, key)
    return rds.String(redis.Do("HGET", key, hashKey))
}

func HGetAll(dbname string, info interface{},key string) (ret map[string]string, err error){
    redis, err :=  GetRedisV2(dbname, info, false)
    if err != nil {
        return
    }
    defer redis.Close()
    key = simpleKey(dbname, key)
    ret = map[string]string{}
    values, err := rds.Strings(redis.Do("HGetAll", key))
    if err !=nil {
        return
    }

    for i:=0; i<len(values);i+=2{
        ret[values[i]] = values[i+1]
    }
    return
}

func HExists(dbname string, info interface{},key string, hashKey string) (ret bool, err error){
    redis, err :=  GetRedisV2(dbname, info, false)
    if err != nil {
        return
    }
    defer redis.Close()
    key = simpleKey(dbname, key)
    return rds.Bool(redis.Do("HExists", key, hashKey))
}

func HIncrBy(dbname string, info interface{},key string, hashKey string, v int) (value int, err error){
    redis, err :=  GetRedisV2(dbname, info, true)
    if err != nil {
        return
    }
    defer redis.Close()
    key = simpleKey(dbname, key)
    return rds.Int(execWrite2(redis,"HIncrBy", key, hashKey, v))
}


func HIncrByFloat(dbname string, info interface{},key string, hashKey string, v float32) (value float64, err error){
    redis, err :=  GetRedisV2(dbname, info, true)
    if err != nil {
        return
    }
    defer redis.Close()
    key = simpleKey(dbname, key)
    return rds.Float64(execWrite2(redis,"HIncrByFloat", key, hashKey, v))
}

func HMSet(dbname string, info interface{},key string, values map[string]string) (err error){
	if len(values) == 0 {
		return  nil
	}
    redis, err :=  GetRedisV2(dbname, info, true)
    if err != nil {
        return
    }
    defer redis.Close()
    key = simpleKey(dbname, key)
	params := []interface{}{}
	for k, v := range  values{
		params = append(params, k ,v )
	}

    return execWrite(redis,"HMSet", key, params...)
}

func HMGet(dbname string, info interface{},key string, hashKeys ...string) (ret map[string]string, err error){
	
    ls, err := HMGetKeys(dbname, info, key, hashKeys...)
	ret = map[string]string{}
	
    if err != nil {
        return 
    }
	cnt := len(ls)
	for i:=0;i<cnt;i++ {
		ret[hashKeys[i]] = ls[i]
	}
    return 
}

func HMGetKeys(dbname string, info interface{},key string, hashKeys ...string) (ret []string, err error){
    redis, err :=  GetRedisV2(dbname, info, false)
    if err != nil {
        return
    }
    defer redis.Close()
    key = simpleKey(dbname, key)
    params := []interface{}{key}
	for _, k := range hashKeys {
		params = append(params, k)
	}
    return rds.Strings(redis.Do("HMGET",  params...))
}

func MSet(dbname string, info interface{}, values map[string]string, expire int) (err error) {
    redis, err :=  GetRedisV2(dbname, info, true)
    if err != nil {
        return
    }
    defer redis.Close()
    for k , v := range values{
        k = simpleKey(dbname, k)
        if expire > 0 {
            redis.Send("SETEX", k, expire, v)
        }else{
            redis.Send("SET", k, v)
        }
    }
    err = redis.Flush()

    return
}

func MGet(dbname string, info interface{}, keys []string) (ret map[string]string, err error) {
    ret = map[string]string{}

    redis, err := GetRedisV2(dbname, info, false)
    if err != nil {
        return
    }
    defer redis.Close()
    params := []interface{}{}
    for _, k := range keys{
        params = append(params,  simpleKey(dbname, k))
    }
    strList, err  := rds.Strings(redis.Do("MGET",  params...))
    if err != nil {
        return
    }

    for i, v := range keys {
        ret[v] = strList[i]
    }
    return
}

//=======================================================================================
//private
func execWrite(c rds.Conn, cmd string, key string, values ...interface{}) (err error){
    params := []interface{}{key}
    if len(values) > 0 {
        params = append(params, values...)
    }

    _, err = c.Do(cmd,params...)
    return
}

func execWrite2(c rds.Conn, cmd string, key string, values ...interface{}) (reply interface{}, err error){
    params := []interface{}{key}
    if len(values) > 0 {
        params = append(params, values...)
    }


    return  c.Do(cmd,params...)
}