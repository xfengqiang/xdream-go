package xorm

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"
)

func init() {
	tz, _ := time.LoadLocation("Aisa/Shanghai")
	SetDefaultTimeZone(tz)
}

func SetDebugModel(debug bool) {
	debugModel = debug
}

func SetDefaultTimeZone(tz *time.Location) {
	defaultTimeZone = tz
}

//=============================================================
func getDbNameForData(data interface{}) string {
	db, ok := getStrValFromMethod(data, "GetDbName")
	if !ok {
		return "default"
	}
	return db
}

func getTableNameForData(data interface{}) string {
	table, ok := getStrValFromMethod(data, "GetTableName")
	if !ok {
		val := reflect.ValueOf(data)
		ind := reflect.Indirect(val)
		return getMapTableFun(ind.Type().Name())
	}
	return table
}

func getStrValFromMethod(data interface{}, methodName string) (string, bool) {
	val, ok := callMethod(data, methodName)
	if !ok {
		return "", false
	}
	strVal, ok := val.(string)
	return strVal, ok
}

func callMethod(data interface{}, methodName string, params ...interface{}) (interface{}, bool) {
	val := reflect.ValueOf(data)
	if val.Kind() != reflect.Ptr {
		panic("object must be a pointer")
	}

	valIndi := reflect.Indirect(val)
	var ePtrVal reflect.Value
	if valIndi.Kind() == reflect.Slice {
		//This is wrong， impossiable
		return nil, false
	} else { //struct
		ePtrVal = val
	}

	m := ePtrVal.MethodByName(methodName)
	if !m.IsValid() {
		return nil, true
	}

	var retVal interface{}
	values := make([]reflect.Value, len(params))
	for i, p := range params {
		values[i] = reflect.ValueOf(p)
	}

	ret := m.Call(values)
	if len(ret) == 0 {
		return nil, true
	}

	ptrDbName := &retVal
	pVal := reflect.Indirect(reflect.ValueOf(ptrDbName))
	if pVal.CanSet() {
		pVal.Set(ret[0])
	} else {
		return nil, false
	}
	return retVal, true
}

func getValFromField(data interface{}, field string) (interface{}, bool) {
	val := reflect.ValueOf(data)
	valIndi := reflect.Indirect(val)

	if val.Kind() != reflect.Ptr || valIndi.Kind() != reflect.Struct {
		return nil, false
	}

	f := valIndi.FieldByName(field)
	if !f.IsValid() {
		return nil, false
	}

	var retVal interface{}
	ptrDbName := &retVal
	pVal := reflect.Indirect(reflect.ValueOf(ptrDbName))
	if pVal.CanSet() {
		pVal.Set(f)
	} else {
		return nil, false
	}
	return retVal, true
}

//Util methods
func GetFieldMap(etype reflect.Type) map[string]string {
	mapInfo := GetModelMapInfo(etype)
	if fieldMap, ok := mapInfo["fd_map"]; ok {
		return fieldMap.(map[string]string)
	}
	return map[string]string{}
}

func GetModelMapInfo(etype reflect.Type) map[string]interface{} {
	var ptrInd reflect.Value
	if etype.Kind() == reflect.Ptr {
		ptrInd = reflect.New(etype.Elem())
	} else {
		ptrInd = reflect.New(etype)
	}

	ind := reflect.Indirect(ptrInd)
	fullName := ind.Type().PkgPath() + "/" + ind.Type().Name()
	mapInfo, ok := modelMapCache[fullName]
	if ok {
		return mapInfo
	}

	mapInfo = map[string]interface{}{}
	m := ptrInd.MethodByName("OrmInfo")
	if m.IsValid() {
		ret := m.Call([]reflect.Value{})
		if len(ret) == 1 {
			ptrMap := &mapInfo
			ptrMapVal := reflect.ValueOf(ptrMap)
			mapInfoVal := reflect.Indirect(ptrMapVal)
			if mapInfoVal.CanSet() {
				mapInfoVal.Set(ret[0])
			}
		}
	}

	var fieldMap map[string]string
	fd_map, ok := mapInfo["fd_map"]
	if ok {
		fieldMap = fd_map.(map[string]string)
	} else {
		fieldMap = map[string]string{}
	}

	for i := 0; i < ind.NumField(); i++ {
		f := ind.Type().Field(i)
		if _, ok := fieldMap[f.Name]; !ok {
			fieldMap[f.Name] = getMapColumnFun(f.Name)
		}
	}

	mapInfo["fd_map"] = fieldMap
	_, ok = mapInfo["pk"]
	if !ok {
		mapInfo["pk"] = "Id"
	}
	_, ok = mapInfo["ignore"]
	if !ok {
		mapInfo["ignore"] = ""
	}
	modelMapCache[fullName] = mapInfo

	//    fmt.Println("Ignore:", mapInfo["ignore"])

	return mapInfo
}

