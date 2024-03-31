package main

import (
	"github.com/alexjoedt/echosight/internal/app"
	"github.com/alexjoedt/echosight/internal/logger"
)

func main() {
	err := app.Run()
	if err != nil {
		logger.Fatalf("%v", err)
	}
}
