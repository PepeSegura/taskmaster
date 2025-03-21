package controller

import (
	"fmt"
	"strings"
	"sync"
	"syscall"
	"time"

	"taskmaster/srcs/execution"
	"taskmaster/srcs/logging"
	"taskmaster/srcs/parser"
)

type Execution struct {
	Programs map[string][]execution.Programs
}

var CMDs Execution

var Config parser.ConfigFile

func AddGroup(name string, program parser.Program) {
	logging.Info(fmt.Sprintf("Adding group: %s", name))
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

	var wg sync.WaitGroup
	numprocs := len(group)
	wg.Add(numprocs)
	logging.Info("Killing and removing group: " + programName)
	for i := range group {
		if group[i].Status == execution.STOPPED || group[i].Status == execution.FINISHED {
			logging.Info("Process [" + programName + "] was already finished/stopped")
			wg.Done()
			continue
		} else if group[i].CmdInstance.Process != nil && (group[i].Status != execution.FINISHED && group[i].Status != execution.FATAL) {
			go sendSignal(&group[i], &wg)
		} else {
			logging.Info("Process [" + programName + "] was already finished")
			wg.Done()
		}
	}
	wg.Wait()
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
		ctr += 1
		if program[i].CmdInstance.Process == nil {
			continue
		}

		logging.Info(fmt.Sprintf("Executing an instance of [%s] with pid %d", (program)[i].Name, (program)[i].CmdInstance.Process.Pid))
	}

	for i := 0; i < ctr; i++ {
		pid := <-done
		for j := 0; j < ctr; j++ {
			if (program)[j].CmdInstance.Process != nil && (program)[j].CmdInstance.Process.Pid == pid {
				i += (program)[j].CheckEnd(done)
			}
		}

		if pid == -1 {
			logging.Error(fmt.Sprintf("An instance of [%s] presented a fatal error (failed to start)", program[0].Name))
		}
	}
	close(done)
}

func programStatus(program []execution.Programs) [][]string {
	rows := [][]string{}
	for _, instance := range program {
		rows = append(rows, instance.PrintStatus())
	}
	return rows
}

func printTable(header []string, rows [][]string, groupLen int) {
	hline := strings.Repeat("─", 39)
	topline_left := strings.Repeat("─", groupLen+2)
	topline_right := strings.Repeat("─", 36-groupLen)

	fmt.Printf("├%s┴%s┐\n", topline_left, topline_right)
	fmt.Printf("│ %-20s │ %-14s │\n", header[0], header[1])
	fmt.Printf("├%s┤\n", hline)
	for _, row := range rows {
		fmt.Printf("│ %-20s │ %-14s │\n", row[0], row[1])
	}
	fmt.Printf("└%s┘\n", hline)
}

func Status() {
	for key, program := range CMDs.Programs {
		fmt.Printf("\n┌ %s ┐\n", key)
		headers := []string{"Status", "PID"}
		rows := programStatus(program)
		printTable(headers, rows, len(key))
	}
}

func Try2StartGroup(name string) {
	logging.Info("Trying to start group: " + name)
	group, exists := CMDs.Programs[name]

	if !exists {
		logging.Error(fmt.Sprintf("Group [%s] doesnt exist", name))
		fmt.Printf("Group [%s] doesnt exist\n", name)
		return
	}

	for _, cmd_conf := range group {
		if cmd_conf.CmdInstance.Process != nil && cmd_conf.Status != execution.FINISHED {
			logging.Error(fmt.Sprintf("Some instance of group %s is already running, stop them first", name))
			fmt.Printf("Some instance of group [%s] is already running, stop them first\n", name)
			return
		}
	}

	go ExecuteGroup(group, false, true)
}

func Try2StopGroup(name string) {
	_, exists := CMDs.Programs[name]
	if !exists {
		logging.Error(fmt.Sprintf("Group [%s] doesnt exist", name))
		fmt.Printf("Group [%s] doesnt exist\n", name)
		return
	}

	logging.Info("Trying to stop group: " + name)
	KillGroup(name, true)
}

func (e *Execution) add(name string, program parser.Program) {
	newCmdInstance := *execution.CreateCmdInstance(program)

	var autorest uint8
	switch program.Autorestart {
	case "always":
		autorest = execution.ALWAYS
	case "never":
		autorest = execution.NEVER
	case "unexpected":
		autorest = execution.UNEXPECTED
	}

	newProgram := execution.Programs{
		Name:             name,
		CmdInstance:      newCmdInstance,
		Exitcodes:        program.Exitcodes,
		StopSignal:       program.Stopsignal,
		Umask:            program.Umask,
		Status:           execution.STOPPED,
		StartRetries:     program.Startretries,
		RetryCtr:         0,
		StopTime:         program.Stoptime,
		StartTime:        program.Starttime,
		Autorestart:      program.Autorestart,
		CmdStr:           program.Cmd,
		EnvMap:           program.Env,
		StdoutStr:        program.Stdout,
		StderrStr:        program.Stderr,
		RestartCondition: autorest,
		ManuallyStopped:  false,
	}
	e.Programs[name] = append(e.Programs[name], newProgram)
}

func sendSignal(cmd_conf *execution.Programs, wg *sync.WaitGroup) {
	cmd_conf.ManuallyStopped = true //set to diferentiate the signals sent from taskmaster
	defer wg.Done()

	signal_name := cmd_conf.StopSignal
	signal_num, exists := parser.SignalTypes[signal_name]

	var pid int

	if !exists { // unknown signal
		logging.Error(fmt.Sprintf("invalid signal: %s", signal_name))
		if cmd_conf.CmdInstance.Process != nil {
			pid = cmd_conf.CmdInstance.Process.Pid
			syscall.Kill(pid, syscall.SIGKILL)
		} else {
			return
		}
	}

	var err error
	sentTime := time.Now().Unix()
	if cmd_conf.CmdInstance.Process != nil {
		pid = cmd_conf.CmdInstance.Process.Pid
		err = syscall.Kill(pid, signal_num)
	} else {
		return
	}
	if err != nil {
		logging.Error(fmt.Sprintf("failed to send signal: %v", err))
		return
	}

	for time.Now().Unix()-sentTime <= int64(cmd_conf.StopTime) && cmd_conf.Status == execution.STARTED {
		time.Sleep(time.Millisecond * 10)
	}

	if cmd_conf.Status == execution.STARTED {
		if cmd_conf.CmdInstance.Process != nil {
			pid = cmd_conf.CmdInstance.Process.Pid
			syscall.Kill(pid, syscall.SIGKILL)
		}
	}
}
