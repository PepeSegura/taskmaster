package main

import (
	"fmt"
	"taskmaster/srcs/exec"
	_ "taskmaster/srcs/input"
	"taskmaster/srcs/parser"
	_ "taskmaster/srcs/signals"
	"time"
	// _ "github.com/chzyer/readline"
)

func main() {
	config := parser.Init("configs/basic.yml")

	// Check if 'nginx' is in the config and print the 'Cmd'
	if nginxConfig, exists := config.Programs["nginx"]; exists {
		fmt.Println("Nginx command:", nginxConfig.Cmd)
	} else {
		fmt.Println("No nginx program found in the config.")
	}
	exec.Init(config)

	command := exec.Cmd("echo testing")
	for {
		fmt.Println("Main loop")
		fmt.Printf("Process with PID %d exists.\n", command.Process.Pid)
		time.Sleep(1 * time.Second)
	}

	// fmt.Println(command)
	// input.Init()

}
