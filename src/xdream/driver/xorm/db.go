package xorm

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
	"xdream/ns"
)

const (
	QueryRetRow = iota
	QueryRetRows
	QueryRetExec
)

var (
	defaultTimeZone *time.Location
	debugModel      bool = false
)

type xdb interface {
	Prepare(query string) (*sql.Stmt, error)
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}

type tx interface {
	Begin() (*sql.Tx, error)
}

type txEnder interface {
	Commit() error
	Rollback() error
}

//db定义
type Db struct {
	Name     string      //名字
	isTx     bool        //是否事务
	txDbConn xdb         //链接对象
	Info     interface{} //相关信息
}

//定义callback
type DBCallBack func(db *Db, isMaster bool, timeused time.Duration, e error, sql string, args ...interface{})

//数据库回调函数
var (
	dbCallBack DBCallBack
	routeConfig ns.NsRoute
)


//注册callback
func RegisterCallBack(callback DBCallBack) {
	dbCallBack = callback
}


//Ns路由配置
func SetNsConfig(config ns.NsRoute)  {
	routeConfig = config
}

func GetTimeZone() *time.Location {
	if defaultTimeZone == nil {
		defaultTimeZone, _ = time.LoadLocation("Asia/Shanghai")
	}
	return defaultTimeZone
}

//创建数据库
func NewDb(name string, info interface{}) (*Db, error) {
	config := GetDbConfig(name)
	if config == nil {
		return nil, errors.New(fmt.Sprintf("Db %s not registered", name))
	}
	return &Db{Name: name, isTx: false, Info: info}, nil
}

func (this *Db) GetDbConnection(isMaster bool) (xdb, error) {
	if this.isTx {
		return this.txDbConn, nil
	}
	return GetDbConnByName(this.Name, isMaster)
}

func (this *Db) QueryResult(sql string, params ...interface{}) (rt *QueryRet) {
	if debugModel {
		fmt.Println("[sql]", sql, params)
	}

	db, err := this.GetDbConnection(false)
	if err != nil {
		return NewQueryRetWithError(QueryRetRows, err)
	}
	startTime := time.Now()
	defer func() {
		if dbCallBack != nil {
			var e error
			if rt != nil {
				e = rt.err
			}
			dbCallBack(this, false, time.Now().Sub(startTime), e, sql, params...)
		}
	}()
	rows, err := db.Query(sql, params...)
	if err != nil {
		return NewQueryRetWithError(QueryRetRows, err)
	}
	defer rows.Close()

	retValues := make([]map[string]interface{}, 0)
	for rows.Next() {
		var rowValue []interface{}
		columns, err := rows.Columns()
		if err != nil {
			return NewQueryRetWithError(QueryRetRows, err)
		}
		columnsMp := make(map[string]interface{}, len(columns))

		rowValue = make([]interface{}, 0, len(columns))
		for _, col := range columns {
			var ref interface{}
			columnsMp[col] = &ref
			rowValue = append(rowValue, &ref)
		}

		if err := rows.Scan(rowValue...); err != nil {
			return NewQueryRetWithError(QueryRetRows, err)
		}
		retValues = append(retValues, columnsMp)
	}

	return NewQueryRet(QueryRetRows, retValues)
}

func (this *Db) QueryRow(data interface{}, sql string, params ...interface{}) error {
	return this.QueryResult(sql, params...).FetchRow(data)
}

func (this *Db) QueryRows(data interface{}, sql string, params ...interface{}) (int64, error) {
	return this.QueryResult(sql, params...).FetchRows(data)
}

func (this *Db) QueryCount(table string, where string, params ...interface{}) (int64, error) {
	var sql string
	if where != "" && !strings.HasPrefix(where, "WHERE") {
		where = "WHERE " + where
	}
	sql = fmt.Sprintf("SELECT count(*) AS cnt FROM  %s %s", table, where)

	if debugModel {
		fmt.Println("[sql] "+sql, "[params]", params)
	}
	db, err := this.GetDbConnection(false)
	if err != nil {
		return 0, err
	}

	rows, err := db.Query(sql, params...)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var count int64
	if rows.Next() {
		err = rows.Scan(&count)
	}
	return count, err
}

func (this *Db) QueryRaw(fectchFunc func(sql *sql.Rows) error, sql string, params ...interface{}) error {
	db, err := this.GetDbConnection(false)
	if err != nil {
		return err
	}
	rows, err := db.Query(sql, params...)
	if err != nil {
		return err
	}

	defer rows.Close()
	if fectchFunc != nil {
		err = fectchFunc(rows)
	}

	return err
}

func (this *Db) QueryRowByCond(data interface{}, table string, cond *Condition) error {
	where, values := cond.GetCondition()
	sql := fmt.Sprintf("SELECT %s FROM %s %s", cond.GetFields(), table, where)
	return this.QueryRow(data, sql, values...)
}

func (this *Db) QueryCountByCond(table string, cond *Condition) (int64, error) {
	where, values := cond.GetCondition()
	return this.QueryCount(table, where, values...)
}

