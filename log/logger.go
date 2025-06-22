package logging

import (
	"context"
	"os"
	"path/filepath"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type loggerCtxKey struct{}

var loggerKey = loggerCtxKey{}

var (
	infoFileName  string = "info.log"
	debugFileName string = "debug.log"
	errorFileName string = "error.log"
	// add new fileName here if requied...

)

// helper function to rotate and keep back up of files.
func fileRotation(filePath string) zapcore.WriteSyncer {

	w := zapcore.AddSync(&lumberjack.Logger{
		Filename:   filePath,
		MaxSize:    1, // megabytes
		MaxBackups: 3,
		MaxAge:     1, // days
	})
	return w
}

// createEncoder returns a zapcore.Encoder configured for console output.
// It uses the production encoder config with a custom time format (ISO8601) and sets the time key to "timestamp".
// This encoder formats log entries for human-readable console output.
func createFileEncoder() zapcore.Encoder {
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderCfg.TimeKey = "timestamp"
	encoderCfg.EncodeLevel = zapcore.CapitalLevelEncoder
	return zapcore.NewJSONEncoder(encoderCfg)
}

// createEncoder returns a zapcore.Encoder configured for console output.
// It uses the production encoder config with a custom time format (ISO8601) and sets the time key to "timestamp".
// This encoder formats log entries for human-readable console output.
func createConsoleEncoder() zapcore.Encoder {
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderCfg.TimeKey = "timestamp"
	encoderCfg.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderCfg.EncodeCaller = zapcore.ShortCallerEncoder
	encoderCfg.EncodeDuration = zapcore.StringDurationEncoder
	return zapcore.NewJSONEncoder(encoderCfg)
}

// createWriteSyncer creates a zapcore.WriteSyncer for the given log file path.
// It ensures that the parent directory exists and then opens the file in append mode with write permissions.
// If the file or directory cannot be created/opened, the function panics.
// The returned WriteSyncer is used by zapcore to write logs to the file.
func createWriteSyncer(filePath string) zapcore.WriteSyncer {
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		panic("failed to create log directory: " + err.Error())
	}
	w := fileRotation(filePath)

	return zapcore.AddSync(w)
}

func createFileCore(level zapcore.LevelEnabler, fileName string) zapcore.Core {
	curDir, _ := os.Getwd()
	filePath := filepath.Join(curDir, "logs", fileName)
	return zapcore.NewCore(createFileEncoder(), createWriteSyncer(filePath), level)
}

// createConsoleCore returns a zapcore.Core for console output (stdout).
func createConsoleCore(level zapcore.Level) zapcore.Core {
	return zapcore.NewCore(createConsoleEncoder(), zapcore.Lock(os.Stdout), level)
}

// Function that returns a bool and sets the exact level to log.
func createExactLevelEnabler(level zapcore.Level) zap.LevelEnablerFunc {
	return zap.LevelEnablerFunc(func(l zapcore.Level) bool {
		return l == level
	})
}

// Function that returns a bool and sets level according to range.
func createRangeLevelEnabler(levelFrom, levelTo zapcore.Level) zap.LevelEnablerFunc {
	return zap.LevelEnablerFunc(func(l zapcore.Level) bool {
		return l >= levelFrom && l <= levelTo
	})
}

// SetLogger configures and returns a new zap.Logger instance.
// It sets up separate log files for info, debug, and error levels, each with its own core.
// Additional cores can be added by appending to the `cores` slice.
// The logger includes caller information and stack traces for error-level logs.
func SetLogger() *zap.Logger {
	infoLevel := createRangeLevelEnabler(zapcore.InfoLevel, zapcore.WarnLevel)
	debugLevel := createExactLevelEnabler(zap.DebugLevel)
	// errorLevel := createExactOrGreaterLevelEnabler(zapcore.ErrorLevel)
	errorLevel := createExactLevelEnabler(zapcore.ErrorLevel)

	// Set all cores
	cores := []zapcore.Core{
		createFileCore(infoLevel, infoFileName),
		createFileCore(debugLevel, debugFileName),
		createFileCore(errorLevel, errorFileName),
		createConsoleCore(zapcore.InfoLevel),
		// add new core here if requied...
	}

	teeCore := zapcore.NewTee(cores...)

	// Returns logger
	return zap.New(teeCore, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
}

// Function to help to add logger key to context.
func AddContext(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// Returns a logger from the go context.
func fromContext(ctx context.Context) *zap.Logger {
	logger, ok := ctx.Value(loggerKey).(*zap.Logger)
	if !ok {
		panic(`logger not found in context: ensure logger is injected into context. 
		Either call "AddContext" function before using "Logger" function or use a middleware that does this for you.`)
	}
	return logger

}

// Client function used at handler level to get the logger from context.
// In case logger is not set in context this function will panic and abort the application.
// Either set the logger using 'AddContext' function before using 'Logger' or create a middleware
// that does this for you.
func Logger(c echo.Context) *zap.Logger {
	return fromContext(c.Request().Context())
}

var baseLogger *zap.Logger

func InitGlobalLogger() {
	baseLogger = SetLogger()
}

func BaseLogger() *zap.Logger {
	if baseLogger == nil {
		panic("Base logger is not initialized. Call InitGlobalLogger() early in main.go.")
	}
	return baseLogger
}
