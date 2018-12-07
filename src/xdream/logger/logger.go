package logger

import (
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap"
	"os"
	"log"
	"path/filepath"
)

var ZLogger *zap.Logger



type LogRoute struct {
	Type string `json:"type"`//日志类型file， console
	Format string `json:"format"`//日志格式，支持json， simple
	FileName string `json:"file_name"`//日志路径,对文件有效
	Rotate string `json:"rotate"`//日志滚动配置，支持两种，不滚动和按日滚动 [day|none],对文件有效
	Level string `json:"level"`//日志级别配置
	LevelKey string `json:"level_key"` //日志级别配置key
	//levelInt zapcore.Level
}

type LogConfig struct {
	CommonField map[string]string `json:"common_field"`//公共参数信息
	LogDir string `json:"log_dir"`
	Routes []LogRoute `json:"routes"`//日志路由配置
}

func ensureDir(path string) error {
	dir := filepath.Dir(path)
	return os.MkdirAll(dir, 0755)
}

func InitLogger(config *LogConfig)  {
	dirPath := config.LogDir
	if dirPath=="" {
		dirPath = "./"
	}

	var coreList []zapcore.Core

	encoderConfig := zap.NewProductionEncoderConfig()
	for _, r := range config.Routes {
		//encoder
		var encoder zapcore.Encoder
		if r.Format == "json" {
			encoder = zapcore.NewJSONEncoder(encoderConfig)
		}else{
			encoder = zapcore.NewConsoleEncoder(encoderConfig)
		}

		//write syncer
		var ws zapcore.WriteSyncer
		if r.Type == "file" {
			fullPath := filepath.Join(dirPath,r.FileName)

			var e error
			ws, e = NewRotateFile(r.Rotate, fullPath, true, nil)
			if e != nil {
				log.Println("[logger-init] create log file failed. Path "+fullPath+" error "+e.Error())
			}
		}else { //console
			ws = os.Stdout
		}

		//error log level
		if r.Level=="" {
			r.Level = "info" //default level
		}

		setLevel := LevelMap[r.Level]
		le := zap.NewAtomicLevelAt(setLevel)
		coreList = append(coreList, zapcore.NewCore(encoder, ws, le))
		if r.LevelKey != "" {
			RegisterLevelSetter(r.LevelKey, NewLevelSetter(r.Level, func(level string) {
				if l , ok := LevelMap[level]; ok {
					le.SetLevel(l)
				}
			}))
		}
	}

	core := zapcore.NewTee(coreList...)
	var fields []zap.Field
	for k, v := range config.CommonField {
		fields = append(fields, zap.String(k, v))
	}

	ZLogger = zap.New(core, zap.Fields(fields...))
}

func FushLog()  {
	if ZLogger != nil {
		ZLogger.Sync()
	}
}

