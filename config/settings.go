package config

import (
	"fmt"
	"github.com/spf13/viper"
	"os"
)

var LLMHubConfig Conf
var ServiceConf *ServiceConfig

type OpenaiConfig struct {
	Key  string `json:"key"`
	Host string `json:"host"`
}

type RedisConfig struct {
	Url      string `json:"url"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	PassWord string `json:"passWord"`
}

type Conf struct {
	Openai OpenaiConfig `json:"openai"`
	Redis  RedisConfig  `json:"redis"`
}

type OpenAI struct {
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

// JWTConfig Mapping token configuration
type JWTConfig struct {
	SigningKey string `mapstructure:"key" json:"key"`
}

// Proxy proxy
type Proxy struct {
	Url string `mapstructure:"url" json:"url"`
}

type APIKey struct {
	Key string `mapstructure:"key" json:"key"`
}

type Auth struct {
	Authorization string `mapstructure:"authorization" json:"authorization"`
}

type MongoDB struct {
	Link string
}

type MQ struct {
	URI string `mapstructure:"uri" json:"uri"`
}

type Token struct {
	Address    []string `mapstructure:"address" json:"address"`
	QueryId    int      `mapstructure:"queryId" json:"queryId"`
	MaxCount   int      `mapstructure:"maxCount" json:"maxCount"`
	QueryLimit int      `mapstructure:"queryLimit" json:"queryLimit"`
}

type AI struct {
	OpenaiApiKey string `mapstructure:"openaiApiKey" json:"openaiApiKey"`
}

type Task struct {
	AiTaskTime          string `mapstructure:"ai_task_time" json:"ai_task_time"`
	TransactionTaskTime string `mapstructure:"transaction_task_time" json:"transaction_task_time"`
}

type ServiceConfig struct {
	Port    int         `mapstructure:"port" json:"port"`
	DB      MysqlConfig `mapstructure:"mysql" json:"mysql"`
	RedisDB RedisConfig `mapstructure:"redis" json:"redis"`
	JWTInfo JWTConfig   `mapstructure:"jwt" json:"jwt"`
	Proxy   Proxy       `mapstructure:"proxy" json:"proxy"`
	APIKey  APIKey      `mapstructure:"APIKey" json:"APIKey"`
	Auth    Auth        `mapstructure:"auth" json:"auth"`
	Mongo   MongoDB     `mapstructure:"mongo" json:"mongo"`
	Token   Token       `mapstructure:"token" json:"token"`
	AI      AI          `mapstructure:"AI" json:"AI"`
	Task    Task        `mapstructure:"task" json:"task"`
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
		configFile = "./config/config-pro.yaml"
		if dev == "debug" {
			configFile = "./config/config-dev.yaml"
		} else if dev == "local" {
			configFile = "./config/config-local.yaml"
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
	os.Setenv("OPENAI_API_KEY", ServiceConf.AI.OpenaiApiKey)
	fmt.Println("data:", ServiceConf)
}
