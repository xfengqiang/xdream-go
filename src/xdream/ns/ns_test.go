package ns

import (
	"testing"
	"encoding/json"
	"log"
)

func printInstances(msg string , list []*Instance, err error)  {
	str, _ :=json.MarshalIndent(list, "", "\t")
	log.Printf("msg=%s err=%v instances=%s", msg, err, string(str))
}


func TestVerify(t *testing.T)  {
	var configStr string = `{
	"instances": [{
		"ip": "127.0.0.1",
		"port": 3306,
	}]
}`
	var config NsConfig
	if err := json.Unmarshal([]byte(configStr), &config); err != nil {
		t.Errorf("decode config error:%s", err.Error())
	}

	//configB, _ := json.MarshalIndent(config, "", "\t")
	//log.Println(string(configB))

	var (
		l []*Instance
		e error
	)


	ns := NameService{}

	routeIdc := NsRoute{
		Idc:"m6",
	}

	l, e = ns.FilterService(&config, &routeIdc)
	log.Println("=============test idc filter=============")
	printInstances("filter idc", l, e)
	for _, instance := range l {
		if instance.Idc != "m6" {
			t.Errorf("idc filter error. shoud be m6, meet %s", instance.Idc)
		}
	}

	log.Println("=============test ab filter=============")
	routeAb := NsRoute{
		Ab:"00-44",
	}

	l, e = ns.FilterService(&config, &routeAb)
	printInstances("filter default  env", l, e)
	for _, instance := range l {
		if instance.Ab != "set-1"  {
			t.Errorf("env filter error. shoud be ol, meet %s", instance.Ab)
		}
	}


	log.Println("=============test default filter=============")
	routeDefault := NsRoute{
	}
	config.RouteRule["idc"]["default"] = "";
	l, e = ns.FilterService(&config, &routeDefault)
	printInstances("filter default  env", l, e)
	for _, instance := range l {
		if instance.Env != "ol" && instance.Env!="smallflow" && instance.Env!="" {
			t.Errorf("env filter error. shoud be ol, meet %s", instance.Env)
		}
	}


}


func TestFilter(t *testing.T)  {
var configStr string = `{
	"dns_type": "raw",
	"lb_type": "wrr",
	"retry_cnt": 2,
	"route_rule": {
		"idc": {
			"m6": "m6",
			"xg": "xg",
			"default": "m6"
		},
		"ab": {
			"00-44": "set-1",
			"45-50": "set-2"
		},
		"env": {
			"sandbox": "sandbox",
			"smallflow": "smallflow"
		}
	},
	"instances": [{
		"ip": "10.12.20.151",
		"port": 6379,
		"idc": "m6",
		"status": 0,
		"env": "sandbox",
		"weight": 10
	},
	{
		"ip": "10.12.20.152",
		"port": 3550,
		"idc": "m6",
		"status": 0,
		"env": "smallflow",
		"weight": 20
	},
	{
		"ip": "10.12.20.153",
		"port": 3550,
		"idc": "m6",
		"status": 0,
		"weight": 20,
		"ab": "set-1"
	},
	{
		"ip": "10.12.20.156",
		"port": 3550,
		"idc": "m6",
		"status": 0,
		"weight": 20,
		"ab": "set-1"
	},
	{
		"ip": "10.12.20.157",
		"port": 3550,
		"idc": "m6",
		"status": 0,
		"weight": 20
	},
	{
		"ip": "10.12.20.158",
		"port": 3550,
		"idc": "xg",
		"status": 0,
		"weight": 70
	}]
}`
	var config NsConfig
	if err := json.Unmarshal([]byte(configStr), &config); err != nil {
		t.Errorf("decode config error:%s", err.Error())
	}

	//configB, _ := json.MarshalIndent(config, "", "\t")
	//log.Println(string(configB))

	var (
		l []*Instance
		e error
	)


	ns := NameService{}

	routeIdc := NsRoute{
		Idc:"m6",
	}

	l, e = ns.FilterService(&config, &routeIdc)
	log.Println("=============test idc filter=============")
	printInstances("filter idc", l, e)
	for _, instance := range l {
		if instance.Idc != "m6" {
			t.Errorf("idc filter error. shoud be m6, meet %s", instance.Idc)
		}
	}

	log.Println("=============test ab filter=============")
	routeAb := NsRoute{
		Ab:"00-44",
	}

	l, e = ns.FilterService(&config, &routeAb)
	printInstances("filter default  env", l, e)
	for _, instance := range l {
		if instance.Ab != "set-1"  {
			t.Errorf("env filter error. shoud be ol, meet %s", instance.Ab)
		}
	}


	log.Println("=============test default filter=============")
	routeDefault := NsRoute{
	}
	config.RouteRule["idc"]["default"] = "";
	l, e = ns.FilterService(&config, &routeDefault)
	printInstances("filter default  env", l, e)
	for _, instance := range l {
		if instance.Env != "ol" && instance.Env!="smallflow" && instance.Env!="" {
			t.Errorf("env filter error. shoud be ol, meet %s", instance.Env)
		}
	}


}

