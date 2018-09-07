package main

var _configJSON = []byte(`{
	"redis": {
		"address" : "redis:6379",
		"password": "",
		"db"      : 0,
		"prefix"  : "dwr_"
	},
	"bind": "localhost:30300"
}`)

type dwrConfig struct {
	Redis struct {
		Address  string `json:"address"`
		Password string `json:"password"`
		DB       int    `json:"db"`
		Prefix   string `json:"prefix"`
	} `json:"redis"`
	Bind string `json:"bind"`
}
