package logging

import (
	"WireguardManager/src/config"
	"io"
	"os"

	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/sirupsen/logrus"
)

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

func Configure() {
	config := config.Get()
	consoleLogLevel, err := logrus.ParseLevel(config.ConsoleLogLevel)
	if err != nil {
		logrus.Fatalf("Failed to parse console log level from config: %v", err.Error())
	}

	fileLogLevel, err := logrus.ParseLevel(config.FileLogLevel)
	if err != nil {
		logrus.Fatalf("Failed to parse file log level from config: %v", err.Error())
	}

	err = os.MkdirAll("/app/log/", 0777)
	if err != nil {
		logrus.Fatalf("Failed to create log dir: %v", err.Error())
	}

	logFile, err := os.OpenFile("/app/log/logs.txt", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0777)
	if err != nil {
		logrus.Fatalf("Failed to open/create log file: %v", err.Error())
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

}