func (this *Db) QueryRowsByCond(data interface{}, table string, cond *Condition) (int64, error) {
	where, values := cond.GetCondition()
	sql := fmt.Sprintf("SELECT %s FROM %s %s", cond.GetFields(), table, where)
	return this.QueryRows(data, sql, values...)
}

func (this *Db) Insert(table string, fields string, values ...interface{}) (*QueryRet, error) {
	holders := ""
	if len(values) > 0 {
		holders = strings.Repeat("?,", len(values))
		holders = holders[:len(holders)-1]
	}

	sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", table, fields, holders)
	return this.Exec(sql, values...)
}

func (this *Db) InsertIgnore(table string, fields string, values ...interface{}) (*QueryRet, error) {
	holders := ""
	if len(values) > 0 {
		holders = strings.Repeat("?,", len(values))
		holders = holders[:len(holders)-1]
	}

	sql := fmt.Sprintf("INSERT IGNORE INTO %s (%s) VALUES (%s)", table, fields, holders)
	return this.Exec(sql, values...)
}

//func (this *Db)InsertOrUpdate(table string, data interface{}, fields...string)(int64, error){return 0, nil} //No need, use Exec
func (this *Db) Update(table string, setField string, where string, params ...interface{}) (*QueryRet, error) {
	var sql string
	if len(where) > 0 {
		sql = fmt.Sprintf("UPDATE %s SET %s WHERE %s", table, setField, where)
	} else {
		sql = fmt.Sprintf("UPDATE %s SET %s", table, setField)
	}
	return this.Exec(sql, params...)
} // No need, use Exec

func (this *Db) UpdateByCond(table string, cond *Condition, setField string, setValues ...interface{}) (*QueryRet, error) {
	where, values := cond.GetCondition()
	values = append(setValues, values...)
	sql := fmt.Sprintf("UPDATE %s SET %s %s", table, setField, where)
	return this.Exec(sql, values...)
}

func (this *Db) Delete(table string, where string, params ...interface{}) (*QueryRet, error) {
	var sql string
	if len(where) > 0 {
		sql = fmt.Sprintf("DELETE FROM `%s` WHERE %s", table, where)
	} else {
		sql = fmt.Sprintf("DELETE FROM `%s` ", table)
	}
	return this.Exec(sql, params...)
}

func (this *Db) DeleteByCond(table string, cond *Condition) (*QueryRet, error) {
	where, values := cond.GetCondition()
	sql := fmt.Sprintf("DELETE FROM %s %s", table, where)
	return this.Exec(sql, values...)
}

func (this *Db) BatchInsert(table string, fields string, rows [][]interface{}, ignoreDuplicate bool) (int64, error) {
	if len(rows) == 0 {
		return 0, nil
	}

	columns := strings.Split(fields, ",")
	colCnt := len(columns)
	if colCnt == 0 {
		return 0, errors.New("No field to insert")
	}

	rowHolder := strings.Repeat("?,", colCnt)
	rowHolder = "(" + rowHolder[:len(rowHolder)-1] + "),"
	holder := strings.Repeat(rowHolder, len(rows))
	holder = holder[:len(holder)-1]
	var sql string
	if ignoreDuplicate {
		sql = fmt.Sprintf("INSERT IGNORE INTO `%s` (%s) VALUES %s", table, fields, holder)
	} else {
		sql = fmt.Sprintf("INSERT INTO `%s` (%s) VALUES %s", table, fields, holder)
	}

	values := make([]interface{}, 0, len(rows)*colCnt)
	for _, row := range rows {
		values = append(values, row...)
	}

	ret, err := this.Exec(sql, values...)
	if err != nil {
		return 0, err
	}

	return ret.RowsAffected(), nil
}

func (this *Db) Exec(sql string, params ...interface{}) (rt *QueryRet, err error) {
	if debugModel {
		fmt.Println("[sql] "+sql, params)
	}

	db, err := this.GetDbConnection(true)

	if err != nil {
		return NewQueryRetWithError(QueryRetRows, err), err
	}
	startTime := time.Now()
	defer func() {
		if dbCallBack != nil {
			dbCallBack(this, true, time.Now().Sub(startTime), err, sql, params...)
		}
	}()
	ret, err := db.Exec(sql, params...)
	return NewQueryRetWithSqlRet(QueryRetExec, ret, err), err
}

func (this *Db) Begin() error {
	if this.isTx {
		return errors.New("A nother transation is running")
	}

	db, err := this.GetDbConnection(true)
	if err == nil {
		this.txDbConn, err = db.(tx).Begin()
		this.isTx = true
	}

	return err
}

func (this *Db) Commit() error {
	if this.txDbConn == nil {
		this.isTx = false
		return nil
	}

	err := this.txDbConn.(txEnder).Commit()

	this.isTx = false
	this.txDbConn = nil

	return err
}

func (this *Db) Rollback() error {
	if this.txDbConn == nil {
		this.isTx = false
		return nil
	}

	err := this.txDbConn.(txEnder).Rollback()
	this.isTx = false
	this.txDbConn = nil

	return err
}