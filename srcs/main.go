package main

import (
	"fmt"
	"sync/atomic"
	"taskmaster/srcs/controller"
	_ "taskmaster/srcs/input"
	"taskmaster/srcs/parser"
	"taskmaster/srcs/signals"
	"time"
	// _ "github.com/chzyer/readline"
)

func main() {
	config := parser.Init("configs/basic.yml")

	controller.Init(config)
	controller.Controller(config)

	signals.Init()

	for signals.FinishProgram == 0 {
		if atomic.LoadInt32(&signals.ReloadProgram) == 1 {
			newConfig := parser.Init("configs/basic.yml")
			signals.DiffConfigs(config, newConfig)
			config = newConfig
			atomic.StoreInt32(&signals.ReloadProgram, 0)
		}
		fmt.Println("Doing things...")
		time.Sleep(time.Second / 2)
	}
}
