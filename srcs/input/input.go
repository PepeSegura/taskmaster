package input

import (
	"fmt"
	"strings"
	"sync/atomic"

	"taskmaster/srcs/controller"
	"taskmaster/srcs/logging"
	"taskmaster/srcs/signals"

	"github.com/chzyer/readline"
)

var FinishProgram int32 = 0
var CheckCmd int32 = 0

type Command struct {
	Name string
	Args []string
}

func RunShell(commandChan chan Command, ackChan chan struct{}) {
	rl, err := readline.New("taskmaster> ")
	if err != nil {
		panic(err)
	}
	defer rl.Close()

	for {
		line, err := rl.Readline()
		if err != nil {
			atomic.StoreInt32(&FinishProgram, 1)
			close(commandChan)
			break
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		command := parts[0]
		args := parts[1:]

		atomic.StoreInt32(&CheckCmd, 1)
		commandChan <- Command{Name: command, Args: args}

		<-ackChan

		if command == "exit" {
			break
		}
	}
}

func CheckForCommands(commandChan chan Command, ackChan chan struct{}) {
	if CheckCmd == 0 {
		return
	}

	cmd, ok := <-commandChan
	if !ok {
		close(ackChan)
		atomic.StoreInt32(&FinishProgram, 1)
		return
	}

	logging.Info(fmt.Sprintf("shell: Executing command %s", cmd.Name))
	switch cmd.Name {
	case "help":
		fmt.Println("Available commands: help, status, start, stop, restart, reload, exit")
	case "status":
		controller.Status()
	case "start":
		if len(cmd.Args) == 0 {
			fmt.Printf("%s command requires an argument.\n", cmd.Name)
		} else {
			for _, arg := range cmd.Args {
				controller.Try2StopGroup(arg)
				controller.Try2StartGroup(arg)
			}
		}
	case "stop":
		if len(cmd.Args) == 0 {
			fmt.Printf("%s command requires an argument.\n", cmd.Name)
		} else {
			for _, arg := range cmd.Args {
				controller.Try2StopGroup(arg)
			}
		}
	case "restart":
		if len(cmd.Args) == 0 {
			fmt.Printf("%s command requires an argument.\n", cmd.Name)
		} else {
			for _, arg := range cmd.Args {
				controller.Try2StopGroup(arg)
				controller.Try2StartGroup(arg)
			}
		}
	case "reload":
		fmt.Println("Reloading...")
		atomic.StoreInt32(&signals.ReloadProgram, 1)
	case "exit":
		atomic.StoreInt32(&FinishProgram, 1)
		close(commandChan)
	default:
		fmt.Printf("Unknown command: %s\n", cmd.Name)
		fmt.Println("Available commands: help, status, start, stop, restart, reload, exit")
	}

	ackChan <- struct{}{}
	atomic.StoreInt32(&CheckCmd, 0)
}
