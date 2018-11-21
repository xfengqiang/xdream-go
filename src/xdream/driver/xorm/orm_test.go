package xorm

import (
	"database/sql"
	"fmt"
	"reflect"
	"runtime/debug"
	"testing"
	"time"
	"strings"
	"io/ioutil"
	"os"
	"encoding/json"
	"log"
)

const (
	dbUrl = "root:1234@/xtest?charset=utf8"
)

func init() {
	SetDebugModel(true)

	//f, _ := os.OpenFile("sample.json",os.O_RDONLY,  0777)

	f, _ := os.OpenFile("qconf.json",os.O_RDONLY,  0777)
	defer f.Close()

	configStr, _ := ioutil.ReadAll(f)

	var configs map[string]map[string]DbConfig
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
		fmt.Println("Register err:", err, string(debug.Stack()))
	}
}

//表结构
//CREATE TABLE `user2` (
//`id` int(11) NOT NULL,
//`name` varchar(64) NOT NULL,
//`status` tinyint(4) NOT NULL,
//`ctime` timestamp NOT NULL DEFAULT '0000-00-00 00:00:00' ON UPDATE CURRENT_TIMESTAMP,
//PRIMARY KEY (`id`)
//)  DEFAULT CHARSET=utf8

type ormUser struct {
	Id         int
	Name       string
	Status     int
	CreateTime time.Time
	Other      string
}
//必选方法，返回对象对应的数据库名
func (this *ormUser) GetDbName() string {
	return "xtest"
}

//可选方法，返回对象对应的表明，默认使用类名转驼峰转下划线的方式匹配
func (this *ormUser) GetTableName() string {
	return "user"
}

func (this *ormUser) OrmInfo() map[string]interface{} {
	//字段映射配置，可选，默认的使用驼峰转下划线的方式匹配
	fieldMap := map[string]string{
		"CreateTime": "create_time", //将CreateTime保存到ctime类
		"Id":         "id", //默认规则，不需配置
	}

	return map[string]interface{}{
		"ignore": "Id,Other", //执行insert时，不需要插入的字段，比如自增的id, 多个字段用逗号分隔,不运行包含空格
		"pk":     "Id", //主键字段，帮助orm识别哪个列是主键
		"fd_map": fieldMap,
	}
}

//可选方法，执行写操作前执行
func (this *ormUser) BeforeWrite(typ int) error {
	fmt.Println("Before write")
	return nil
}

//可选方法，执行写操作后执行
func (this *ormUser) AfterWrite(typ int) error {
	fmt.Println("After write")
	return nil
}

func TestDbQuery(t *testing.T) {
	fmt.Println("=========test db query=========")
	var err error
	db, err := NewDb("xtest", nil)
	//    if err != nil {
	//        t.Error(err)
	//    }
	u := ormUser{}
	err = db.QueryRow(&u, "SELECT * FROM user WHERE id=17")
	if err != nil && err != ErrNoRows {
		t.Error(err.Error())
	}

	fmt.Println(u)

	fmt.Println("=========test query row=========")
	users := []ormUser{}
	cnt, err := db.QueryRows(&users, "SELECT * FROM USER limit 1,2")
	fmt.Println("result count:", cnt)
	for _, u := range users {
		fmt.Println(u)
	}

	cnt, err = db.QueryCount("user", "")
	fmt.Println("counter:", cnt, "err:", err)

	fmt.Println("fetch raw:")
	db.QueryRaw(func(rows *sql.Rows) error {
		for rows.Next() {
			var name string
			rows.Scan(&name)
			fmt.Println("name:", name)
		}
		return nil
	}, "SELECT name from user where id>?", 1)
}

