package logger

import (
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap"
	"os"
	"log"
	"path/filepath"
)

var g_logger *zap.Logger

type LogRoute struct {
	Type string `json:"type"`//日志类型file， console
	Format string `json:"format"`//日志格式，支持json， simple
	FileName string `json:"file_name"`//日志路径,对文件有效
	Rotate string `json:"rotate"`//日志滚动配置，支持两种，不滚动和按日滚动 [day|none],对文件有效
	Level int `json:"level"`//日志级别配置
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
		checkLevel := r.Level
		le := zap.LevelEnablerFunc(func(level zapcore.Level) bool{
			return level >= zapcore.Level(checkLevel)
		})

		coreList = append(coreList, zapcore.NewCore(encoder, ws, le))

	}

	core := zapcore.NewTee(coreList...)

	var fields []zap.Field
	for k, v := range config.CommonField {
		fields = append(fields, zap.String(k, v))
	}

	g_logger = zap.New(core, zap.Fields(fields...))
}

func FushLog()  {
	if g_logger!= nil {
		g_logger.Sync()
	}
}

