package input

import (
	"fmt"
	"strings"

	"github.com/chzyer/readline"
)

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
		if err != nil { // EOF
			close(commandChan)
			break
		}

		// Trim spaces
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Split command and arguments
		parts := strings.Fields(line)
		command := parts[0]
		args := parts[1:]

		// Send through the channel
		commandChan <- Command{Name: command, Args: args}

		// Wait for main thread to process
		<-ackChan

		// Exit if "exit"
		if command == "exit" {
			break
		}
	}
}

func CheckForCommands(commandChan chan Command, ackChan chan struct{}) {
	for {
		cmd, ok := <-commandChan
		if !ok {
			fmt.Println("Command channel closed. Exiting.")
			close(ackChan)
			return
		}

		switch cmd.Name {
		case "help":
			fmt.Println("Available commands: help, start, stop, restart, reload, exit")
		case "start", "stop", "restart":
			if len(cmd.Args) == 0 {
				fmt.Printf("%s command requires an argument.\n", cmd.Name)
			} else {
				fmt.Printf("Executing %s with argument: %s\n", cmd.Name, strings.Join(cmd.Args, " "))
			}
		case "reload":
			fmt.Println("Reloading...")
		case "exit":
			fmt.Println("Goodbye!")
			close(commandChan)
		default:
			fmt.Printf("Unknown command: %s\n", cmd.Name)
		}

		// tell shell reader command has been processed, print new prompt
		ackChan <- struct{}{}
	}
}

// func Init() {
//     rl, err := readline.New("readline> ")
//     if err != nil {
//         panic(err)
//     }
//     defer rl.Close()

//     for {
//         line, err := rl.Readline()
//         if err != nil { // io.EOF on Ctrl+D
//             break
// 		}
// 		fmt.Println(err)
//         fmt.Println("line: ", line)
//     }
// }
