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

	var wg sync.WaitGroup
	numprocs := len(group)
	wg.Add(numprocs) // stea sem a nmprocs
	logging.Info("Killing group: " + programName)
	for _, cmd_conf := range group {
		if cmd_conf.Status == execution.STOPPED || cmd_conf.Status == execution.FINISHED {
			wg.Done()
			continue
		} else if cmd_conf.CmdInstance.Process != nil {
			go sendSignal(&cmd_conf, &wg)
		} else {
			cmd_conf.Status = execution.FINISHED
			logging.Warning("Process [" + programName + "] was already finished")
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
		logging.Info(fmt.Sprintf("Executing an instance of %s with pid %d", (program)[i].Name, (program)[i].CmdInstance.Process.Pid))
		ctr += 1
	}

	//logging.Info(fmt.Sprintf("Starting monitoring for process group " + (program)[0].Name))
	for i := 0; i < ctr; i++ {
		pid := <-done
		if pid == -1 {
			logging.Error("A command failed to start or was nil") // revisar mas tarde
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
		logging.Error(fmt.Sprintf("Group %s doesnt exist", name))
		fmt.Printf("Group %s doesnt exist\n", name)
		return
	}

	for _, cmd_conf := range group {
		if cmd_conf.CmdInstance.Process != nil && cmd_conf.Status != execution.FINISHED {
			logging.Error(fmt.Sprintf("Some instance of group %s is already running, stop them first", name))
			fmt.Printf("Some instance of group %s is already running, stop them first\n", name)
			return
		}
	}

	go ExecuteGroup(group, false, true)
}

// falta logica de esperar stoptime, sino mandar SIGKILL
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

	newProgram := execution.Programs{
		Name:         name,
		CmdInstance:  newCmdInstance,
		Exitcodes:    program.Exitcodes,
		StopSignal:   program.Stopsignal,
		Umask:        program.Umask,
		Status:       execution.STOPPED,
		StartRetries: program.Startretries,
		RetryCtr:     0,
		StopTime:     program.Stoptime,
		StartTime:    program.Starttime,
		Autorestart:  program.Autorestart,
	}
	fmt.Println("DateLaunched: ", time.Now().Unix())

	e.Programs[name] = append(e.Programs[name], newProgram)
}

func sendSignal(cmd_conf *execution.Programs, wg *sync.WaitGroup) {
	defer wg.Done()

	signal_name := cmd_conf.StopSignal
	pid := cmd_conf.CmdInstance.Process.Pid

	signal_num, exists := parser.SignalTypes[signal_name]
	if !exists {
		logging.Error(fmt.Sprintf("invalid signal: %s", signal_name))
		syscall.Kill(pid, syscall.SIGKILL)
	}

	sentTime := time.Now().Unix()
	err := syscall.Kill(pid, signal_num)
	if err != nil {
		logging.Error(fmt.Sprintf("failed to send signal: %v", err))
		return
	}

	for time.Now().Unix()-sentTime <= int64(cmd_conf.StopTime) && cmd_conf.Status == execution.STARTED {
		time.Sleep(time.Second / 4)
	}

	if cmd_conf.Status == execution.STARTED {
		syscall.Kill(pid, syscall.SIGKILL)
	}
}
