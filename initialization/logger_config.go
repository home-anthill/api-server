package initialization

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// InitLogger function
func InitLogger() *zap.SugaredLogger {
	// Init logger
	logger := getLogger()
	logger.Info("InitLogger - called")
	return logger
}

func getLogger() *zap.SugaredLogger {
	// taken from https://codewithmukesh.com/blog/structured-logging-in-golang-with-zap/
	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.ISO8601TimeEncoder

	consoleEncoder := zapcore.NewConsoleEncoder(config)

	var core zapcore.Core
	if os.Getenv("ENV") == "testing" {
		// if "testing" skip file logger
		core = zapcore.NewTee(
			zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), zapcore.DebugLevel),
		)
	} else {
		fileEncoder := zapcore.NewJSONEncoder(config)
		core = zapcore.NewTee(
			zapcore.NewCore(fileEncoder, getLogFileWriter(), zapcore.DebugLevel),
			zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), zapcore.DebugLevel),
		)
	}
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	sugarLogger := logger.Sugar()
	return sugarLogger
}

func getLogFileWriter() zapcore.WriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   "./logs/app.log",
		MaxSize:    10, // in MB
		MaxBackups: 5,
		MaxAge:     30,
		Compress:   false,
	}
	return zapcore.AddSync(lumberJackLogger)
}
