package sqlite

import (
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func NewSqliteConnection(filePath string) gorm.Dialector {
	folderPath, _ := filepath.Split(filePath)
	if err := os.MkdirAll(folderPath, 0777); err != nil {
		logrus.Fatal("Failed to create db folder:", err.Error())
	}

	return sqlite.Open(filePath)
}
