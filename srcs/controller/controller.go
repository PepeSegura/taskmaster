package controller

import (
	"fmt"
	"os/exec"
	"strings"

	"taskmaster/srcs/execution"
	_ "taskmaster/srcs/input"
	"taskmaster/srcs/parser"
	_ "taskmaster/srcs/signals"
	// _ "github.com/chzyer/readline"
)

type Execution struct {
	Programs map[string][]execution.Programs
}

var CMDs Execution

func Init(config parser.ConfigFile) {
	CMDs = Execution{
		Programs: make(map[string][]execution.Programs),
	}

	for name, program := range config.Programs {
		fmt.Println("Name: ", name)
		for index := range program.Numprocs {
			fmt.Printf("[%s] Executing %d/%d\n", strings.ToUpper(name), index+1, program.Numprocs)
			CMDs.add(name, program)
		}
	}
}

func Controller(config parser.ConfigFile) {

	fmt.Println("program map contents")

	for _, program := range CMDs.Programs {
		executeGroup(program)
	}
}

func executeGroup(program []execution.Programs) {
	// Create a channel to receive completion signals
	done := make(chan int, 5) // Buffered to handle up to 5 commands
	var cmds []*exec.Cmd

	for _, cmd_conf := range program {
		fmt.Println("Configuring and instance of " + cmd_conf.Name)
		execution.Cmd(cmd_conf, done)
		cmds = append(cmds, &cmd_conf.CmdInstance)
	}

	fmt.Println("Starting monitoring for process group " + "sleep")
	for i := 0; i < len(cmds); i++ {
		pid := <-done
		if pid == -1 {
			fmt.Println("A command failed to start or was nil.")
		} else {
			fmt.Printf("Command with PID %d has completed.\n", pid)
		}
	}
}

func (e *Execution) add(name string, program parser.Program) {
	newCmdInstance := *execution.SetCmdInfo(program)

	newProgram := execution.Programs{
		Name:         name,
		CmdInstance:  newCmdInstance,
		Pid:          newCmdInstance.Process.Pid,
		DateLaunched: "25/12/2024",
		DateFinish:   "26/12/2024",
		StopSignal:   "SIGTERM",
	}

	e.Programs[name] = append(e.Programs[name], newProgram)
}
