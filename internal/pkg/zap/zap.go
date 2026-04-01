package zap

import (
	"os"

	zapv2 "github.com/go-kratos/kratos/contrib/log/zap/v2"

	"github.com/tpl-x/kratos/internal/conf"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// NewLoggerWithLumberjack returns a HandlerFunc that adds a zap logger to the context.
// 统一使用 JSON 输出，避免控制台格式与 log.With 注入的 ts/caller 等字段重复。
func NewLoggerWithLumberjack(logConfig *conf.Log) *zapv2.Logger {
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	lumberjackLogger := &lumberjack.Logger{
		Filename:   logConfig.LogPath,
		MaxSize:    int(logConfig.MaxSize),
		MaxBackups: int(logConfig.MaxKeepFiles),
		MaxAge:     int(logConfig.MaxKeepDays),
		Compress:   logConfig.Compress,
	}
	writeSyncer := zapcore.NewMultiWriteSyncer(
		zapcore.AddSync(os.Stdout),
		zapcore.AddSync(lumberjackLogger),
	)
	encoder := zapcore.NewJSONEncoder(encoderCfg)
	var logLevel zapcore.Level
	logLevel = convertInnerLogLevelToZapLogLevel(logConfig.LogLevel, logLevel)
	core := zapcore.NewCore(
		encoder,
		writeSyncer,
		zap.NewAtomicLevelAt(logLevel),
	)
	// AddCallerSkip(2)：跳过 zap 内部 + kratos contrib/log/zap 封装，caller 指向实际业务代码（如 biz/evaluator_usecase.go:545）
	zapCore := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(3))
	zapLog := zapv2.NewLogger(zapCore)
	defer func() { _ = zapLog.Sync() }()
	return zapLog
}

// convertServerLogLevel converts server log level to zap log level.
func convertInnerLogLevelToZapLogLevel(svrLogLevel conf.LogLevel, logLevel zapcore.Level) zapcore.Level {
	switch svrLogLevel {
	case conf.LogLevel_Debug:
		logLevel = zapcore.DebugLevel
	case conf.LogLevel_Info:
		logLevel = zapcore.InfoLevel
	case conf.LogLevel_Warn:
		logLevel = zapcore.WarnLevel
	case conf.LogLevel_Error:
		logLevel = zapcore.ErrorLevel
	case conf.LogLevel_Fatal:
		logLevel = zapcore.FatalLevel
	}
	return logLevel
}

