package main

import (
	// "fmt"
	"fmt"
	"sync"
	"sync/atomic"
	"taskmaster/srcs/controller"
	"taskmaster/srcs/input"
	"taskmaster/srcs/logging"
	"taskmaster/srcs/parser"
	"taskmaster/srcs/signals"
	"time"
	// _ "github.com/chzyer/readline"
)

func main() {
	logging.Init("/var/log/taskmaster")

	controller.Config = parser.Init("configs/basic.yml")
	logging.Info("--------------------------- Starting up... ---------------------------")

	controller.Init(controller.Config)
	signals.Init()

	commandChan := make(chan input.Command)
	ackChan := make(chan struct{})
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		input.RunShell(commandChan, ackChan)
	}()

	for signals.FinishProgram == 0 && input.FinishProgram == 0 {
		if atomic.LoadInt32(&signals.ReloadProgram) == 1 {
			newConfig := parser.Init("configs/basic.yml")
			signals.DiffConfigs(controller.Config, newConfig)
			controller.Config = newConfig
			atomic.StoreInt32(&signals.ReloadProgram, 0)
		}
		input.CheckForCommands(commandChan, ackChan)
		time.Sleep(time.Second / 4)
	}
	logging.Info("Exiting...")
	fmt.Println("Exiting...")
	controller.KillAll()
}
