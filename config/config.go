package config

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var (
	appConfig   *Config
	agentConfig *Agent
)

func InitialAppConfig() *Config {
	viper.SetConfigName("app")
	viper.AddConfigPath(".")
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		log.Panicf("Error reading config file, %s", err)
	}

	err := viper.Unmarshal(&appConfig)
	if err != nil {
		log.Panicf("Unable to decode config into struct, %v", err)
	}

	return appConfig
}

func InitialAgentConfig() *Agent {
	viper.SetConfigName("agent")
	viper.AddConfigPath(".")
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		log.Panicf("Error reading config file, %s", err)
	}

	err := viper.Unmarshal(&agentConfig)
	if err != nil {
		log.Panicf("Unable to decode config into struct, %v", err)
	}

	return agentConfig
}

func GetAgentConfig() *Agent {
	return agentConfig
}

func GetAppConfig() *Config {
	return appConfig
}
