package sqlite

import (
	"fmt"
	"os"
	"path/filepath"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func NewSqliteDb(filePath string) (*gorm.DB, error) {
	folderPath, _ := filepath.Split(filePath)
	if err := os.MkdirAll(folderPath, 0777); err != nil {
		return nil, fmt.Errorf("failed to create db folder: %v", err.Error())
	}

	db, err := gorm.Open(sqlite.Open(filePath), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("db error: %v", err.Error())
	}

	return db, nil
}
