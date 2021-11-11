package config

import (
	"fmt"
	"log"
	"sync"

	"github.com/spf13/viper"
)

var mu sync.Mutex

func ReadConfig() error {
	mu.Lock()
	defer mu.Unlock()
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$PWD/app")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Fatal("Config file not  config.yaml not found ")
		} else {
			log.Fatal(err)
		}
	}
	return nil
}

type ViperKeyNotFoundError struct {
	Key    string
	Config string
}

func (e *ViperKeyNotFoundError) Error() string {
	return fmt.Sprintf("Key %v not found in config file %v", e.Key, e.Config)

}
