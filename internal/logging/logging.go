// internal/logging/logging.go
package logging

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/sirupsen/logrus"
)

// Deps хранит параметры конфигурации логгирования, которые будут переданы в Configure.
type Deps struct {
	ConsoleLogLevel string // Уровень логов для консоли
	FileLogLevel    string // Уровень логов для файла
	FilePath        string // Путь к файлу логов
}

// SetTempConfiguration задаёт временную конфигурацию логгирования (вывод в консоль, форматирование и уровень Info).
func SetTempConfiguration() {
	formatter := &nested.Formatter{
		HideKeys:        true,
		NoColors:        true,
		ShowFullLevel:   true,
		TimestampFormat: "2006-01-02 15:04:05.000",
	}

	logrus.SetFormatter(formatter)
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.InfoLevel)
}

// Configure инициализирует логгер согласно переданным зависимостям.
// Создаёт директорию для файлов, открывает/создаёт файл логов,
// устанавливает уровни логирования для консоли и файлов, и добавляет собственный writerHook.
func Configure(deps Deps) error {
	//
	// Разбор уровней логов из зависимостей.
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
	// Подготовка writer‑а для логов в файл.
	//
	folderPath, _ := filepath.Split(deps.FilePath)
	if err = os.MkdirAll(folderPath, 0777); err != nil {
		return fmt.Errorf("failed to create log dir: %v", err.Error())
	}

	fileWriter, err := os.OpenFile(deps.FilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0777)
	if err != nil {
		return fmt.Errorf("failed to open/create log file: %v", err.Error())
	}

	//
	// Устанавливаем уровень Trace, чтобы последующие MinLogLevel‑ы
	// в writerHook могли «открепить» нужный минимум.
	// Отключаем стандартный вывод.
	//
	logrus.SetLevel(logrus.TraceLevel)
	logrus.SetOutput(io.Discard)

	//
	// Добавляем консольный и файловый output‑ы с помощью writerHook.
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
