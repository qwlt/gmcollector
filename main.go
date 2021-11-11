package main

import (
	"github.com/qwlt/gmcollector/app"
)

func main() {
	App := app.NewApplication("viper")
	err := App.Init()
	if err != nil {
		panic(err)
	}
	App.Run()

}

// TODO update tests
// TODO update config file
// TODO test docker image build pipeline
