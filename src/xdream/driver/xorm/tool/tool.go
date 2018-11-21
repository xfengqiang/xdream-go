package main

import (
	"fly/components/config"
	"fmt"
	"fly/components/xorm"
	"strings"
	"os"
	"io/ioutil"
	"errors"
	"flag"
)

var enableDebug bool

func fieldName(cname string)  string{
	parts := strings.Split(strings.Trim(cname,"`"), "_")
	for idx, _ := range parts {
		parts[idx] = strings.Title(parts[idx] )
	}
	return strings.Join(parts, "")
}

func fieldType(typ string, sign string)  string{
	typ = strings.ToLower(strings.TrimSpace(typ))
	idx := strings.LastIndex(typ, "(")
	typStr := ""
	if idx > 0 {
		typStr = typ[0:idx]
	}else{
		typStr = typ;
	}
	switch typStr {
	case "bigint":
		if "unsigned"==sign {
			return "uint64"
		}
		return "int64"
	case "tinyint":
		if "unsigned"==sign {
			return "uint8"
		}
		return "int8"
	case "int":
		fallthrough
	case "integer":
		if "unsigned"==sign {
			return "uint32"
		}
		return "int"
	case "smallint":
		if "unsigned"==sign {
			return "uint16"
		}
		return "int16"
	case "mediumint":
		if "unsigned"==sign {
			return "uint32"
		}
		return "int32"
	case "double":
		return "float64"
	case "float":
		return "float32"
	case "datetime":
		return "time.Time"
	case "timestamp":
		return "time.Time"
	case "char":
		fallthrough
	case "varchar":
		fallthrough
	case "tinyblob":
		fallthrough
	case "mediumblob":
		fallthrough
	case "longblob":
		fallthrough
	case "longtext":
		fallthrough
	case "varbinary":
		fallthrough
	case "binary":
		fallthrough
	case "text":
		return "string"
	}

	return ""
}

type filedItem struct {
	Name string
	Type string
	Cmt string
	IsPk bool
	AutoIncr bool
}

type generateConfig struct {
	DbName string
	Prefix string
	Table string
	Overwrite bool
	SavePath string
	Fields []filedItem
	PkFields []string
	Pkg string
}

func logInfo(info string, args...interface{})  {
	if enableDebug {
		fmt.Printf(info+"\n", args...)
	}
}
func generateClass(cfg generateConfig)  {
	//dbName string, prefix string, table string, fields []filedItem, pkFields []string, savePath string, overwite bool
	lines := []string{}
	className := cfg.Prefix+fieldName(cfg.Table)
	saveFile := cfg.SavePath+"/"+className+".go"
	if _, err := os.Stat(saveFile); err != nil {
		if cfg.Overwrite {
			fmt.Printf("[Info] Replaceing file %s\n", saveFile)
		}else {
			fmt.Printf("[Waring] file %s aready exsit, ingnored.\n", saveFile)
			return
		}
	}

	lines = append(lines, fmt.Sprintf("package %s\n", cfg.Pkg))

	imports := map[string]bool{}
	for _, f := range cfg.Fields {
		if f.Type == "time.Time" {
			imports["time"] = true
			break
		}
	}
	if len(imports) > 0 {
		lines = append(lines, "import (")
		for item, _ := range imports {
			lines = append(lines, fmt.Sprintf("\t\"%s\"",item))
		}
		lines = append(lines, ")\n")
	}

	lines = append(lines, fmt.Sprintf("type %s struct{", className))
	for _, f := range cfg.Fields {
		lines = append(lines, fmt.Sprintf("\t%s\t%s\t//%s",f.Name,f.Type, f.Cmt))
	}
	lines = append(lines, "}\n")

	lines = append(lines, fmt.Sprintf("func (this *%s)GetDbName() string{", className))
	lines = append(lines, fmt.Sprintf("\treturn \"%s\"", cfg.DbName))
	lines = append(lines, "}\n")

	lines = append(lines, fmt.Sprintf("func (this *%s)GetTableName() string{", className))
	lines = append(lines, fmt.Sprintf("\treturn \"%s\"", cfg.Table))
	lines = append(lines, "}\n")

	lines = append(lines, fmt.Sprintf("func (this *%s) OrmInfo() map[string]interface{} {", className))
	lines = append(lines, fmt.Sprintf("\tfieldMap := map[string]string{}\n"))
	lines = append(lines, fmt.Sprintf("\treturn map[string]interface{}{"))
	ignoreFields :=[]string{}
	for _, f := range cfg.Fields {
		if f.AutoIncr {
			ignoreFields = append(ignoreFields, f.Name)
			break
		}
	}
	lines = append(lines, fmt.Sprintf("\t\t\"ignore\": \"%s\",", strings.Join(ignoreFields, ",")))
	if len(cfg.PkFields) > 0 {
		lines = append(lines, fmt.Sprintf("\t\t\"pk\": \"%s\",", strings.Join(cfg.PkFields, ",")))
	}

	lines = append(lines, fmt.Sprintf("\t\t\"fd_map\": fieldMap,"))
	lines = append(lines, "\t}")

	lines = append(lines, "}")
	logInfo(strings.Join(lines,"\n"))

	//fmt.Printf(strings.Join(lines,"\n")) 
	err := writeFile(saveFile, strings.Join(lines,"\n"))
	if err != nil{
		fmt.Printf("[Error] failed go write file:%s err:%s\n", saveFile, err.Error())
	}
}

func writeFile(fileName string, content string)  (err error){
	return ioutil.WriteFile(fileName, []byte(content),  os.FileMode(0664))
}

