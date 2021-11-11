package server

import (
	"log"

	"github.com/gofiber/fiber/v2"
	cfg "github.com/qwlt/gmcollector/app/config"
	"github.com/qwlt/gmcollector/app/server/handlers"
	"github.com/qwlt/gmcollector/app/server/middlewares"
	"github.com/spf13/viper"
)

func NewServer() *fiber.App {
	conf := ServerConfig{}
	if viper.IsSet("server") {
		err := viper.UnmarshalKey("server", &conf)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		log.Fatal(&cfg.ViperKeyNotFoundError{Key: "server", Config: viper.ConfigFileUsed()})
	}

	server := fiber.New(
		fiber.Config{
			AppName:     "Data collector",
			Concurrency: conf.MaxConnections,
		},
	)
	handlers.InitValidator()
	SetupRoutes(server)
	return server
}

type ServerConfig struct {
	Host           string `mapstruct:"host"`
	Port           string `mapstruct:"port"`
	MaxConnections int    `mapstruct:"maxConnections"`
	BodyLimit      int    `mapstruct:"maxBodySize"`
}

func SetupRoutes(app *fiber.App) error {
	// TODO remove always pass filter before build
	app.Use(
		middlewares.New(
			middlewares.Config{
				Filter: middlewares.AlwaysPassFilter,
			},
		),
	)
	app.Add("get", "/", handlers.MainHandler)
	app.Add("post", "/test", handlers.TestHandler)
	// app.Add("post", "/test", handlers.AnotherHandler)
	return nil
}
