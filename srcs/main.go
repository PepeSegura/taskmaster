package main

import (
	"fmt"
	"os/exec"
	"time"

	"taskmaster/srcs/execution"
	_ "taskmaster/srcs/input"
	"taskmaster/srcs/parser"
	_ "taskmaster/srcs/signals"
	// _ "github.com/chzyer/readline"
)

func monitorCmd(cmd *exec.Cmd, done chan int) {
	// Wait for the command to complete
	if cmd == nil {
		done <- (-1)
	}
	err := cmd.Wait()
	if err != nil {
		fmt.Printf("Command: [%s] PID: [%d] finished with error: %v\n", cmd.Path, cmd.Process.Pid, err)
	} else {
		fmt.Printf("Command: [%s] PID: [%d] finished successfully!\n", cmd.Path, cmd.Process.Pid)
	}
	// Send the PID to the channel
	done <- cmd.Process.Pid
}

func main() {
	config := parser.Init("configs/basic.yml")

	execution.Init(config)

	// Create a channel to receive completion signals
	done := make(chan int, 15) // Buffered to handle up to 5 commands

	// Launch 5 commands in parallel
	var cmds []*exec.Cmd
	for i := 5; i < 10; i++ {
		cmd := execution.Cmd("sleep 4")
		time.Sleep(1 * time.Second / 2)
		cmds = append(cmds, cmd)

		// Start monitoring the command in a goroutine
		go monitorCmd(cmd, done)
	}

	// Wait for all commands to finish
	for index, cmd := range cmds {
		pid := <-done
		fmt.Printf("Command %s index %d with PID %d has completed.\n", cmd.Path, index, pid)
	}

	
	fmt.Println("All commands have completed.")
	fmt.Println("Infinite loop now")
	for {

	}
}
