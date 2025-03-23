package logger

import (
	"github.com/sirupsen/logrus"
)

// Logger интерфейс для логирования в приложении
type Logger interface {
	WithField(key string, value interface{}) Logger
	WithFields(fields logrus.Fields) Logger
	WithError(err error) Logger
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
}

// LogrusAdapter адаптирует logrus.Logger к интерфейсу Logger
type LogrusAdapter struct {
	logger *logrus.Logger
}

// NewLogrusAdapter создает новый адаптер для logrus логгера
func NewLogrusAdapter(logger *logrus.Logger) *LogrusAdapter {
	return &LogrusAdapter{logger: logger}
}

// WithField добавляет поле к логгеру
func (l *LogrusAdapter) WithField(key string, value any) Logger {
	return &LogrusEntryAdapter{entry: l.logger.WithField(key, value)}
}

// WithFields добавляет несколько полей к логгеру
func (l *LogrusAdapter) WithFields(fields logrus.Fields) Logger {
	return &LogrusEntryAdapter{entry: l.logger.WithFields(fields)}
}

// WithError добавляет ошибку к логгеру
func (l *LogrusAdapter) WithError(err error) Logger {
	return &LogrusEntryAdapter{entry: l.logger.WithError(err)}
}

// Debug логирует сообщение с уровнем Debug
func (l *LogrusAdapter) Debug(args ...any) {
	l.logger.Debug(args...)
}

// Info логирует сообщение с уровнем Info
func (l *LogrusAdapter) Info(args ...any) {
	l.logger.Info(args...)
}

// Warn логирует сообщение с уровнем Warn
func (l *LogrusAdapter) Warn(args ...interface{}) {
	l.logger.Warn(args...)
}

// Error логирует сообщение с уровнем Error
func (l *LogrusAdapter) Error(args ...interface{}) {
	l.logger.Error(args...)
}

// Fatal логирует сообщение с уровнем Fatal
func (l *LogrusAdapter) Fatal(args ...interface{}) {
	l.logger.Fatal(args...)
}

// Fatalf логирует отформатированное сообщение с уровнем Fatal
func (l *LogrusAdapter) Fatalf(format string, args ...interface{}) {
	l.logger.Fatalf(format, args...)
}

// LogrusEntryAdapter адаптирует logrus.Entry к интерфейсу Logger
type LogrusEntryAdapter struct {
	entry *logrus.Entry
}

// WithField добавляет поле к логгеру
func (l *LogrusEntryAdapter) WithField(key string, value interface{}) Logger {
	return &LogrusEntryAdapter{entry: l.entry.WithField(key, value)}
}

// WithFields добавляет несколько полей к логгеру
func (l *LogrusEntryAdapter) WithFields(fields logrus.Fields) Logger {
	return &LogrusEntryAdapter{entry: l.entry.WithFields(fields)}
}

// WithError добавляет ошибку к логгеру
func (l *LogrusEntryAdapter) WithError(err error) Logger {
	return &LogrusEntryAdapter{entry: l.entry.WithError(err)}
}

// Debug логирует сообщение с уровнем Debug
func (l *LogrusEntryAdapter) Debug(args ...interface{}) {
	l.entry.Debug(args...)
}

// Info логирует сообщение с уровнем Info
func (l *LogrusEntryAdapter) Info(args ...interface{}) {
	l.entry.Info(args...)
}

// Warn логирует сообщение с уровнем Warn
func (l *LogrusEntryAdapter) Warn(args ...interface{}) {
	l.entry.Warn(args...)
}

// Error логирует сообщение с уровнем Error
func (l *LogrusEntryAdapter) Error(args ...interface{}) {
	l.entry.Error(args...)
}

// Fatal логирует сообщение с уровнем Fatal
func (l *LogrusEntryAdapter) Fatal(args ...interface{}) {
	l.entry.Fatal(args...)
}

// Fatalf логирует отформатированное сообщение с уровнем Fatal
func (l *LogrusEntryAdapter) Fatalf(format string, args ...interface{}) {
	l.entry.Fatalf(format, args...)
}
