package xorm

import (
	"database/sql"
	"errors"
	"fmt"
	"runtime/debug"
	"xdream/ns"
)

const (
	format_Date     = "2006-01-02"
	format_DateTime = "2006-01-02 15:04:05"
)

const (
	DB_DNS_TYPE_IP = "ip"
)

var (
	dbConfigs   map[string]*DbConfig
	dbConnCache map[string]*sql.DB
)

type DbConfig struct {
	Auth struct{
		Tag string `json:"tag"`
		Username string `json:"username"`
		Password string `json:"password"`
		Dbname string `json:"dbname"`
		Args string `json:"args"`
		MaxIdle int `json:"max_idle"`
		MaxOpen int `json:"max_open"`
	}
	Services ns.NsConfig

	ip string
	port int
	fullUrl string `json:"-"`
}

func (this *DbConfig) GetFullUrl() (string, error) {
	if len(this.Auth.Args) == 0 {
		this.Auth.Args = "charset=utf8&loc=Asia%2FShanghai"
	}

	if len(this.fullUrl) == 0 {
		nameService := ns.NameService{}
		var (
			instances []*ns.Instance
			err error
			)
		if instances, err = nameService.GetServices(this.Services, routeConfig); err != nil {
			return "", fmt.Errorf("no host found for %s-%s err:%s", this.Auth.Dbname, this.Auth.Tag, err.Error())
		}

		if len(instances) == 0 {
			return "", fmt.Errorf("no host found for %s-%s ", this.Auth.Dbname, this.Auth.Tag)
		}

		this.ip = instances[0].Ip
		this.port = instances[0].Port
		this.fullUrl = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s", this.Auth.Username, this.Auth.Password, instances[0].Ip, instances[0].Port, this.Auth.Dbname, this.Auth.Args)
	}

	//log.Println("full url", this.fullUrl)
	return this.fullUrl, nil
}

func (this *DbConfig) GetCacheKey() (url string, e error) {
	url, e = this.GetFullUrl()
	return
}

func GetDbConfig(name string) *DbConfig {
	if dbConfig, ok := dbConfigs[name]; ok {
		return dbConfig
	}
	return nil
}

func GetConnForConfig(dbConfig *DbConfig) (db *sql.DB, err error) {
	defer func() {
		if err != nil {
			fmt.Println(err, string(debug.Stack()))
		}
	}()

	cacheKey, err := dbConfig.GetCacheKey()
	if err != nil {
		return nil, err
	}

	dbConn, ok := dbConnCache[cacheKey]
	if ok {
		return dbConn, nil
	}
	url, err := dbConfig.GetFullUrl()
	if err != nil {
		return nil, err
	}

	dbConn, err = sql.Open("mysql", url)
	if err != nil {
		err = fmt.Errorf("open db error . %s [url:%s]", err.Error(), url)
		if dbConn != nil {
			dbConn.Close()
		}
		return nil, err
	}

	if dbConfig.Auth.MaxIdle > 0 {
		dbConn.SetMaxIdleConns(dbConfig.Auth.MaxIdle)
	}
	if dbConfig.Auth.MaxOpen > 0 {
		dbConn.SetMaxOpenConns(dbConfig.Auth.MaxOpen)
	}
	dbConnCache[cacheKey] = dbConn

	return dbConn, nil
}

func GetDbConnByName(name string, master bool) (db *sql.DB, err error) {
	var dbConfig *DbConfig
	var ok bool
	if !master {
		slaveName := name + "-slave"
		dbConfig, ok = dbConfigs[slaveName]
	}
	if dbConfig == nil { //没有配置丛库
		dbConfig, ok = dbConfigs[name]
		master = true
		if !ok {
			return nil, errors.New(fmt.Sprintf("Db %s not registered", name))
		}
	} else {
		//        fmt.Println("use slave db")
	}
	return GetConnForConfig(dbConfig)
}

func CheckDbConn(dbName string, isMaster bool) error {
	db, err := GetDbConnByName(dbName, isMaster)
	if err != nil {
		return err
	}
	return db.Ping()
}

func RegisterDb(dbName string, config *DbConfig) error {
	dbConfigs[dbName] = config
	return CheckDbConn(dbName, true)
}

func RegisterSlaveDb(dbName string, config *DbConfig) error {
	dbConfigs[dbName+"-slave"] = config
	return CheckDbConn(dbName, false)
}

//*=====================
func init() {
	dbConfigs = map[string]*DbConfig{}
	dbConnCache = map[string]*sql.DB{}
}
