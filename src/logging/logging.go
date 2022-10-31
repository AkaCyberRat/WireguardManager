package logging

import (
	"io"
	"os"

	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/sirupsen/logrus"
)

func Configure() {
	formatter := &nested.Formatter{
		HideKeys:        true,
		NoColors:        true,
		ShowFullLevel:   true,
		TimestampFormat: "2006-01-02 15:04:05.999 \t",
	}

	file, err := os.OpenFile("logs/logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0777)
	if err != nil {
		logrus.Fatalf("Failed to open/create log file: %v", err.Error())
	}

	logrus.SetOutput(io.MultiWriter(os.Stdout, file))
	logrus.SetFormatter(formatter)
	logrus.SetLevel(logrus.DebugLevel)

}
