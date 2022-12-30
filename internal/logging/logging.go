package logging

import (
	"fmt"
	"io"
	"os"

	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/sirupsen/logrus"
)

type Deps struct {
	ConsoleLogLevel string
	FileLogLevel    string
}

func TempConfig() {
	formatter := &nested.Formatter{
		HideKeys:        true,
		NoColors:        true,
		ShowFullLevel:   true,
		TimestampFormat: "2006-01-02 15:04:05.999 \t",
	}

	logrus.SetFormatter(formatter)
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.InfoLevel)
}

func Configure(deps Deps) error {

	consoleLogLevel, err := logrus.ParseLevel(deps.ConsoleLogLevel)
	if err != nil {
		return fmt.Errorf("failed to parse console log level from config: %v", err.Error())
	}

	fileLogLevel, err := logrus.ParseLevel(deps.FileLogLevel)
	if err != nil {
		return fmt.Errorf("failed to parse file log level from config: %v", err.Error())
	}

	err = os.MkdirAll("/app/log/", 0777)
	if err != nil {
		return fmt.Errorf("failed to create log dir: %v", err.Error())
	}

	logFile, err := os.OpenFile("/app/log/logs.txt", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0777)
	if err != nil {
		return fmt.Errorf("failed to open/create log file: %v", err.Error())
	}

	logrus.SetOutput(io.Discard)
	logrus.AddHook(&writerHook{
		Writer:      os.Stdout,
		MinLogLevel: consoleLogLevel,
	})
	logrus.AddHook(&writerHook{
		Writer:      logFile,
		MinLogLevel: fileLogLevel,
	})

	return nil
}
