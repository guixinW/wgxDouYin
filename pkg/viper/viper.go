package viper

import (
	"github.com/spf13/viper"
	"log"
)

type Config struct {
	Viper *viper.Viper
}

func Init(configName string) Config {
	config := Config{Viper: viper.New()}
	v := config.Viper
	v.SetConfigFile("yml")
	v.SetConfigName(configName)
	v.AddConfigPath("./config")
	v.AddConfigPath("../config")
	v.AddConfigPath("../../config")
	v.AddConfigPath("/app/config")
	if err := v.ReadInConfig(); err != nil {
		log.Fatalf("error is %v", err)
	}
	return config
}
