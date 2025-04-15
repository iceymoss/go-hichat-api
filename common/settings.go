package common

import (
	"fmt"
	"github.com/spf13/viper"
)

var ServiceConf *ServiceConfig

type RedisConfig struct {
	Url      string `json:"url"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	PassWord string `json:"passWord"`
}

// MysqlConfig mysql information configuration
type MysqlConfig struct {
	Host     string `mapstructure:"host" json:"host"`
	Port     int    `mapstructure:"port" json:"port"`
	Name     string `mapstructure:"name" json:"Name"`
	User     string `mapstructure:"user" json:"user"`
	Password string `mapstructure:"password" json:"password"`
	DbName   string `mapstructure:"dbname" json:"dbname"`
	LogLevel string `mapstructure:"logLevel" json:"logLevel"`
}

type MongoDB struct {
	Link string
}

type MQ struct {
	URI string `mapstructure:"uri" json:"uri"`
}

type ServiceConfig struct {
	DB      MysqlConfig `mapstructure:"mysql" json:"mysql"`
	RedisDB RedisConfig `mapstructure:"redis" json:"redis"`
	Mongo   MongoDB     `mapstructure:"mongo" json:"mongo"`
	MQ      MQ          `mapstructure:"mq"   json:"mq"`
}

func InitConfig(dev string, serveType string, configPath string) {
	//Instantiating an object
	v := viper.New()

	configFile := ""
	if serveType == "task" {
		configFile = "../../config/config-pro.yaml"
		if dev == "debug" {
			configFile = "../../config/config-dev.yaml"
		} else if dev == "local" {
			configFile = "../../config/config-local.yaml"
		}
	} else {
		configFile = "../config/config-pro.yaml"
		if dev == "debug" {
			configFile = "../config/config-dev.yaml"
		} else if dev == "local" {
			configFile = "../config/config-local.yaml"
		}
	}

	if configPath != "" {
		configFile = fmt.Sprintf("%s/config-%s.yaml", configPath, dev)
	}

	//Reading Configuration Files
	v.SetConfigFile(configFile)

	//Reading in a file
	if err := v.ReadInConfig(); err != nil {
		panic(err)
	}

	//How to use the ServerConf object in other files - global variables
	if err := v.Unmarshal(&ServiceConf); err != nil {
		panic(err)
	}
	fmt.Println("data:", ServiceConf)
}
