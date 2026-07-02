package config

import (
	"fmt"
	"shorturl/model"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitMySQL() error {
	dsn := AppConfig.MySQL.DSN
	if dsn == "" {
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			AppConfig.MySQL.User,
			AppConfig.MySQL.Password,
			AppConfig.MySQL.Host,
			AppConfig.MySQL.Port,
			AppConfig.MySQL.DBName,
		)
	}

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error),
	})
	if err != nil {
		return err
	}
	return DB.AutoMigrate(&model.ShortURL{}, &model.VisitLog{})
}
