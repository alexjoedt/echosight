package main

import (
	"log"

	"github.com/alexjoedt/echosight/internal/agent"
)

func main() {
	run()
}

func run() {
	err := agent.ListenAndServe(":8089")
	if err != nil {
		log.Panicln(err)
	}
}
