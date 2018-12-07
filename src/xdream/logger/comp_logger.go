package logger

import (
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap"
)

type CompLogger struct{
	Name string `json:"name"`//组件名
	Enabled bool `json:"enabled"`//是否启用
	Level string `json:"level"`//日志级别
	zapLevel zapcore.Level
}

func (l *CompLogger)Write(p []byte) (n int, err error)  {
	if ZLogger == nil {
		return 0, nil
	}

	if !l.Enabled {
		return
	}

	switch l.zapLevel {
	case zapcore.DebugLevel:
		ZLogger.Debug("", zap.String("comp", l.Name), zap.ByteString("msg", p))
	case zapcore.InfoLevel:
		ZLogger.Info("", zap.String("comp", l.Name), zap.ByteString("msg", p))
	case zapcore.WarnLevel:
		ZLogger.Warn("", zap.String("comp", l.Name), zap.ByteString("msg", p))
	case zapcore.ErrorLevel:
		ZLogger.Error("", zap.String("comp", l.Name), zap.ByteString("msg", p))
	case zapcore.FatalLevel:
		ZLogger.Fatal("", zap.String("comp", l.Name), zap.ByteString("msg", p))
	}

	return
}

func (l *CompLogger)SetLevel(level string)  {
	if zl, ok :=  LevelMap[level]; ok {
		l.zapLevel = zl
	}
}