package xorm

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"

	_ "github.com/go-sql-driver/mysql"
)

var (
	dbObjecCache = map[string]*Db{}
)

const (
	DB_WRITE_TYPE_INSERT = iota
	DB_WRITE_TYPE_UPDATE
	DB_WRITE_TYPE_DELETE
)

var ErrNoRows error = sql.ErrNoRows

type OrmInterface interface {
	GetDbName() string
	GetTableName() string
	GetOrmInfo() map[string]interface{}
}

/**
实现ORM, 可以自动根据Model的提供的接口推导出db和table
*/
type Orm struct {
	supportTx bool
	innerDb   *Db
	info      interface{}
}

func NewOrm(info interface{}) *Orm {
	return &Orm{info: info}
}

func NewOrmWithDb(db *Db, info interface{}) *Orm {
	return &Orm{innerDb: db, info: info}
}

func (this *Orm) GetDb(dbName string) (*Db, error) {
	if this.innerDb != nil {
		return this.innerDb, nil
	}

	db, ok := dbObjecCache[dbName]
	if ok {
		return db, nil
	}

	db, err := NewDb(dbName, this.info)
	if err == nil {
		dbObjecCache[dbName] = db
	}
	return db, err
}

/*
   data 必须是指向struct的指针
*/
func (this *Orm) QueryObject(data interface{}, where string, param ...interface{}) error {
	dbName := getDbNameForData(data)
	table := getTableNameForData(data)
	sql := fmt.Sprintf("SELECT * FROM %s WHERE %s LIMIT 1", table, where)

	db, err := this.GetDb(dbName)
	if err != nil {
		return err
	}

	return db.QueryRow(data, sql, param...)
}

func (this *Orm) QueryObjectByPk(data interface{}, v interface{}) error {
	pkColumn, _, err := this.getPkCondition(data)
	if err != nil {
		return err
	}
	return this.QueryObject(data, pkColumn, v)
}

func (this *Orm) QueryObjects(data interface{}, dbName string, sql string, params ...interface{}) (int64, error) {
	db, err := this.GetDb(dbName)
	if err != nil {
		return 0, err
	}
	return db.QueryRows(data, sql, params...)
}

func (this *Orm) QueryObjectByCond(data interface{}, dbName string, table string, cond *Condition) error {
	db, err := this.GetDb(dbName)
	if err != nil {
		return err
	}
	return db.QueryRowByCond(data, table, cond)
}

func (this *Orm) QueryObjectsByCond(data interface{}, dbName string, table string, cond *Condition) (int64, error) {
	db, err := this.GetDb(dbName)
	if err != nil {
		return 0, err
	}
	return db.QueryRowsByCond(data, table, cond)
}

func (this *Orm) doWrite(data interface{}, typ int, writeFunc func(db *Db, table string, val interface{}) (*QueryRet, error)) (*QueryRet, error) {
	dbName := getDbNameForData(data)
	table := getTableNameForData(data)

	//before insert
	err := this.beforeWrite(data, typ)
	if err != nil {
		return nil, err
	}

	db, err := this.GetDb(dbName)
	if err != nil {
		return nil, err
	}

	ret, err := writeFunc(db, table, data)
	if err != nil {
		return ret, err
	}

	//after insert
	err = this.afterWrite(data, typ)
	return ret, err
}

func (this *Orm) Insert(data interface{}) (*QueryRet, error) {
	return this.doWrite(data, DB_WRITE_TYPE_INSERT, func(db *Db, table string, data interface{}) (*QueryRet, error) {
		fields, values, err := getInsertFields(data, "", "")
		if err != nil {
			return nil, err
		}
		return db.Insert(table, fields, values...)
	})
}

func (this *Orm) InsertIgnore(data interface{}) (*QueryRet, error) {
	return this.doWrite(data, DB_WRITE_TYPE_INSERT, func(db *Db, table string, data interface{}) (*QueryRet, error) {
		fields, values, err := getInsertFields(data, "", "")
		if err != nil {
			return nil, err
		}
		return db.InsertIgnore(table, fields, values...)
	})
}

func (this *Orm) InsertOrUpdate(data interface{}) (*QueryRet, error) {
	return this.doWrite(data, DB_WRITE_TYPE_INSERT, func(db *Db, table string, data interface{}) (*QueryRet, error) {
		fields, values, err := getSetFields(data, "", "")
		if err != nil {
			return nil, err
		}
		sql := fmt.Sprintf("INSERT INTO `%s` SET %s ON DUPLICATE KEY UPDATE %s", table, fields, fields)
		values = append(values, values...)
		return db.Exec(sql, values...)
	})
}

//批量插入
//tmpl 批量插入的一个元素的指针，用于orm自动映射表和db
//daa 批量插入的数据
func (this *Orm) BatchInsert(tmpl interface{}, data interface{}, fields string, ignoreDuplicate bool) (int64, error) {
	val := reflect.ValueOf(data)
	ind := reflect.Indirect(val)
	if val.Kind() != reflect.Ptr || ind.Kind() != reflect.Slice {
		return 0, errors.New("Param data must be a slice pointer")
	}

	dbName := getDbNameForData(tmpl)
	db, err := this.GetDb(dbName)
	if err != nil {
		return 0, err
	}

	table := getTableNameForData(tmpl)

	fields, values, err := getBatchFieldValues(tmpl, data, fields, "")
	if err != nil {
		return 0, err
	}
	return db.BatchInsert(table, fields, values, ignoreDuplicate)
}

