package main

import (
  "github.com/natefinch/lumberjack"
  "go.uber.org/zap"
  "go.uber.org/zap/zapcore"
)

//var once sync.Once
//
//var singleInstance *zap.SugaredLogger

//func getInstance() *zap.SugaredLogger {
//  if singleInstance == nil {
//    once.Do(
//      func() {
//        fmt.Println("Creating single instance now.")
//        singleInstance = InitLogger()
//      })
//  } else {
//    fmt.Println("Single instance already created.")
//  }
//
//  return singleInstance
//}

func InitLogger() *zap.SugaredLogger {
  writerSyncer := getLogWriter()
  encoder := getEncoder()

  core := zapcore.NewCore(encoder, writerSyncer, zapcore.DebugLevel)

  logger := zap.New(core, zap.AddCaller())
  sugarLogger := logger.Sugar()
  return sugarLogger
}

func getEncoder() zapcore.Encoder {
  return zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
}

func getLogWriter() zapcore.WriteSyncer {
  lumberJackLogger := &lumberjack.Logger{
    Filename:   "./logs/api-server.log",
    MaxSize:    10, // in MB
    MaxBackups: 5,
    MaxAge:     30,
    Compress:   false,
  }
  return zapcore.AddSync(lumberJackLogger)
}
