package logger_utils

import (
	"io"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// // NewZapLogger configures an advanced production logger that outputs structured logs to console and a rolling file.
// func NewZapLogger(logDir, filePrefix string) (*zap.SugaredLogger, io.Closer) {
// 	env := os.Getenv("ENVIRONMENT")

// 	// 1. Configure the Lumberjack daily rolling file writer
// 	lumberjackLogger := &lumberjack.Logger{
// 		Filename:   logDir + "/" + filePrefix + ".log",
// 		MaxSize:    100,   // Megabytes before split (failsafe)
// 		MaxBackups: 30,    // Keep up to 30 historical log files
// 		MaxAge:     28,    // Days to retain old files
// 		Compress:   true,  // Compress old log files automatically
// 		LocalTime:  false, // Force UTC file naming and rotation calculations
// 	}

// 	// 2. Define how our log lines should look
// 	encoderConfig := zap.NewProductionEncoderConfig()
// 	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder        // Readable timestamps: 2026-05-24T14:32:01.000Z
// 	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder // Colored text for console (INFO, ERROR)

// 	// JSON encoder for files (highly search indexable for tools like ElasticSearch)
// 	fileEncoderConfig := encoderConfig
// 	fileEncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder // Plain text level for JSON files

// 	// 3. Create execution cores for both console and file destinations
// 	consoleCore := zapcore.NewCore(
// 		zapcore.NewConsoleEncoder(encoderConfig),
// 		zapcore.AddSync(os.Stdout),
// 		zap.DebugLevel, // Capture everything Debug and above in console
// 	)

// 	fileCore := zapcore.NewCore(
// 		zapcore.NewJSONEncoder(fileEncoderConfig),
// 		zapcore.AddSync(lumberjackLogger),
// 		zap.InfoLevel, // Only capture Info and above inside the files
// 	)

// 	// Combine both outputs into a single core engine
// 	combinedCore := zapcore.NewTee(consoleCore, fileCore)

// 	// 4. Instantiate the core logger, enabling line-number tracking
// 	logger := zap.New(combinedCore, zap.AddCaller(), zap.AddCallerSkip(0))

// 	// Return the "Sugared" version which allows simple, formatted syntax like .Infof() and .Errorf()
// 	return logger.Sugar(), lumberjackLogger
// }

// func NewZapLogger(logDir, filePrefix string) (*zap.SugaredLogger, io.Closer) {
// 	env := os.Getenv("ENVIRONMENT")

// 	lumberjackLogger := &lumberjack.Logger{
// 		Filename:   logDir + "/" + filePrefix + ".log",
// 		MaxSize:    100,
// 		MaxBackups: 30,
// 		MaxAge:     28,
// 		Compress:   true,
// 		LocalTime:  false,
// 	}

// 	// File encoder (JSON)
// 	fileEncoderConfig := zap.NewProductionEncoderConfig()
// 	fileEncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
// 	fileEncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

// 	fileCore := zapcore.NewCore(
// 		zapcore.NewJSONEncoder(fileEncoderConfig),
// 		zapcore.AddSync(lumberjackLogger),
// 		zap.InfoLevel,
// 	)

// 	var core zapcore.Core
// 	fmt.Println(env)
// 	if env == "production" {
// 		// File only
// 		core = fileCore
// 	} else {
// 		// Console encoder (pretty + colors)
// 		consoleEncoderConfig := zap.NewDevelopmentEncoderConfig()
// 		consoleEncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
// 		consoleEncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

// 		consoleCore := zapcore.NewCore(
// 			zapcore.NewConsoleEncoder(consoleEncoderConfig),
// 			zapcore.AddSync(os.Stdout),
// 			zap.DebugLevel,
// 		)

// 		core = zapcore.NewTee(consoleCore, fileCore)
// 	}

// 	logger := zap.New(core, zap.AddCaller())

// 	return logger.Sugar(), lumberjackLogger
// }

func NewZapLogger(logDir, filePrefix, env string) (*zap.SugaredLogger, io.Closer) {
	lumberjackLogger := &lumberjack.Logger{
		Filename:   filepath.Join(logDir, filePrefix+".log"),
		MaxSize:    100,
		MaxBackups: 30,
		MaxAge:     28,
		Compress:   true,
		LocalTime:  false,
	}

	// File encoder (JSON)
	fileEncoderConfig := zap.NewProductionEncoderConfig()
	fileEncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	fileEncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	fileCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(fileEncoderConfig),
		zapcore.AddSync(lumberjackLogger),
		zap.InfoLevel,
	)

	var core zapcore.Core

	if env == "production" {
		// File only
		core = fileCore
	} else {
		// Console encoder (pretty + colors)
		consoleEncoderConfig := zap.NewDevelopmentEncoderConfig()
		consoleEncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		consoleEncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

		consoleCore := zapcore.NewCore(
			zapcore.NewConsoleEncoder(consoleEncoderConfig),
			zapcore.AddSync(os.Stdout),
			zap.DebugLevel,
		)

		core = zapcore.NewTee(consoleCore, fileCore)
	}

	logger := zap.New(core, zap.AddCaller())

	return logger.Sugar(), lumberjackLogger
}
