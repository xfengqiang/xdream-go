package logger

import (
	"go.uber.org/zap/zapcore"
	"log"
)

var LevelMap map[string]zapcore.Level = map[string]zapcore.Level{
	"debug":zapcore.DebugLevel,
	"info":zapcore.InfoLevel,
	"warn":zapcore.WarnLevel,
	"error":zapcore.ErrorLevel,
	"fatal":zapcore.FatalLevel,
}

var LevelStrMap map[zapcore.Level]string = map[zapcore.Level]string{
	zapcore.DebugLevel:"debug",
	zapcore.InfoLevel:"info",
	zapcore.WarnLevel:"warn",
	zapcore.ErrorLevel:"error",
	zapcore.FatalLevel:"fatal",
}


func StrToLevel(k string) zapcore.Level  {
	return LevelMap[k]
}

type LevelSetter struct {
	defaultLevel string
	currentLevel string
	levelSetter func (level string)
}

func NewLevelSetter(defaultLevel string, setter func(level string)) *LevelSetter  {
	return &LevelSetter{defaultLevel:defaultLevel, currentLevel:defaultLevel, levelSetter:setter}
}

func (ls *LevelSetter)SetLevelFunc(setter func(level string))  {
	ls.levelSetter = setter
}

func (ls *LevelSetter)Reset()   {
	if ls.levelSetter!= nil {
		ls.levelSetter(ls.defaultLevel)
		ls.currentLevel = ls.defaultLevel
	}
}

func (ls *LevelSetter)SetLevel(level string)  {
	if ls.levelSetter!= nil {
		ls.levelSetter(level)
		ls.currentLevel = level
	}
}

func (ls *LevelSetter)Level() string  {
	return ls.currentLevel
}

var levelLevelSetter map[string]*LevelSetter = map[string]*LevelSetter{}


func RegisterLevelSetter(key string, ls *LevelSetter)  {
	log.Println("register log level setter:", key)
	levelLevelSetter[key] = ls
}

func SetLevel(key string, level string) bool {
	if setter, ok := levelLevelSetter[key]; ok {
		setter.SetLevel(level)
		return true
	}
	return false
}

func ResetLevel(key string)  bool{
	if setter, ok := levelLevelSetter[key]; ok {
		setter.Reset()
		return  true
	}
	return  false
}

func GetLevelStatus() map[string]string {
	res := map[string]string{}
	for key, ls := range levelLevelSetter {
		res[key] = ls.Level()
	}
	return  res
}