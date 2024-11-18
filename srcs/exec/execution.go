package exec

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

func Cmd(path string) {
	var args_command []string = strings.Split(path, " ")

	cmd := exec.Command(args_command[0], args_command[1:]...)

	cmd.Env = append(os.Environ(), "MY_VAR=some_value")

	// cmd.Dir = "/"

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Store old umask and apply new one
	newUmask := 022
	oldUmask := syscall.Umask(newUmask)

	// Start the command
	err := cmd.Start()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	// Restore umask
	syscall.Umask(oldUmask)
	cmd.Wait()
	fmt.Println("command finished!!")
}
