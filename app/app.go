package app

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v4/pgxpool"
	cfg "github.com/qwlt/gmcollector/app/config"
	"github.com/qwlt/gmcollector/app/db"
	"github.com/qwlt/gmcollector/app/server"
	wb "github.com/qwlt/gmcollector/app/writebuffer"
	"github.com/spf13/viper"
)

type KeyNotFound struct {
	Key string
}

func (e *KeyNotFound) Error() string {
	return fmt.Sprintf("Key `%v` not found", e.Key)
}

var App *Application

type Application struct {
	Server         *fiber.App
	WriteBuffer    *wb.WriteBuffer
	PGPool         *pgxpool.Pool
	ConfigProvider string
}

func NewApplication(configProvider string) *Application {

	App := &Application{}
	App.ConfigProvider = configProvider
	return App
}

func (app *Application) InitWriteBuffer() error {
	wb, err := wb.GetBuffer()
	if err != nil {
		panic(err)
	}
	app.WriteBuffer = wb
	return nil

}

func (app *Application) InitDB() error {
	app.PGPool = db.GetDB()
	return nil
}

func (app *Application) InitSever() error {
	app.Server = server.NewServer()
	return nil
}

func (app *Application) Run() {

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		fmt.Println("Gracefully shutting down...")
		_ = app.Server.Shutdown()
		app.WriteBuffer.Shutdown()
	}()
	go app.WriteBuffer.RunDataHandler()
	host := viper.GetString("server.host")
	port := viper.GetString("server.port")

	if err := app.Server.Listen(fmt.Sprintf("%v:%v", host, port)); err != nil {
		log.Panic(err)
	}
	// Final flush
	app.WriteBuffer.FlushBuffer()

}

func (app *Application) GetWriteBuffer() *wb.WriteBuffer {
	return app.WriteBuffer
}

func (app *Application) GetConnPool() *pgxpool.Pool {
	return app.PGPool
}

// Initialize app components, read config, check db connection ,etc.
// Init function call order must be preserver to proper initialization
func (app *Application) Init() error {

	err := cfg.ReadConfig()
	if err != nil {
		log.Fatal(err)
	}
	err = app.InitDB()
	if err != nil {
		return err
	}
	err = app.InitWriteBuffer()
	if err != nil {
		return err
	}

	err = app.InitSever()
	if err != nil {
		return err
	}
	return nil
}
