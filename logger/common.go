package logger

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"path/filepath"
	"time"
)

var Logger *zap.Logger

type Config struct {
	LogDirectory string
	BaseLogName  string
}

func Initialize(conf Config) {

	baseLogFileName := filepath.Join(conf.LogDirectory, conf.BaseLogName)

	getCurrentLogFileName := func(levelPrefix string) string {
		return fmt.Sprintf("%s_%s_%s.log", baseLogFileName, time.Now().Format("2006-01-02"), levelPrefix)
	}

	fileLogInfo := newLumberjackLogger(getCurrentLogFileName("info"))
	fileLogError := newLumberjackLogger(getCurrentLogFileName("error"))

	fileLevelInfo := zap.NewAtomicLevel()
	fileLevelInfo.SetLevel(zap.InfoLevel)

	fileLevelError := zap.NewAtomicLevel()
	fileLevelError.SetLevel(zap.ErrorLevel)

	consoleEncoderConfig := zapcore.EncoderConfig{
		LevelKey:    "level",
		TimeKey:     "timestamp",
		CallerKey:   "caller",
		MessageKey:  "message",
		EncodeLevel: zapcore.CapitalLevelEncoder,
		EncodeTime:  customTimeEncoder,
	}

	consoleEncoder := zapcore.NewConsoleEncoder(consoleEncoderConfig)

	consoleCore := zapcore.NewCore(
		consoleEncoder,
		zapcore.AddSync(os.Stdout),
		zap.NewAtomicLevel(),
	)

	fileCoreInfo := newZapCore(fileLogInfo, fileLevelInfo, false)
	fileCoreError := newZapCore(fileLogError, fileLevelError, false)

	Logger = zap.New(zapcore.NewTee(
		fileCoreInfo,
		fileCoreError,
		consoleCore,
	))
}

func newLumberjackLogger(filename string) *lumberjack.Logger {
	return &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    100,   // Максимальный размер файла в мегабайтах перед ротацией
		MaxBackups: 7,     // Максимальное количество ротированных файлов (недельная ротация)
		MaxAge:     7,     // Максимальный возраст ротированных файлов в днях
		LocalTime:  true,  // Использовать локальное время
		Compress:   false, // Опциональное сжатие архивов (gzip)
	}
}

func newZapCore(logRotate *lumberjack.Logger, level zap.AtomicLevel, jsonEncoder bool) zapcore.Core {
	var encoder zapcore.Encoder
	if jsonEncoder {
		encoder = zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	} else {
		encoder = zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	}

	return zapcore.NewCore(
		encoder,
		zapcore.AddSync(logRotate),
		level,
	)
}

func customTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
}

func Info(msg string, fields ...zap.Field) {
	Logger.Info(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	Logger.Error(msg, fields...)
}
