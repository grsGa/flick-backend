package logger

import (
	"github.com/rs/zerolog"
	"os"
)

// Logger 定义日志接口
type Logger interface {
	Debug(msg string, keysAndValues ...interface{})
	Info(msg string, keysAndValues ...interface{})
	Warn(msg string, keysAndValues ...interface{})
	Error(msg string, keysAndValues ...interface{})
	Fatal(msg string, keysAndValues ...interface{})
	With(key string, value interface{}) Logger
}

// zerologLogger 实现 Logger 接口
type zerologLogger struct {
	logger zerolog.Logger
}

// NewLogger 创建新的日志实例
func NewLogger() Logger {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	return &zerologLogger{
		logger: logger,
	}
}

// Debug 输出调试级别日志
func (l *zerologLogger) Debug(msg string, keysAndValues ...interface{}) {
	l.log(l.logger.Debug(), msg, keysAndValues...)
}

// Info 输出信息级别日志
func (l *zerologLogger) Info(msg string, keysAndValues ...interface{}) {
	l.log(l.logger.Info(), msg, keysAndValues...)
}

// Warn 输出警告级别日志
func (l *zerologLogger) Warn(msg string, keysAndValues ...interface{}) {
	l.log(l.logger.Warn(), msg, keysAndValues...)
}

// Error 输出错误级别日志
func (l *zerologLogger) Error(msg string, keysAndValues ...interface{}) {
	l.log(l.logger.Error(), msg, keysAndValues...)
}

// Fatal 输出致命错误日志并退出程序
func (l *zerologLogger) Fatal(msg string, keysAndValues ...interface{}) {
	l.log(l.logger.Fatal(), msg, keysAndValues...)
}

// With 添加上下文信息到日志器
func (l *zerologLogger) With(key string, value interface{}) Logger {
	return &zerologLogger{
		logger: l.logger.With().Interface(key, value).Logger(),
	}
}

// log 处理日志
func (l *zerologLogger) log(event *zerolog.Event, msg string, keysAndValues ...interface{}) {
	// 成对处理键值对
	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 < len(keysAndValues) {
			key, ok := keysAndValues[i].(string)
			if !ok {
				key = "unknown"
			}
			event.Interface(key, keysAndValues[i+1])
		}
	}
	event.Msg(msg)
}