func TestSelect(t *testing.T)  {
var  configStr string = `
{
	"dns_type": "raw",
	"dns_name": "com.x.user",
	"lb_type": "wrr",
	"retry_cnt": 2,
	"route_rule": {
		"idc": {
			"default": ""
		}
	},
	"instances": [{
		"ip": "10.12.20.151",
		"port": 6379,
		"idc": "m6",
		"status": 0,
		"weight": 10
	},
	{
		"ip": "10.12.20.152",
		"port": 3550,
		"idc": "m6",
		"status": 0,
		"weight": 20
	},
	{
		"ip": "10.12.20.153",
		"port": 3550,
		"idc": "m6",
		"status": 0,
		"weight": 70
	}]
}
`
	var config NsConfig
	if err := json.Unmarshal([]byte(configStr), &config); err != nil {
		t.Errorf("decode config error:%s", err.Error())
	}
	ns := NameService{}
	route  := NsRoute{
		Idc:"",
	}

	total := 10000
	var (
		instances []*Instance
		err error
	)

	statMap := map[string]int{}
	addTotal := 0
	for i:=0;i<total;i++ {
		instances, err = ns.GetServices(config, route)

		if err != nil {
			t.Errorf("get service error:%s",err.Error())
		}

		for _, instance := range instances {
			statMap[instance.Ip]++
			addTotal++
		}
	}

	log.Printf("addTotal:%d\n",addTotal)
	if addTotal != total*config.RetryCnt {
		t.Errorf("total count wrong, should be %d\n", total*config.RetryCnt)
	}
	for ip, cnt := range statMap{
		log.Printf("ip:%s cnt:%d rate:%f\n", ip, cnt, float32(cnt)/float32(config.RetryCnt)/float32(total))
	}
}


func TestQconf(t *testing.T)  {
	var  configStr string = `
{
	"dns_type": "qconf",
	"dns_name": "/name_service/mysql.user",
	"lb_type": "wrr",
	"retry_cnt": 1
}
`
	var config NsConfig
	if err := json.Unmarshal([]byte(configStr), &config); err != nil {
		je, _ := err.(interface{}).(json.SyntaxError)

		t.Errorf("decode config error:%s %d", err.Error(),je.Offset)
	}
	ns := NameService{}
	route  := NsRoute{
		Idc:"",
	}

	total := 10000
	var (
		instances []*Instance
		err error
	)

	statMap := map[string]int{}
	addTotal := 0
	for i:=0;i<total;i++ {
		instances, err = ns.GetServices(config, route)

		if err != nil {
			t.Errorf("get service error:%s",err.Error())
		}

		for _, instance := range instances {
			statMap[instance.Ip]++
			addTotal++
		}
	}

	log.Printf("addTotal:%d\n",addTotal)
	if addTotal != total*config.RetryCnt {
		t.Errorf("total count wrong, should be %d\n", total*config.RetryCnt)
	}
	for ip, cnt := range statMap{
		log.Printf("ip:%s cnt:%d rate:%f\n", ip, cnt, float32(cnt)/float32(config.RetryCnt)/float32(total))
	}
}
