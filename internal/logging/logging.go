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
	FilePath        string
}

func PreConfigure() {
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

	//
	// Parse log levels from deps.
	//
	consLogLevel, err := logrus.ParseLevel(deps.ConsoleLogLevel)
	if err != nil {
		return fmt.Errorf("failed to parse console log level: %v", err.Error())
	}

	fileLogLevel, err := logrus.ParseLevel(deps.FileLogLevel)
	if err != nil {
		return fmt.Errorf("failed to parse file log level: %v", err.Error())
	}

	//
	// Prepare file writer for file logging.
	//
	if err = os.MkdirAll(deps.FilePath, 0777); err != nil {
		return fmt.Errorf("failed to create log dir: %v", err.Error())
	}

	fileWriter, err := os.OpenFile(deps.FilePath+"logs.txt", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0777)
	if err != nil {
		return fmt.Errorf("failed to open/create log file: %v", err.Error())
	}

	//
	// Set minimum level, then below override
	// levels for every log output. And remove standart output.
	//
	logrus.SetLevel(logrus.TraceLevel)
	logrus.SetOutput(io.Discard)

	//
	// Add self-written console and file outputs
	//
	logrus.AddHook(&writerHook{
		Writer:      os.Stdout,
		MinLogLevel: consLogLevel,
	})

	logrus.AddHook(&writerHook{
		Writer:      fileWriter,
		MinLogLevel: fileLogLevel,
	})

	return nil
}
