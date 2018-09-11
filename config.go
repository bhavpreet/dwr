package main

var _configJSON = []byte(`{
	"redis": {
		"host" : "redis",
		"port" : "6379",
		"password": "",
		"db"      : 0,
		"prefix"  : "dwr_"
	},
	"bind": "localhost:30300"
}`)

type dwrConfig struct {
	Redis struct {
		Host     string `json:"host"`
		Port     string `json:"port"`
		Password string `json:"password"`
		DB       int    `json:"db"`
		Prefix   string `json:"prefix"`
	} `json:"redis"`
	Bind string `json:"bind"`
}
