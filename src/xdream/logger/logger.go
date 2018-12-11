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


type Context struct {
	ReqId string
	RefId string
	LogId string
}

func ctxToLogFields(ctx *Context) []zapcore.Field  {
	if ctx == nil {
		return nil
	}
	//TODO 这里有优化的空间，可以缓存变量，避免重复创建
	return []zapcore.Field{zap.String("reqid", ctx.ReqId), zap.String("refid", ctx.RefId), zap.String("logid",  ctx.LogId)}
}

func ctxToLogInterface(ctx *Context) []interface{}  {
	if ctx == nil {
		return nil
	}
	//TODO 这里有优化的空间，可以缓存变量，避免重复创建
	return []interface{}{zap.String("reqid", ctx.ReqId), zap.String("refid", ctx.RefId), zap.String("logid",  ctx.LogId)}
}


//method with ctx
func Debug(ctx *Context, msg string, fields ...zapcore.Field)  {
	ZLogger.Debug(msg, append(ctxToLogFields(ctx), fields...)...)
}
func Info(ctx *Context, msg string, fields ...zapcore.Field)  {
	ZLogger.Info(msg, append(ctxToLogFields(ctx), fields...)...)
}
func Warn(ctx *Context, msg string, fields ...zapcore.Field)  {
	ZLogger.Warn(msg, append(ctxToLogFields(ctx), fields...)...)
}
func Error(ctx *Context, msg string, fields ...zapcore.Field)  {
	ZLogger.Error(msg, append(ctxToLogFields(ctx), fields...)...)
}
func Fatal(ctx *Context, msg string, fields ...zapcore.Field)  {
	ZLogger.Fatal(msg, append(ctxToLogFields(ctx), fields...)...)
}

//sugar log methods
func SDebug(ctx *Context,fields...interface{})  {
	ZLogger.Sugar().With(ctxToLogInterface(ctx)...).Debug(fields...)
}

func SInfo(ctx *Context,fields...interface{})  {
	ZLogger.Sugar().With(ctxToLogInterface(ctx)...).Info(fields...)
}

func SWarn(ctx *Context,fields...interface{})  {
	ZLogger.Sugar().With(ctxToLogInterface(ctx)...).Warn(fields...)
}

func SError(ctx *Context,fields...interface{})  {
	ZLogger.Sugar().With(ctxToLogInterface(ctx)...).Error(fields...)
}

func SFatal(ctx *Context,fields...interface{})  {
	ZLogger.Sugar().With(ctxToLogInterface(ctx)...).Fatal(fields...)
}

func SDebugf(ctx *Context,template string, fields...interface{})  {
	ZLogger.Sugar().With(ctxToLogInterface(ctx)...).Debugf(template, fields...)
}
func SInfof(ctx *Context,template string, fields...interface{})  {
	ZLogger.Sugar().With(ctxToLogInterface(ctx)...).Infof(template, fields...)
}
func SWarnf(ctx *Context,template string, fields...interface{})  {
	ZLogger.Sugar().With(ctxToLogInterface(ctx)...).Warnf(template, fields...)
}
func SErrorf(ctx *Context,template string, fields...interface{})  {
	ZLogger.Sugar().With(ctxToLogInterface(ctx)...).Errorf(template, fields...)
}
func SFatalf(ctx *Context,template string, fields...interface{})  {
	ZLogger.Sugar().With(ctxToLogInterface(ctx)...).Fatalf(template, fields...)
}
