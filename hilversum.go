package main

import (
	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"
	"github.com/vaysman/hilversum/hijack_dns"
	"github.com/vaysman/hilversum/http_proxy"
	"github.com/vaysman/hilversum/web_interface"
)

var config = viper.New()

func loadDefaultSettings() {
	config.SetDefault("hijack_enable", false)
	config.SetDefault("log_dns_request", true)
	config.SetDefault("log_http_request", true)
	config.SetDefault("http_proxy_port", 8080)
	config.SetDefault("dns_port", 53)
}


func loadConfig() {
	config.SetConfigType("json")
	config.AddConfigPath(".")
	err := config.ReadInConfig()

	if err != nil { // Handle errors reading the config file
		jww.ERROR.Panicf("Fatal error config file: %s \n", err)
	}

}

func init() {
	jww.SetLogThreshold(jww.LevelTrace)
	jww.SetStdoutThreshold(jww.LevelInfo)
}

func main() {
	loadDefaultSettings()
	loadConfig()

	jww.INFO.Println("Welcome to Hilversum")
	// load config from json
	// start dns proxy
	hijackdns.Run(config)
	// start http(s) proxy
	httpproxy.Run(config)
	// start web interface
	webinterface.Run(config)
}