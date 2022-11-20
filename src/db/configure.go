package db

import (
	"os"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var Instance *gorm.DB

func Configure() {
	var err error

	err = os.MkdirAll("/app/db/", 0777)
	if err != nil {
		logrus.Fatalf("Failed to create db path: %v", err.Error())
	}

	Instance, err = gorm.Open(sqlite.Open("/app/db/service.db"), &gorm.Config{})
	if err != nil {
		logrus.Fatal("Failed to open/create db: %v", err)
	}
}
