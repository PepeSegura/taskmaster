package controller

import (
	"fmt"
	"os/exec"
	"strings"
	"syscall"

	"taskmaster/srcs/execution"
	_ "taskmaster/srcs/input"
	"taskmaster/srcs/parser"
	// _ "github.com/chzyer/readline"
)

type Execution struct {
	Programs map[string][]execution.Programs
}

var CMDs Execution

func AddGroup(name string, program parser.Program) {
	for index := range program.Numprocs {
		fmt.Printf("[%s] Executing %d/%d\n", strings.ToUpper(name), index+1, program.Numprocs)
		CMDs.add(name, program)
	}
	go ExecuteGroup(CMDs.Programs[name], true, program.Autostart)
	// for _, program := range CMDs.Programs {
	// }
}

func Init(config parser.ConfigFile) {
	CMDs = Execution{
		Programs: make(map[string][]execution.Programs),
	}

	for name, program := range config.Programs {
		fmt.Println("Name: ", name)
		AddGroup(name, program)
	}

}

func KillGroup(programName string) {
	group, exists := CMDs.Programs[programName]
	if !exists {
		fmt.Println("NO EXISTE :(")
		return
	}

	fmt.Println("Killing group: " + programName)
	for _, cmd_conf := range group {
		if cmd_conf.CmdInstance.Process != nil {
			sendSignal(cmd_conf.StopSignal, cmd_conf.CmdInstance.Process.Pid)
		} else {
			cmd_conf.Status = execution.FINISHED
			fmt.Println("ESTA MUELTO")
		}
	}
	delete(CMDs.Programs, programName)
}

func ExecuteGroup(program []execution.Programs, autocall, autostart bool) {
	if autocall && !autostart {
		return
	}
	done := make(chan int, len(program))
	var cmds []*exec.Cmd

	for i := range program {
		(program)[i].ExecCmd(done)
		cmds = append(cmds, &(program)[i].CmdInstance)
		fmt.Printf("Configuring an instance of %s with pid %d\n", (program)[i].Name, (program)[i].CmdInstance.Process.Pid)
	}

	fmt.Println("Starting monitoring for process group " + (program)[0].Name)
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
	newCmdInstance := *execution.CreateCmdInstance(program)

	//aqui faltan cosas creo
	// pepe: si que faltan si :(
	newProgram := execution.Programs{
		Name:         name,
		CmdInstance:  newCmdInstance,
		DateLaunched: "25/12/2024",
		DateFinish:   "26/12/2024",
		StopSignal:   program.Stopsignal,
		Umask:        program.Umask,
	}

	e.Programs[name] = append(e.Programs[name], newProgram)
}

func sendSignal(signal_name string, pid int) error {
	var sig syscall.Signal

	switch signal_name {
	case "SIGTERM":
		sig = syscall.SIGTERM
	case "SIGKILL":
		sig = syscall.SIGKILL
	case "SIGINT":
		sig = syscall.SIGINT
	case "SIGSTOP":
		sig = syscall.SIGSTOP
	case "SIGUSR1":
		sig = syscall.SIGUSR1
	case "SIGUSR2":
		sig = syscall.SIGUSR2
	default:
		return fmt.Errorf("invalid signal: %s", signal_name)
	}

	err := syscall.Kill(pid, sig)
	if err != nil {
		return fmt.Errorf("failed to send signal: %v", err)
	}
	return nil
}
