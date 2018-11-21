package ns

import (
	"fmt"
	"math/rand"
	"log"
	"encoding/json"
	"infra/go_qconf"
)

var (
	envValues map[string]int = map[string]int{
		"sandbox":1,
		"smallflow":2,
		"ol":6,
	}
	targetValues map[string]int = map[string]int{
		"sandbox": 1,
		"smallflow": 2,
		"ol" : 4,
	}
)

type Instance struct {
	Ip string `json:"ip"`
	Port int `json:"port"`
	Status int `json:"status"`
	Weight int `json:"weight"`
	Idc string `json:"idc"`
	Ab string `json:"ab"`
	Env string `json:"env"`
}

type NsRoute struct {
	Idc string
	Ab string
	Env string
}

type NsConfig struct {
	DnsType string `json:"dns_type"`
	DnsName string `json:"dns_name"`
	LbType  string `json:"lb_type"`
	RetryCnt int `json:"retry_cnt"`
	RouteRule map[string]map[string]string `json:"route_rule"`
	Instances []*Instance `json:"instances"`
}

type NameService struct {

}

func (ns *NameService)GetServices(config NsConfig,  route NsRoute)  (instances []*Instance, err error){
	if config.RouteRule == nil {
		config.RouteRule  = map[string]map[string]string{}
	}

	//默认LbType
	if config.LbType == "" {
		config.LbType = "rr"
	}

	//默认重试次数
	if config.RetryCnt <= 0 {
		config.RetryCnt = 1
	}

	if instances, err = ns.FilterService(&config, &route); err != nil {
		return
	}

	instances, err = ns.SelectService(instances, &config)
	return
}

func (ns *NameService)FilterService(config *NsConfig, route *NsRoute) (instances []*Instance, err error) {
	dnsType := config.DnsType
	if dnsType == "" {
		dnsType = "raw"
	}
	switch dnsType {
	case "raw":
		//do nothing
	case "qconf":
		value, e := go_qconf.GetConf(config.DnsName, "")
		if e != nil {
			err = fmt.Errorf("get qconf config failed for key %s, err:%s", config.DnsName, e.Error())
			return
		}
		var  cfg NsConfig
		e = json.Unmarshal([]byte(value), &cfg)
		if e!= nil {
			err = fmt.Errorf("json  decode config failed for key %s, err:%s", config.DnsName, e.Error())
		}
		config.RouteRule = cfg.RouteRule
		config.Instances = cfg.Instances
	default:
		err = fmt.Errorf("unsupport dns type:%s ", dnsType)
		return
	}

	if len(config.Instances) == 0 {
		return []*Instance{}, nil
	}else if len(config.Instances) == 1 {
		return config.Instances, nil
	}

	var defaultIdc string
	if idc, ok := config.RouteRule["idc"]["default"]; ok {
		defaultIdc = idc
	}

	//idc 为空时，使用配置的默认idc
	if route.Idc == "" {
		route.Idc = defaultIdc
	}

	//env 默认匹配线上
	if  route.Env == "" {
		route.Env = "ol"
	}

	var instanceIdc string
	var instanceEnv string

	var targetAb string
	if route.Ab != "" {
		targetAb = config.RouteRule["ab"][route.Ab]
	}

		//根据route参数筛选匹配的实例
	for _, instance := range config.Instances {
		if instance.Status != 0 { //非0表示下线
			log.Println("ignore for status=0")
			continue
		}

		//实例没有配置idc属性，归类为default idc
		instanceIdc = instance.Idc
		if instanceIdc == "" {
			instanceIdc = defaultIdc
		}
		if route.Idc !="" && route.Idc != instanceIdc {
			//log.Printf("ignore for idc supposeIdc=%s instanceIdc=%s\n", route.Idc, instanceIdc)
			continue
		}

		//match ab
		if targetAb != instance.Ab {
			//log.Printf("ignore for ab supposeAb=%s instanceAb=%s\n", route.Ab, instance.Ab)
			continue
		}

		//match env
		//sandbox 仅匹配sandbox
		//smallflow仅匹配smallflow
		//ol匹配smallflow和sandbox
		instanceEnv = instance.Env
		if instanceEnv == "" {
			instanceEnv = "ol"
		}
		if (envValues[route.Env] & targetValues[instanceEnv]) == 0 {
			//log.Printf("ignore for env supposeEnv=%s instanceEnv=%s\n", route.Env, instance.Env)
			continue
		}

		instances = append(instances, instance)
	}

	return
}

func (ns *NameService)SelectService(list []*Instance, config *NsConfig) (instances []*Instance, err error)  {
	cnt := len(list)
	if cnt == 0 {
		return
	}else if cnt == 1 {
		instances = list
		return
	}

	switch config.LbType {
		case "rr":
			idx := rand.Int()%cnt
			for i:=0; i< config.RetryCnt; i++ {
				instances = append(instances, list[idx+i])
			}
		case "wrr":
			var totalWeight int
			for _, instance := range list {
				if instance.Weight < 0 {
					continue
				}else if instance.Weight == 0 {
					instance.Weight = 100
				}

				totalWeight += instance.Weight
			}

			for i:=0; i<config.RetryCnt; i++ {
				offset := rand.Int()%totalWeight
				for _, instance := range list {
					offset -= instance.Weight
					if offset<=0 {
						instances = append(instances, instance)
						break
					}
				}
			}
		default:
			err = fmt.Errorf("wrong rr type:%s", config.LbType)
			return
	}

	return
}