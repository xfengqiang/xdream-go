package xredis
import (
    "testing"
    "time"
    "fmt"
    "os"
    "io/ioutil"
    "encoding/json"
    "log"
)

func init() {
    //f, _ := os.OpenFile("sample.json",os.O_RDONLY,  0777)

    f, _ := os.OpenFile("sample.json",os.O_RDONLY,  0777)
    defer f.Close()

    configStr, _ := ioutil.ReadAll(f)

    var configs map[string]RedisConfig
    err := json.Unmarshal([]byte(configStr), &configs)
    if err!= nil {
        log.Println("decode config error :",err)
    }

    if false {
        v, _ := json.MarshalIndent(configs, "", "\t")
        log.Println(string(v))

    }

    err = InitWithConfig(configs)

	if err != nil {
		panic(err)
	}


    RegisterCallBack(func (c *RedisConn, timeused time.Duration, err error,  cmd string, args...interface{}){
        fmt.Printf("callback conf:%v info:%v cmd:%v args:%v\n", c.Conf, c.Conf.GetInfo(), cmd, args)
    })
}


func TestStr(t *testing.T){
    Set("topic", "abc", "key1", "value1")
	Set("topic", "abc", "key2", "value2")
    v, _ := Get("topic", "abc", "key1")
    if v!="value1" {
        t.Error("str set err")
    }
	values, _ := MGet("topic", "abc", []string{"key1", "key2"})
	fmt.Println("MGet values:", values)
	
    Del("topic", "abc", "key1")
    vdel, _ := Get("topic", "abc", "key1")
    if vdel!="" {
        t.Error("str del err")
    }
    
    //incr incrby decr decrby ttl
    db := "topic"
    Set(db, nil , "key2", "5")
    v0, err := Incr(db, nil, "key2")
    fmt.Println("v", v, err)
    if v0 != 6 {
        t.Error("Incr err")
    }
    v2, err := Decr(db, nil, "key2")
    if v2 != 5 {
        t.Error("Incr err")
    }

    v3, _ := IncrBy(db, nil, "key2", 5)
    if v3 != 10{
        t.Error("IncrBy err", v3)
    }
    v4, _ := DecrBy(db, nil, "key2", 5)
    if v4 != 5 {
        t.Error("DecrBy err", v4)
    }
}

func TestZset(t *testing.T) {
    fmt.Println("run test~~~")
    db := "topic"
    Del(db, nil, "zset")
    ZAdd(db, nil, "zset", 100, "abc", 200, "bac")
    cnt, _ := ZCard(db, nil, "zset")
    if cnt != 2 {
        t.Error("zadd or zcard err")
    }
 
    ret, _ := ZRevRangeByScore(db, nil, "zset", "+inf", "-inf", true)
    fmt.Println(ret)

    ret2, _ := ZRange(db, nil, "zset", 0, -1)
    fmt.Println("zrange", ret2)

    ret3, _ := ZRevRange(db, nil, "zset", 0, -1)
    fmt.Println("zrevrange", ret3)

	retset1, _ := ZRangeWithScore(db, nil, "zset", 0, -1)
	fmt.Println("ZRangeWithScore", retset1)

	retset2, _ := ZRevRangeWithScore(db, nil, "zset", 0, -1)
	fmt.Println("ZRevRangeWithScore", retset2)
	
	
    ZRem(db, nil, "zset", "abc")

    cnt, _ = ZCard(db, nil, "zset")
    if cnt != 1 {
        t.Error("zrem err")
    }

    ret4, _ := ZIncrBy(db, nil, "zset", 10.5, "incrv")
    if ret4 != 10.5 {
        t.Error("ZIncrBy err")
    }
}

func TestList(t *testing.T) {
    fmt.Println("run test~~~")
    db := "topic"
    Del(db, nil, "list")
    LPush(db, nil, "list", "b")
    LPush(db, nil, "list", "a")
   
   
    ret, _ := LRange(db, nil, "list", 0, 2)
    fmt.Println(ret)
    if len(ret) != 2 {
        t.Error("list rpush err or range err")
    }
    
    l, _:= LLen(db,nil,"list")
    if l != 2 {
        t.Error("LLen err")
    }

    v2, _ := LIndex(db,nil, "list", 0)
    fmt.Println("v2", v2)
    if v2 !="a" {
        t.Error("LIndex err")
    }
    
    v1, _ := LPop(db,nil, "list")
    fmt.Println("v1", v1)
    if v1 !="a" {
        t.Error("LPop err")
    }
    LPush(db, nil, "list", "a")
    
    
    RPush(db, nil, "list", "c")
    v, _ := RPop(db,nil, "list")
    fmt.Println("v", v)
    if v !="c" {
        t.Error("rpush or rpop err")
    }
}

func TestSet(t *testing.T) {
    fmt.Println("run set test~~~")
    db := "topic"

    Del(db, nil, "set")
    
    v, err := SAdd(db, nil, "set", "a")
	fmt.Println("sadd 1", v, err)
	v, err  = SAdd(db, nil, "set", "a")
	fmt.Println("sadd 2", v, err)
	
	
    isMem, _ := SIsMember(db, nil, "set", "a")
    if !isMem {
        t.Error("SAdd  err")
    }
    
    cnt, _ := SCard(db, nil, "set")
    if cnt != 1 {
        t.Error("SCard  err")
    }
    
    SRem(db, nil, "set", "a")
    isMem, _ = SIsMember(db, nil, "set", "a")
    if isMem {
        t.Error("Srem  err")
    }

    v2,_ := SPop(db, nil, "set")
    fmt.Println("spop", v2)

	SAdd(db, nil, "set", "a", "b", "c")
	idx, list, err := SScan(db, nil, "set", 0, 1)
	fmt.Println("sscan", idx, list, err)
}

func TestHash(t *testing.T) {
    fmt.Println("run test~~~")
    db := "topic"
    Del(db, nil, "hash")
    HSet(db, nil, "hash", "key1", "v1")
    v, err := HGet(db, nil, "hash", "key1")
    fmt.Println("v", v, err)
    if v !=  "v1"{
        t.Error("HGet  err")
    }
    
    all, _ := HGetAll(db, nil, "hash")
    fmt.Println("all", all)
    
    keys, _ := HKeys(db, nil, "hash")
    fmt.Println("keys", keys)

    vals, _ := HVals(db, nil, "hash")
    fmt.Println("vals", vals)

    HSet(db, nil, "hash", "key2", "1")
    l, _ := HLen(db, nil, "hash")
    if l != 2 {
        t.Error("HLen  err")
    }
    HIncrBy(db, nil, "hash", "key2", 1)
    v2, _ := HGet(db, nil, "hash", "key2")
    if v2 != "2"{
        t.Error("HIncrBy  err")
    }
    HDel(db, nil, "hash", "key1")
    l, err = HLen(db, nil, "hash")
    if l != 1 {
        t.Error("HDel  err")
    }
	
	HMSet(db, nil, "hash", map[string]string{"key1":"v1", "key2":"v22"})
	values, err := HMGet(db, nil,  "hash", "key1", "key2","key3")
	fmt.Println("test HMGet", values, err)
	
}