func parseScheme(table, scheme string) (fields []filedItem, pkFields []string, err error){
	parts := strings.Split(scheme, "\n")
	//fmt.Println(scheme);
	fields = []filedItem{}
	pkFields = []string{}
	for _, line := range parts {
		line = strings.Trim(strings.TrimSpace(line), ",")
		if line[0:1] == "`" {
			lineParts := strings.Split(line, " ")
			var comment string
			cmtPos := strings.LastIndex(line, "COMMENT")
			if cmtPos > 0 {
				comment = strings.Trim(line[cmtPos+8:len(line)-1], "'")
			} else {
				comment = lineParts[0]
			}
			var field filedItem
			field.Name = fieldName(lineParts[0])
			sign := ""
			if len(lineParts)>=3 {
				sign = lineParts[2]
			}
			field.Type = fieldType(lineParts[1], sign)
			logInfo("line:%s => %s\t%s\n", line, field.Name, field.Type)
			field.Cmt = comment
			if field.Type != "" {
				fields = append(fields, field)
			} else {
				fmt.Printf("Parse field failed. table:%s line:%s\n", table, line)
			}

			if strings.LastIndex(line, "AUTO_INCREMENT") > 0 {
				field.AutoIncr = true
			}
		} else {
			if strings.LastIndex(line, "PRIMARY") >= 0 {
				idx := strings.LastIndex(line, "(")
				substr := strings.Trim(strings.TrimSpace(line[idx + 1:]), "()")
				parts := strings.Split(substr, ",")
				for _, part := range parts {
					pkFields = append(pkFields, fieldName(strings.Trim(part, "`")))
				}
			}
		}
	}
	return
}

func ensureDir(dir string)  (err error){
	info, err:=os.Stat(dir);
	if err!=nil {
		err = os.MkdirAll(dir, os.FileMode(0755))
		return
	}
	if !info.IsDir() {
		return errors.New("Create dir failed:"+dir)
	}
	return nil
}

func main()  {
	//classPrefix := "X"
	//dbName := "sports_mall"
	//tableName := "product_test"
	//savePath := "./model"
	//overwrite := true

	dbName := flag.String("db", "", "数据库名")
	tableName := flag.String("table", "*", "表名")
	savePath := flag.String("out", "./model", "代码保存路径")
	configPath := flag.String("config", ".", "配置文件路径")
	overwrite := flag.Bool("force", false, "是否覆盖已经存在的文件")
	prefix := flag.String("prefix", "", "类名前缀")
	pkgName := flag.String("pkg", "", "类所属的package,默认从savePath中获取")
	debug := flag.Bool("debug", false, "是否打印调试信息")
	enableDebug = *debug
	flag.Parse()
	if *dbName=="" {
		fmt.Printf("db参数不能为空")
		os.Exit(-1)
	}

	configFile := *configPath+"/config.json"
	if _, err:=os.Stat(configFile);err!=nil {
		fmt.Printf("未找到配置文件：%s\n",configFile)
		os.Exit(-2)
	}

	configKeys, err := config.Parse(*configPath)
	if err != nil {
		fmt.Println(err)
		return
	}

	var processed bool
	for _, configKey := range configKeys {
		if("xmysql" == configKey) {
			xorm.SetConfig(configKey,"","","", true)
			processed = true;
			break;
		}
	}

	//创建输出目录
	err = ensureDir(*savePath)
	if *pkgName=="" {
		*pkgName = "model"
		parts := strings.Split(*savePath, "/")
		for i:=len(parts)-1;i>=0;i--{
			if parts[i]!="/"{
				*pkgName =  parts[i]
				break
			}
		}
	}

	if err!=nil {
		fmt.Printf("%s \n", err.Error())
		return
	}

	db, err := xorm.GetDbConnByName(*dbName, true)
	if err != nil {
		fmt.Printf("Get db connection failed.%s\n", err.Error())
		return
	}

	tables := []string{}
	if *tableName=="" || *tableName=="*" {
		rows, err := db.Query(fmt.Sprintf("SHOW TABLES"))
		if err != nil {
			fmt.Printf("Tables query err %s\n", err.Error())
		}
		for rows.Next() {
			table := ""
			err = rows.Scan(&table)
			if err== nil {
				tables = append(tables, table)
			}else{
				fmt.Printf("Scan table name err %s\n", err.Error())
			}
		}
	}else{
		tables = strings.Split(*tableName, ",")
	}

	for _, table := range tables {
		fmt.Printf("processing table %s\n", table)
		rows, err := db.Query(fmt.Sprintf("SHOW CREATE TABLE `%s`", table))
		if err != nil {
			fmt.Printf("Table query err:%s\n", err.Error())
			continue
		}

		if rows.Next() {
			table, scheme := "",""
			// get RawBytes from data
			err = rows.Scan(&table, &scheme)
			if err != nil {
				fmt.Printf("scan err:%s\n", err.Error())
				continue
			}
			fields, pkFields, err := parseScheme(table, scheme)
			if err != nil {
				fmt.Printf("[Error]parse scheme failed. err:%s\n", err.Error())
				continue
			}

			cfg := generateConfig{
				DbName:*dbName,
				Prefix:*prefix,
				Table:table,
				Fields:fields,
				PkFields:pkFields,
				SavePath:*savePath,
				Overwrite:*overwrite,
				Pkg:*pkgName,
			}

			generateClass(cfg)
		}
	}

	if !processed {
		fmt.Printf("no mysql config found\n")
	}else {
		fmt.Printf("All finihsed\n")
	}
}
