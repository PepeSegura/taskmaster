package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"taskmaster/srcs/parser"
)

func ExecCmd(path string) {
	var args_command []string = strings.Split(path, " ")

	cmd := exec.Command(args_command[0], args_command[1:]...)
	fmt.Println("CMD: ", cmd)

	cmd.Env = append(os.Environ(), "MY_VAR=some_value")

	cmd.Dir = "/"

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Store old system umask
	oldUmask := syscall.Umask(0)
	fmt.Printf("Old Umask: %o\n", oldUmask)

	// Change umask for the process
	syscall.Umask(024)

	// Start the command
	err := cmd.Start()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	// Restore umask
	restoredUmask := syscall.Umask(oldUmask)
	fmt.Printf("Restored Umask: %o\n", restoredUmask)

	fmt.Printf("Cmd: [%s] with pid: [%d] has started\n", path, cmd.Process.Pid)
	cmd.Wait()
	fmt.Println("command finished!!")
}

func main() {
	config := parser.Parser("configs/config.yml")

	// Check if 'nginx' is in the config and print the 'Cmd'
	if nginxConfig, exists := config.Programs["nginx"]; exists {
		fmt.Println("Nginx command:", nginxConfig.Cmd)
	} else {
		fmt.Println("No nginx program found in the config.")
	}
	ExecCmd("ls -la")

}

// func main() {
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
//         fmt.Println("line:", line)
//     }
// }