func setStructValue(obj reflect.Value, values map[string]interface{}, fieldMap map[string]string) {
	if obj.Kind() == reflect.Ptr {
		obj = reflect.Indirect(obj)
	}

	for i := 0; i < obj.NumField(); i++ {
		f := obj.Field(i)
		fe := obj.Type().Field(i)

		colName, ok := fieldMap[fe.Name]
		if !ok {
			continue
		}

		v, ok := values[colName]
		if !ok {
			continue
		}
		//        fmt.Println(colName, "=>", v, "type", reflect.ValueOf(v).Type())
		setFieldValue(f, v)
	}
}

// set field value to row container
func setFieldValue(ind reflect.Value, valuePtr interface{}) {
	value := reflect.Indirect(reflect.ValueOf(valuePtr)).Interface()
	tz := GetTimeZone()
	switch ind.Kind() {
	case reflect.Bool:
		if value == nil {
			ind.SetBool(false)
		} else if v, ok := value.(bool); ok {
			ind.SetBool(v)
		} else {
			v, _ := StrTo(ToStr(value)).Bool()
			ind.SetBool(v)
		}

	case reflect.String:
		if value == nil {
			ind.SetString("")
		} else {
			ind.SetString(ToStr(value))
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if value == nil {
			ind.SetInt(0)
		} else {
			val := reflect.ValueOf(value)
			switch val.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				ind.SetInt(val.Int())
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				ind.SetInt(int64(val.Uint()))
			default:
				v, _ := StrTo(ToStr(value)).Int64()
				ind.SetInt(v)
			}
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if value == nil {
			ind.SetUint(0)
		} else {
			val := reflect.ValueOf(value)
			switch val.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				ind.SetUint(uint64(val.Int()))
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				ind.SetUint(val.Uint())
			default:
				v, _ := StrTo(ToStr(value)).Uint64()
				ind.SetUint(v)
			}
		}
	case reflect.Float64, reflect.Float32:
		if value == nil {
			ind.SetFloat(0)
		} else {
			val := reflect.ValueOf(value)
			switch val.Kind() {
			case reflect.Float64:
				ind.SetFloat(val.Float())
			default:
				v, _ := StrTo(ToStr(value)).Float64()
				ind.SetFloat(v)
			}
		}

	case reflect.Struct:
		if value == nil {
			ind.Set(reflect.Zero(ind.Type()))

		} else if _, ok := ind.Interface().(time.Time); ok {

			var str string
			switch d := value.(type) {
			case time.Time:
				d = d.In(tz)
				ind.Set(reflect.ValueOf(d))
			case []byte:
				str = string(d)
			case string:
				str = d
			}
			if str != "" {
				if len(str) >= 19 {
					str = str[:19]
					t, err := time.ParseInLocation(format_DateTime, str, tz)
					if err == nil {
						t = t.In(tz)
						ind.Set(reflect.ValueOf(t))
					}
				} else if len(str) >= 10 {
					str = str[:10]
					t, err := time.ParseInLocation(format_Date, str, tz)
					if err == nil {
						ind.Set(reflect.ValueOf(t))
					}
				}
			}
		}
	}
}

func getInsertFields(data interface{}, fields string, ignoreFields string) (string, []interface{}, error) {
	return getFieldsParams(data, false, fields, ignoreFields)
}

func getSetFields(data interface{}, fields string, ignoreFields string) (string, []interface{}, error) {
	return getFieldsParams(data, true, fields, ignoreFields)
}

//过滤掉不可更新的feilds
func GetFilterFields(fields string, ignore string) string {
	if ignore == "" {
		return fields
	}
	updateAry := GetConvertedFieldAry(fields)
	ignoreAry := GetConvertedFieldAry(ignore)
	ignoreMap := map[string]bool{}
	for _, v := range ignoreAry {
		ignoreMap[v] = true
	}
	var ret []string = []string{}
	for _, v := range updateAry {
		if _, ok := ignoreMap[v]; !ok {
			ret = append(ret, v)
		}
	}
	return strings.Join(ret, ",")
}

func GetConvertedFieldAry(fields string) []string {
	if fields == "" {
		return []string{}
	}
	parts := strings.Split(fields, ",")
	var retParts []string = []string{}
	for _, f := range parts {
		retParts = append(retParts, strings.Title(f))
	}
	return retParts
}

func GetConvertedField(fields string) string {
	if fields == "" {
		return fields
	}
	return strings.Join(GetConvertedFieldAry(fields), ",")
}

