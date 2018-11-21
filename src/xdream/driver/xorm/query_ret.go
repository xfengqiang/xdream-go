package xorm

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
)

var (
	modelMapCache map[string]map[string]interface{}
)

func init() {
	modelMapCache = map[string]map[string]interface{}{}
}

type QueryRet struct {
	queryRetType int
	rowValues    []map[string]interface{}
	sqlRet       sql.Result
	err          error
}

func NewQueryRet(typ int, values []map[string]interface{}) *QueryRet {
	return &QueryRet{queryRetType: typ, rowValues: values, err: nil}
}

func NewQueryRetWithSqlRet(typ int, sqlRet sql.Result, err error) *QueryRet {
	return &QueryRet{queryRetType: typ, sqlRet: sqlRet, err: err}
}

func NewQueryRetWithError(typ int, err error) *QueryRet {
	return &QueryRet{queryRetType: typ, err: err}
}

func (this *QueryRet) RowCount() int {
	return len(this.rowValues)
}

func (this *QueryRet) FetchRow(data interface{}) error {
	if this.err != nil {
		return this.err
	}

	val := reflect.ValueOf(data)
	if val.Kind() != reflect.Ptr {
		return errors.New("Fetch row param must be ptr")
	}

	datas := []interface{}{}
	cnt, err := this.doFetchRows(&datas, val.Elem().Type())
	if err != nil {
		return err
	}
	if cnt <= 0 {
		return sql.ErrNoRows
	}

	ind := reflect.Indirect(val)
	if ind.CanSet() {
		v := reflect.ValueOf(datas[0])
		ind.Set(v)
	}
	return nil
}

func (this *QueryRet) FetchRows(data interface{}) (int64, error) {
	if this.err == ErrNoRows {
		return 0, nil
	} else if this.err != nil {
		return 0, this.err
	}
	return this.doFetchRows(data, nil)
}

func (this *QueryRet) doFetchRows(data interface{}, etype reflect.Type) (int64, error) {
	val := reflect.ValueOf(data)
	sInd := reflect.Indirect(val)
	if val.Kind() != reflect.Ptr || sInd.Kind() != reflect.Slice {
		return 0, errors.New(fmt.Sprintf("param must be a ptr that points to a slice"))
	}

	if etype == nil {
		etype = sInd.Type().Elem()
	}

	fieldMap := GetFieldMap(etype)

	ret := sInd
	if !ret.IsNil() && this.RowCount() > 0 {
		ret.Set(reflect.New(sInd.Type()).Elem())
	}

	isSingle := true
	if this.queryRetType == QueryRetRows {
		isSingle = false
	}

	var ind reflect.Value
	var cnt int64
	for _, columnsMp := range this.rowValues {
		if etype.Kind() == reflect.Ptr {
			ind = reflect.New(etype.Elem())
		} else {
			ind = reflect.New(etype)
		}
		setStructValue(ind, columnsMp, fieldMap)

		if etype.Kind() != reflect.Ptr {
			ind = ind.Elem()
		}
		ret = reflect.Append(ret, ind)
		cnt++

		if isSingle {
			break
		}
	}

	sInd.Set(ret)
	return cnt, nil
}

func (this *QueryRet) RowsAffected() int64 {
	if this.err != nil {
		return 0
	}
	if this.queryRetType == QueryRetExec {
		cnt, _ := this.sqlRet.RowsAffected()
		return cnt
	}
	return 0
}

//
func (this *QueryRet) LastInertId() int64 {
	if this.err != nil {
		return 0
	}

	if this.queryRetType == QueryRetExec {
		id, _ := this.sqlRet.LastInsertId()
		return id
	}
	return 0
}
