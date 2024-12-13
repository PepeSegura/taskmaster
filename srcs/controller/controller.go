package controller

import (
	"fmt"
	"strings"
	"syscall"

	"taskmaster/srcs/execution"
	"taskmaster/srcs/logging"
	"taskmaster/srcs/parser"
	// _ "github.com/chzyer/readline"
)

type Execution struct {
	Programs map[string][]execution.Programs
}

var CMDs Execution

var Config parser.ConfigFile

func AddGroup(name string, program parser.Program) {
	logging.Info(fmt.Sprintf("Adding group: %s", strings.ToUpper(name)))
	for i := 0; i < program.Numprocs; i++ {
		CMDs.add(name, program)
	}
}

func Init(config parser.ConfigFile) {
	CMDs = Execution{
		Programs: make(map[string][]execution.Programs),
	}

	for name, program := range config.Programs {
		AddGroup(name, program)
		go ExecuteGroup(CMDs.Programs[name], true, program.Autostart)
	}

}

func KillGroup(programName string, reAdd bool) {
	group, exists := CMDs.Programs[programName]
	if !exists {
		return
	}

	logging.Info("Killing group: " + programName)
	for _, cmd_conf := range group {
		if cmd_conf.CmdInstance.Process != nil {
			sendSignal(cmd_conf.StopSignal, cmd_conf.CmdInstance.Process.Pid)
		} else {
			cmd_conf.Status = execution.FINISHED
			logging.Warning("Process [" + programName + "] was already finished")
		}
	}
	delete(CMDs.Programs, programName)
	if reAdd {
		AddGroup(programName, Config.Programs[programName])
	}
}

func KillAll() {
	for key := range CMDs.Programs {
		KillGroup(key, false)
	}
}

func ExecuteGroup(program []execution.Programs, autocall, autostart bool) {
	if autocall && !autostart {
		return
	}
	done := make(chan int, len(program))
	ctr := 0

	for i := range program {
		(program)[i].ExecCmd(done)
		logging.Info(fmt.Sprintf("Executing an instance of %s with pid %d", (program)[i].Name, (program)[i].CmdInstance.Process.Pid))
		ctr += 1
	}

	logging.Info(fmt.Sprintf("Starting monitoring for process group " + (program)[0].Name))
	for i := 0; i < ctr; i++ {
		pid := <-done
		if pid == -1 {
			logging.Error("A command failed to start or was nil") // revisar mas tarde
		} else {
			logging.Info(fmt.Sprintf("Command with PID %d has completed.", pid))
		}
	}
}

func programStatus(program []execution.Programs) {
	for _, instance := range program {
		instance.PrintStatus()
	}
}

func Status() {
	for key, program := range CMDs.Programs {
		fmt.Printf("\nStatus of group \"%s\":\n\n", key)
		programStatus(program)
	}
}

func Try2StartGroup(name string) {
	logging.Info("Trying to start group: " + name)
	group, exists := CMDs.Programs[name]

	if !exists {
		logging.Error(fmt.Sprintf("Group %s doesnt exist", name))
		fmt.Printf("Group %s doesnt exist\n", name)
		return
	}

	for _, cmd_conf := range group {
		if cmd_conf.CmdInstance.Process != nil {
			logging.Error(fmt.Sprintf("Some instance of group %s is already running, stop them first", name))
			fmt.Printf("Some instance of group %s is already running, stop them first\n", name)
			return
		}
	}

	go ExecuteGroup(group, false, true)
}

func Try2StopGroup(name string) {
	_, exists := CMDs.Programs[name]
	if !exists {
		logging.Error(fmt.Sprintf("Group %s doesnt exist", name))
		fmt.Printf("Group %s doesnt exist\n", name)
		return
	}

	logging.Info("Trying to stop group: " + name)
	KillGroup(name, true)
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

func sendSignal(signal_name string, pid int) {
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
		logging.Error(fmt.Sprintf("invalid signal: %s", signal_name))
	}

	err := syscall.Kill(pid, sig)
	if err != nil {
		logging.Error(fmt.Sprintf("failed to send signal: %v", err))
	}
}
