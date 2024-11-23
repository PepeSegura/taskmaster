package main

import (
	"taskmaster/srcs/controller"
	_ "taskmaster/srcs/input"
	"taskmaster/srcs/parser"
	_ "taskmaster/srcs/signals"
	// _ "github.com/chzyer/readline"
)

func main() {
	config := parser.Init("configs/basic.yml")

	controller.Init(config)
	controller.Controller(config)

}
