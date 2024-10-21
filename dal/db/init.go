package db

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/dbresolver"
	"time"
	"wgxDouYin/pkg/viper"
	"wgxDouYin/pkg/zap"
)

var (
	db        *gorm.DB
	config    = viper.Init("db")
	zapLogger = zap.InitLogger()
	tables    = map[string]interface{}{
		"users":     &User{},
		"videos":    &Video{},
		"comments":  &Comment{},
		"relations": &FollowRelation{},
	}
)

func getDsn(driverWithRole string) string {
	userName := config.Viper.GetString(fmt.Sprintf("%s.username", driverWithRole))
	password := config.Viper.GetString(fmt.Sprintf("%s.password", driverWithRole))
	host := config.Viper.GetString(fmt.Sprintf("%s.host", driverWithRole))
	port := config.Viper.GetInt(fmt.Sprintf("%s.port", driverWithRole))
	dbName := config.Viper.GetString(fmt.Sprintf("%s.database", driverWithRole))
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", userName, password, host, port, dbName)
	return dsn
}

func init() {
	zapLogger.Info("MySQL server conncetion sucessful!")
	dsn1 := getDsn("mysql.source")
	var err error
	db, err = gorm.Open(mysql.Open(dsn1), &gorm.Config{
		Logger:                 logger.Default.LogMode(logger.Info),
		PrepareStmt:            true,
		SkipDefaultTransaction: true,
	})
	if err != nil {
		panic(err.Error())
	}
	dsn2 := getDsn("mysql.replica1")
	err = db.Use(dbresolver.Register(dbresolver.Config{
		Sources:           []gorm.Dialector{mysql.Open(dsn1)},
		Replicas:          []gorm.Dialector{mysql.Open(dsn2)},
		Policy:            dbresolver.RandomPolicy{},
		TraceResolverMode: false,
	}))
	if err != nil {
		panic(err.Error())
	}
	//if err := db.AutoMigrate(&User{}, &Video{}, &Comment{}, &FollowRelation{}); err != nil {
	//	zapLogger.Fatalln(err.Error())
	//}
	_db, err := db.DB()
	if err != nil {
		zapLogger.Fatalln(err.Error())
	}
	if _db != nil {
		_db.SetMaxIdleConns(1000)
		_db.SetMaxIdleConns(20)
		_db.SetConnMaxIdleTime(60 * time.Minute)
	}
}

func GetDB() *gorm.DB {
	return db
}