func TestDbCondition(t *testing.T) {
	fmt.Println("=========test db query=========")
	var err error
	db, err := NewDb("xtest", nil)
	//    if err != nil {
	//        t.Error(err)
	//    }
	u := ormUser{}
	err = db.QueryRowByCond(&u, "user", NewCondition().And("id", 1))
	if err != nil {
		t.Error(err.Error())
	}

	fmt.Println(u)

	fmt.Println("=========test query row=========")
	users := []ormUser{}
	cnt, err := db.QueryRowsByCond(&users, "user", NewCondition().Append("id", 1, ">"))
	fmt.Println("result count:", cnt)
	for _, u := range users {
		fmt.Println(u)
	}

	cnt, err = db.QueryCount("user", "")
	fmt.Println("counter:", cnt, "err:", err)

}

func TestDbWrite(t *testing.T) {
	db, err := NewDb("xtest", nil)
	if err != nil {
		t.Error(err.Error())
	}

	ret, err := db.Insert("user", "name,create_time", "abc", time.Now())
	if err != nil {
		t.Error(err.Error())
	}

	ret.LastInertId()
	if err != nil {
		t.Error(err.Error())
	} else {
		fmt.Println("Insert ok, insertId:", ret.LastInertId(), "cnt:", ret.RowsAffected())
	}
}

func TestOrmUtil(t *testing.T) {
	u := ormUser{Id: 100}

	f, _ := getValFromField(&u, "Id")
	fmt.Println(f)

	val, _ := callMethod(&u, "GetDbName")
	fmt.Println(val)

	//    return
	dbName := getDbNameForData(&u)
	tableName := getTableNameForData(&u)
	fmt.Println("db name:", dbName, "table name:", tableName)

	us := []ormUser{}
	dbName = getDbNameForData(&us)
	tableName = getTableNameForData(&us)
	fmt.Println("db name:", dbName, "table name:", tableName)

	ptrUs := []*ormUser{}
	dbName = getDbNameForData(&ptrUs)
	tableName = getTableNameForData(&ptrUs)
	fmt.Println("db name:", dbName, "table name:", tableName)

}

func TestOrmMapUtil(t *testing.T) {
	obj := ormUser{Id: 100, Name: "fank"}
	fMap := GetFieldMap(reflect.TypeOf(obj))
	fmt.Println(fMap)

	fd, vals, err := getInsertFields(&obj, "", "")
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println("fd:", fd, "vals:", vals, "err:", err)
}

func TestOrmQuery(t *testing.T) {
	fmt.Println("=========test orm query=========")

	u := ormUser{Id: 3}
	orm := NewOrm(nil)

	orm.QueryObjectByPk(&u, 3)

	fmt.Println("user is:", u)

	orm.QueryObject(&u, "id=?", 6)
	fmt.Println(u)

	users := []ormUser{}
	cnt, err := orm.QueryObjects(&users, "xtest", "SELECT * FROM user WHERE id > ? LIMIT 1", 1)
	if err != nil {
		fmt.Println("error:", err)
	} else {
		fmt.Println("Ret count:", cnt)
		for _, u := range users {
			fmt.Println(u)
		}
	}

	userPtrs := []*ormUser{}
	cnt, err = orm.QueryObjects(&userPtrs, "xtest", "SELECT * FROM user WHERE id > ? AND id < 10 LIMIT 2", 1)
	if err != nil {
		fmt.Println("error:", err)
	} else {
		fmt.Println("Ret count:", cnt)
		for _, u := range userPtrs {
			fmt.Println(u)
		}
	}

	fmt.Println("Query by In Condition:")
	cnt, err = orm.QueryObjectsByCond(&users, "xtest", "user", NewCondition().In("id", 1, 2, 3).Fields("id,name").OrderBy("id", "desc").Limit(2))
	if err != nil {
		fmt.Println("error:", err)
	} else {
		fmt.Println("Ret count:", cnt)
		for _, u := range users {
			fmt.Println(u)
		}
	}

}

//func TestGenQuery(t *testing.T){
//    return
//    // col1 = b AND col2>c OR d!=d  AND col4 IN (a,b,c) ORDER BY time DESC,name ASC LIMIT 10, 1
//    cond := NewCondition()
//    cond.Eq("col1", "v1").Gt("col2", 100).In("col3", 1,2,3,4).NotIn("col4", "a").Like("col5", "%5").OrderBy("col1", "DESC").OrderBy("col2", "ASC").Limit(1,100)
//    sql, params := cond.GenQuery()
//    fmt.Println("sql:", sql)
//    fmt.Println("params:", params)
//}

