package logging

import (
	"io"

	"github.com/sirupsen/logrus"
)

type writerHook struct {
	Formatter   logrus.Formatter
	Writer      io.Writer
	MinLogLevel logrus.Level
}

func (hook *writerHook) Fire(entry *logrus.Entry) error {
	line, err := entry.String()
	if err != nil {
		return err
	}
	_, err = hook.Writer.Write([]byte(line))

	// data, err := hook.Formatter.Format(entry)
	// if err != nil {
	// 	return err
	// }

	// _, err = hook.Writer.Write(line)

	return err
}

// Levels define on which log levels this hook would trigger
func (hook *writerHook) Levels() []logrus.Level {
	logLevels := []logrus.Level{}

	for _, level := range logrus.AllLevels {
		if level <= hook.MinLogLevel {
			logLevels = append(logLevels, level)
		}
	}

	return logLevels
}
