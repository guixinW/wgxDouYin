package main

import (
	"fmt"
	"wgxDouYin/pkg/viper"
)

var (
	config      = viper.Init("user")
	serviceName = config.Viper.GetString("server.name")
	serviceAddr = fmt.Sprintf("%s:%d", config.Viper.GetString("server.host"),
		config.Viper.GetInt("server.port"))
	etcdAddr = fmt.Sprintf("%s:%d", config.Viper.GetString("etcd.host"),
		config.Viper.GetInt("server.port"))
	signinKey = config.Viper.GetString("JWT.signinKey")
)

func main() {
	fmt.Printf("%v\n", serviceName)
	fmt.Printf("%v\n", serviceAddr)
	fmt.Printf("%v\n", etcdAddr)
	//for key, value := range config.Viper.AllSettings() {
	//	fmt.Printf("%s, %v\n", key, value)
	//}
}
