package utils

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/whoisfisher/mykubespray/pkg/logger"
)

// Info logs an informational message
func Info(message string) {
	logger.GetLogger().Info(message)
}

// Warn logs a warning message
func Warn(message string) {
	logger.GetLogger().Warn(message)
}

// Error logs an error message
func Error(message string) {
	logger.GetLogger().Error(message)
}

// Fatal logs a fatal error message and exits
func Fatal(message string) {
	logger.GetLogger().Fatal(message)
}

// Panic logs a panic message and panics
func Panic(message string) {
	logger.GetLogger().Panic(message)
}

// SetLogLevel sets the log level of the global logger
func SetLogLevel(level logrus.Level) {
	logger.GetLogger().SetLevel(level)
}

// SetOutputType sets the output type of the global logger
func SetOutputType(outputType string) {
	viper.Set("log.output_type", outputType)
}

// SetLogFile sets the log file path of the global logger
func SetLogFile(logfile string) {
	viper.Set("log.logfile", logfile)
}

// ShutdownLogger shuts down the global logger
func ShutdownLogger() {
	// Perform any cleanup if needed
}