func TestOrmWriteBatch(t *testing.T) {
	users := []*ormUser{}
	for i := 0; i < 2; i++ {
		u := ormUser{Id: 100, Name: "aaaaabbc", CreateTime: time.Now()}
		users = append(users, &u)
	}

	orm := NewOrm(nil)
	cnt, err := orm.BatchInsert(users[0], &users, "Name", true)
	if err != nil {
		fmt.Println("err:", err)
	} else {
		fmt.Println("Insert count", cnt)
	}
}

func TestCondition(t *testing.T) {
	cond := NewCondition().And("id", 1).And("gender", "0", "!=").Or("age", 10, ">").In("name", []interface{}{"1", "2"}).OrderBy("id", "desc").OrderBy("age", "ASC").Limit(1, 100)
	condStr, values := cond.GetCondition()
	fmt.Println("cond:", condStr, "values:", values)
}

func TestOrmWrite(t *testing.T) {
	fmt.Println("=====Test orm write=========")

	u := &ormUser{Id: 117, Name: "2222", CreateTime: time.Now(), Status: 5}
	orm := NewOrm(nil)
	//    ret, err:=orm.Insert(u)
	//
	//    if err!=nil{
	//        t.Error(err.Error())
	//    }else{
	//        fmt.Println("Insert Ok:", ret.LastInertId())
	//    }
//	    ret, err := orm.InsertOrUpdate(u)
	//    if err!=nil{
	//        t.Error(err.Error())
	//    }else{
	//        fmt.Println("Insert  or update Ok:", ret.LastInertId(), "cnt:", ret.RowsAffected())
	//    }
	//
	u.Name = "b"
	u.Status = 5
	_, err := orm.Update(u, "", "Name,CreateTimes")

	if err != nil {
		t.Error(err.Error())
	}

	//    deleteCnt, err := orm.Delete(u)
	//    if err!=nil{
	//        t.Error(err.Error())
	//    }else{
	//        fmt.Println("Delete Ok: ", deleteCnt)
	//    }
}

func TestTransation(t *testing.T) {
	return
	db, err := NewDb("xtest", nil)
	if err != nil {
		t.Error(err.Error())
	}

	orm := NewOrmWithDb(db, nil)
	err = db.Begin()
	if err != nil {
		t.Error(err.Error())
	}

	u := &ormUser{Id: 117, Name: "rollback", CreateTime: time.Now()}
	_, err = orm.Insert(u)
	if err != nil {
		t.Error(err.Error())
	}
	err = db.Rollback()
	if err != nil {
		t.Error(err.Error())
	} else {
		fmt.Println("Rollback ok")
	}

	u.Name = "commit"
	err = db.Begin()
	_, err = orm.Insert(u)
	if err != nil {
		t.Error(err.Error())
	}
	err = db.Commit()
	if err != nil {
		t.Error(err.Error())
	} else {
		fmt.Println("Commit ok")
	}
}

//设置自定义的表名、列名映射函数
func TestCustomMapFunc(t *testing.T)  {
	
	//设置表名映射规则
	SetTableMapFunc(func (tableName string) string{
		col := strings.ToLower(tableName[0:1]) + tableName[1:]
		fmt.Printf("className:%v tableName:%v\n", tableName, col)
		return col
	})

	//设置列名映射规则
	SetFieldMapFunc(func (fieldName string) string{
		col := strings.ToLower(fieldName[0:1]) + fieldName[1:]
		fmt.Printf("fieldName:%v columnName:%v\n", fieldName, col)
		return col
	})
	
	var user ormUser
	orm := NewOrm(nil)
	err := orm.QueryObject(&user, "id=?", 1)
	fmt.Println("TestCustomMapFunc err:", err)
}