func getFieldsParams(data interface{}, isSet bool, specFields string, ignoreFields string) (string, []interface{}, error) {
	val := reflect.ValueOf(data)
	ind := reflect.Indirect(val)

	if val.Kind() != reflect.Ptr || ind.Kind() != reflect.Struct {
		return "", nil, errors.New("data must be a struct pointer")
	}
	m := GetModelMapInfo(ind.Type())
	iFieldMap, _ := m["fd_map"]

	fieldMap := iFieldMap.(map[string]string)
	_iIgnoreField, _ := m["ignore"].(string)
	if len(_iIgnoreField) > 0 {
		ignoreFields = ignoreFields + "," + _iIgnoreField
	}

	specFields = GetConvertedField(specFields)
	ignoreFieldAry := GetConvertedFieldAry(ignoreFields)
	ignoreMap := map[string]bool{}
	for _, v := range ignoreFieldAry {
		ignoreMap[v] = true
	}
	fields := []string{}
	values := []interface{}{}

	if len(specFields) > 0 { //更新ignore之外的特定字段
		fieldNames := strings.Split(specFields, ",")
		for _, fieldName := range fieldNames {
			fe, found := ind.Type().FieldByName(fieldName)
			if !found {
				return "", values, errors.New(fmt.Sprintf("data %s have no filed:%s", ind.Type(), fieldName))
			}

			if _, ok := ignoreMap[fe.Name]; ok { //忽略的字段
				continue
			}

			f := ind.FieldByName(fieldName)

			colName, ok := fieldMap[fe.Name]
			if !ok {
				colName = getMapColumnFun(fe.Name)
			}
			if isSet {
				fields = append(fields, fmt.Sprintf("`%s`=?", colName))
			} else {
				fields = append(fields, fmt.Sprintf("`%s`", colName))
			}

			v, err := valueForField(f)
			if err != nil {
				return "", values, nil
			}
			values = append(values, v)
		}
	} else { //更新除了ignoreField之外的所有字段
		for i := 0; i < ind.NumField(); i++ {
			f := ind.Field(i)
			fe := ind.Type().Field(i)

			if _, ok := ignoreMap[fe.Name]; ok { //忽略的字段
				continue
			}

			colName, ok := fieldMap[fe.Name]
			if !ok {
				colName = getMapColumnFun(fe.Name)
			}
			if isSet {
				fields = append(fields, fmt.Sprintf("`%s`=?", colName))
			} else {
				fields = append(fields, fmt.Sprintf("`%s`", colName))
			}

			v, err := valueForField(f)
			if err != nil {
				return "", values, nil
			}
			values = append(values, v)
		}

	}

	return strings.Join(fields, ","), values, nil
}

func getAllMapFields(data interface{}) ([]string, error) {
	val := reflect.ValueOf(data)
	ind := reflect.Indirect(val)

	if val.Kind() != reflect.Ptr || ind.Kind() != reflect.Struct {
		return nil, errors.New("data must be a struct pointer")
	}
	m := GetModelMapInfo(ind.Type())
	iIgnoreField := m["ignore"]
	ignoreField := iIgnoreField.(string)

	ret := make([]string, 0, ind.NumField())
	for i := 0; i < ind.NumField(); i++ {
		fe := ind.Type().Field(i)
		if !strings.Contains(ignoreField, fe.Name) {
			ret = append(ret, fe.Name)
		}
	}
	return ret, nil
}

func getBatchFieldValues(tmpl interface{}, datas interface{}, specFields string, ingnoreFields string) (string, [][]interface{}, error) {
	val := reflect.ValueOf(datas)
	sind := reflect.Indirect(val)

	if val.Kind() != reflect.Ptr || sind.Kind() != reflect.Slice {
		return "", nil, errors.New("data must be a struct pointer")
	}

	if sind.Len() < 1 {
		return "", nil, nil
	}

	fields, _, err := getInsertFields(tmpl, specFields, ingnoreFields)
	if err != nil {
		return "", nil, err
	}

	var valueFields []string
	if len(specFields) == 0 {
		valueFields, err = getAllMapFields(tmpl)
		if err != nil {
			return "", nil, err
		}
	} else {
		valueFields = strings.Split(specFields, ",")
	}

	values := make([][]interface{}, sind.Len())
	for i := 0; i < sind.Len(); i++ {
		elem := sind.Index(i)
		elem = reflect.Indirect(elem)
		rowValues := make([]interface{}, len(valueFields))
		for idx, fieldName := range valueFields {
			f := elem.FieldByName(fieldName)
			v, err := valueForField(f)
			if err != nil {
				return "", nil, err
			}
			//            fmt.Println(fieldName, ":", v)
			rowValues[idx] = v
		}
		values[i] = rowValues
	}

	return fields, values, nil
}

func valueForField(f reflect.Value) (interface{}, error) {
	if f.Kind() == reflect.Ptr {
		f = f.Elem()
	}

	var value interface{}
	switch f.Kind() {
	case reflect.Bool:
		value = f.Bool()
	case reflect.String:
		value = f.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		value = f.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		value = f.Uint()
	case reflect.Float64, reflect.Float32:
		value = f.Float()
	case reflect.Struct:
		value = f.Interface()
		//        if f.Type() == time.Time{
		//            value = f.
		//        }
	}
	return value, nil
}