//根据主键更新对象
//data 对象指针
//fields 需要更新的字段.
//       传入空字符串时，根据ormInfo配置信息跟新所有字段；
//       更新多个字段时使用逗号分隔，如："Id,Name", 字段名字必须是model对象的字段名，而不是数据库的列名，大小写敏感
func (this *Orm) Update(data interface{}, fields string, ingoreFields string) (int64, error) {
	ret, err := this.doWrite(data, DB_WRITE_TYPE_DELETE, func(db *Db, table string, data interface{}) (*QueryRet, error) {
		where, val, err := this.getPkCondition(data)
		if err != nil {
			return nil, err
		}
		setStr, setValues, err := getSetFields(data, fields, ingoreFields)
		if err != nil {
			return nil, err
		}
		if setStr == "" {
			return nil, nil
		}
		return db.Update(table, setStr, where, append(setValues, val)...)
	})
	if err != nil {
		return 0, err
	}
	return ret.RowsAffected(), nil
}

func (this *Orm) UpdateIgnore(data interface{}, ingoreFields string) error {
	_, err := this.doWrite(data, DB_WRITE_TYPE_DELETE, func(db *Db, table string, data interface{}) (*QueryRet, error) {
		where, val, err := this.getPkCondition(data)
		if err != nil {
			return nil, err
		}
		setStr, setValues, err := getSetFields(data, "", ingoreFields)
		if err != nil {
			return nil, err
		}
		return db.Update(table, setStr, where, append(setValues, val)...)
	})
	return err
}

func (this *Orm) UpdateWhere(dbName string, table string, setField string, where string, params ...interface{}) (int64, error) {
	db, err := this.GetDb(dbName)
	if err != nil {
		return 0, err
	}

	ret, err := db.Update(table, setField, where, params...)
	if err != nil {
		return 0, err
	}
	return ret.RowsAffected(), err
}

func (this *Orm) UpdateByCond(dbName string, table string, cond *Condition, setField string,  setValues ...interface{}) (int64, error) {
	db, err := this.GetDb(dbName)
	if err != nil {
		return 0, err
	}

	ret, err := db.UpdateByCond(table, cond, setField, setValues...)
	if err != nil {
		return 0, err
	}
	return ret.RowsAffected(), err
}

//根据主键删除
func (this *Orm) Delete(data interface{}) error {
	_, err := this.doWrite(data, DB_WRITE_TYPE_DELETE, func(db *Db, table string, data interface{}) (*QueryRet, error) {
		where, val, err := this.getPkCondition(data)
		if err != nil {
			return nil, err
		}
		return db.Delete(table, where, val)
	})

	return err
}

func (this *Orm)DeleteWhere(dbName string, table string, where string, params ...interface{}) (int64, error) {
	db, err := this.GetDb(dbName)
	if err != nil {
		return 0, err
	}

	ret, err := db.Delete(table, where, params...)
	if err != nil {
		return 0, err
	}
	return ret.RowsAffected(), err
}

func (this *Orm)DeleteByCond(dbName string, table string, cond *Condition) (int64, error) {
	db, err := this.GetDb(dbName)
	if err != nil {
		return 0, err
	}

	ret, err := db.DeleteByCond(table, cond)
	if err != nil {
		return 0, err
	}
	return ret.RowsAffected(), err
}

func (this *Orm) getPkCondition(data interface{}) (string, interface{}, error) {
	m := GetModelMapInfo(reflect.TypeOf(data))
	ipkField, ok := m["pk"]

	if !ok {
		ipkField = "Id"
	}
	pkField, _ := ipkField.(string)

	val, ok := getValFromField(data, pkField)
	if !ok {
		return "", nil, fmt.Errorf("Pk field `%s` not found for object :%s", ipkField, reflect.ValueOf(data).Elem())
	}

	ifdMap, ok := m["fd_map"]
	pkColumn := pkField
	if ok {
		fdMap, _ := ifdMap.(map[string]string)
		pkColName, ok := fdMap[pkColumn]
		if ok {
			pkColumn = pkColName
		}
	}

	return fmt.Sprintf("`%s`=?", pkColumn), val, nil
}

func (this *Orm) beforeWrite(data interface{}, typ int) error {
	val, ok := callMethod(data, "BeforeWrite", typ)
	if !ok { //未实现BeforeWrite方法
		return nil
	}

	if val == nil {
		return nil
	}
	err, _ := val.(error)
	return err
}

func (this *Orm) afterWrite(data interface{}, typ int) error {
	val, ok := callMethod(data, "AfterWrite", typ)
	if !ok { //未实现BeforeWrite方法
		return nil
	}

	if val == nil {
		return nil
	}
	err, _ := val.(error)
	return err
}

func (this *Orm) QueryCount(data interface{}, where string, param ...interface{}) (cnt int64, e error) {
	dbName := getDbNameForData(data)
	table := getTableNameForData(data)

	db, e := this.GetDb(dbName)
	if e != nil {
		return
	}

	return db.QueryCount(table, where, param...)
}

func (this *Orm) QueryCountByCond(data interface{}, cond *Condition) (cnt int64, e error) {
	dbName := getDbNameForData(data)
	table := getTableNameForData(data)
	db, err := this.GetDb(dbName)
	if err != nil {
		return 0, err
	}

	return db.QueryCountByCond(table, cond)
}
