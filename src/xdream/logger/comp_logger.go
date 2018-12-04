package logger

import (
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap"
)

type Logger struct{
	Component string //组件名
	Enabled bool //是否启用
	Level int //日志级别
}

func (l *Logger)Write(p []byte) (n int, err error)  {
	if g_logger == nil {
		return 0, nil
	}

	level := zapcore.Level(l.Level)

	switch level {
	case zapcore.DebugLevel:
		g_logger.Debug("", zap.String("component", l.Component), zap.ByteString("msg", p))
	case zapcore.InfoLevel:
		g_logger.Info("", zap.String("component", l.Component), zap.ByteString("msg", p))
	case zapcore.WarnLevel:
		g_logger.Warn("", zap.String("component", l.Component), zap.ByteString("msg", p))
	case zapcore.ErrorLevel:
		g_logger.Error("", zap.String("component", l.Component), zap.ByteString("msg", p))
	case zapcore.FatalLevel:
		g_logger.Fatal("", zap.String("component", l.Component), zap.ByteString("msg", p))
	}

	return
}
