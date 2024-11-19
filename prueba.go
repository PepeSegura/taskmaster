package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

func Cmd(path string) *exec.Cmd {
	var argsCommand []string = strings.Split(path, " ")

	cmd := exec.Command(argsCommand[0], argsCommand[1:]...)

	cmd.Env = append(os.Environ(), "MY_VAR=some_value")

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Store old umask and apply new one
	newUmask := 022
	oldUmask := syscall.Umask(newUmask)

	// Start the command
	err := cmd.Start()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return nil
	}

	// Restore umask
	syscall.Umask(oldUmask)

	fmt.Println("Command started in the background!")
	return cmd
}

func monitorCmd(cmd *exec.Cmd, done chan bool) {
	// Wait for the command to complete
	err := cmd.Wait()
	if err != nil {
		fmt.Printf("Command finished with error: %v\n", err)
	} else {
		fmt.Println("Command finished successfully!")
	}

	// Notify that the command has finished
	done <- true
}

func main() {
	cmd := Cmd("sleep 5") // Replace with your command
	if cmd == nil {
		return
	}

	// Channel to signal when the command finishes
	done := make(chan bool)

	// Start monitoring the command in a goroutine
	go monitorCmd(cmd, done)

	// Simulate other tasks in the main program
	fmt.Println("Doing other work in the main program...")
	for {
		select {
		case <-done:
			fmt.Println("Background command completed!")
			return
		default:
			fmt.Println("Still working...")
			time.Sleep(1 * time.Second / 2)
		}
	}
}
