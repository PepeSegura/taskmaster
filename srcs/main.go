package main

import (
	// "fmt"
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

	config := parser.Init("configs/basic.yml")

	controller.Init(config)

	signals.Init()

	// Channel for shell thread
	commandChan := make(chan input.Command)
	ackChan := make(chan struct{})
	var wg sync.WaitGroup

	// Start shell in a separate thread
	wg.Add(1)
	go func() {
		defer wg.Done()
		input.RunShell(commandChan, ackChan)
	}()

	for signals.FinishProgram == 0 {
		if atomic.LoadInt32(&signals.ReloadProgram) == 1 {
			newConfig := parser.Init("configs/basic.yml")
			signals.DiffConfigs(config, newConfig)
			config = newConfig
			atomic.StoreInt32(&signals.ReloadProgram, 0)
		}
		// fmt.Println("Doing things...")
		input.CheckForCommands(commandChan, ackChan)
		time.Sleep(time.Second / 4)
	}
}
