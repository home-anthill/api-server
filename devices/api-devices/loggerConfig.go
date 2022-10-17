package main

import (
  "github.com/natefinch/lumberjack"
  "go.uber.org/zap"
  "go.uber.org/zap/zapcore"
  "os"
)

func InitLogger() *zap.SugaredLogger {
  // taken from https://codewithmukesh.com/blog/structured-logging-in-golang-with-zap/
  config := zap.NewProductionEncoderConfig()
  config.EncodeTime = zapcore.ISO8601TimeEncoder

  fileEncoder := zapcore.NewJSONEncoder(config)
  consoleEncoder := zapcore.NewConsoleEncoder(config)

  core := zapcore.NewTee(
    zapcore.NewCore(fileEncoder, getLogFileWriter(), zapcore.DebugLevel),
    zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), zapcore.DebugLevel),
  )
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
