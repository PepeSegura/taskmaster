package execution

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"taskmaster/srcs/logging"
	"taskmaster/srcs/parser"
)

const (
	STARTED  uint8 = 1
	STOPPED  uint8 = 2
	FINISHED uint8 = 3
	FAILED   uint8 = 4
	FATAL    uint8 = 5
)

const (
	NEVER      uint8 = 1
	ALWAYS     uint8 = 2
	UNEXPECTED uint8 = 3
)

type Programs struct {
	Name             string
	CmdInstance      exec.Cmd
	Autorestart      string
	Exitcodes        []int
	DateLaunched     int64
	StopSignal       string
	Umask            int
	Status           uint8
	StartRetries     int
	RetryCtr         int
	StopTime         int
	StartTime        int
	CmdStr           string
	EnvMap           map[string]string
	StderrStr        string
	StdoutStr        string
	RestartCondition uint8
}

func (cmd_conf *Programs) ExecCmd(done chan int) {
	// Store old umask and apply new one
	oldUmask := syscall.Umask(cmd_conf.Umask)

	// Start the command
	err := cmd_conf.CmdInstance.Start()
	cmd_conf.DateLaunched = time.Now().Unix()
	if err != nil {
		logging.Error(fmt.Sprintf("Error: %v", err))
		cmd_conf.Status = FATAL
		syscall.Umask(oldUmask)
		return
	} else {
		cmd_conf.Status = STARTED
	}

	// Restore umask
	syscall.Umask(oldUmask)

	go monitorCmd(cmd_conf, done)
}

func setEnv(cmd *exec.Cmd, newEnv map[string]string) {
	env := os.Environ()

	for key, value := range newEnv {
		env = append(env, fmt.Sprintf("%s=%s", key, value))
	}

	cmd.Env = env
}

func OpenFile(path string, truncate bool) *os.File {
	var outputFile *os.File
	var err error
	if truncate {
		outputFile, err = os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	} else {
		outputFile, err = os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	}
	if err != nil {
		logging.Error(fmt.Sprintf("Error opening output file: %v", err))
		return os.Stderr
	}
	return outputFile
}

func CreateCmdInstance(program parser.Program) *exec.Cmd {
	var args_command []string = strings.Split(program.Cmd, " ")

	cmd := exec.Command(args_command[0], args_command[1:]...)

	setEnv(cmd, program.Env)

	cmd.Dir = program.Workingdir

	if program.Stdout != "" {
		cmd.Stdout = OpenFile(program.Stdout, false)
	} else {
		cmd.Stdout = os.Stdout
	}
	if program.Stderr != "" {
		cmd.Stderr = OpenFile(program.Stderr, false)
	} else {
		cmd.Stderr = os.Stderr
	}

	/* 	defer func() {
		if stdout != os.Stderr {
			stdout.Close()
		}
		if stderr != os.Stderr {
			stderr.Close()
		}
	}()*/

	return cmd
}

func monitorCmd(cmd_conf *Programs, done chan int) {
	err := cmd_conf.CmdInstance.Wait()
	if err != nil {
		logging.Warning(fmt.Sprintf("Command: [%s] PID: [%d] finished with error: %v", cmd_conf.CmdInstance.Path, cmd_conf.CmdInstance.Process.Pid, err))
	} else {
		exitCode := cmd_conf.CmdInstance.ProcessState.ExitCode()

		for _, validCode := range cmd_conf.Exitcodes {
			if exitCode == validCode {
				if time.Now().Unix()-cmd_conf.DateLaunched >= int64(cmd_conf.StartTime) {
					logging.Info(fmt.Sprintf("Command: [%s] PID: [%d] finished successfully!", cmd_conf.CmdInstance.Path, cmd_conf.CmdInstance.Process.Pid))
					fmt.Printf("Command: [%s] PID: [%d] finished successfully!\n", cmd_conf.CmdInstance.Path, cmd_conf.CmdInstance.Process.Pid)
					cmd_conf.Status = FINISHED
				} else {
					logging.Info(fmt.Sprintf("Command: [%s] PID: [%d] finished too early!", cmd_conf.CmdInstance.Path, cmd_conf.CmdInstance.Process.Pid))
					fmt.Printf("Command: [%s] PID: [%d] failed !!! :(\n", cmd_conf.CmdInstance.Path, cmd_conf.CmdInstance.Process.Pid)
					cmd_conf.Status = FAILED
				}
				done <- cmd_conf.CmdInstance.Process.Pid
				return
			}
		}
		time.Sleep(time.Second / 4)
	}

	cmd_conf.Status = FAILED
	done <- (-1)
}

func (cmd_conf *Programs) PrintStatus() []string {
	var statusstr string
	switch cmd_conf.Status {
	case STOPPED:
		statusstr = "STOPPED"
	case STARTED:
		statusstr = "RUNNING"
	case FINISHED:
		statusstr = "FINISHED"
	case FAILED:
		statusstr = "FAILED"
	case FATAL:
		statusstr = "FATAL"
	}

	if cmd_conf.CmdInstance.Process != nil {
		return []string{statusstr, fmt.Sprintf("%d", cmd_conf.CmdInstance.Process.Pid)}
	}
	return []string{statusstr, "-"}
}

func (cmd_conf *Programs) CheckEnd(done chan int) int {
	if cmd_conf.RestartCondition == NEVER {
		return 0
	}
	if cmd_conf.Status == FAILED { // tanto always como unexpected en caso de mal exit code
		cmd_conf.RetryCtr++
		if cmd_conf.RetryCtr >= cmd_conf.StartRetries {
			cmd_conf.Status = FATAL
		} else {
			cmd_conf.Retry(done)
			return -1
		}
	} else if cmd_conf.Status == FINISHED && cmd_conf.RestartCondition == ALWAYS { // solo always con exit code valido
		cmd_conf.RetryCtr++
		if cmd_conf.RetryCtr < cmd_conf.StartRetries {
			cmd_conf.Retry(done)
			return -1
		}
	}
	return 0
}

func (cmd_conf *Programs) Retry(done chan int) {
	var args_command []string = strings.Split(cmd_conf.CmdStr, " ")

	cmd := exec.Command(args_command[0], args_command[1:]...)

	dir := cmd_conf.CmdInstance.Dir

	setEnv(cmd, cmd_conf.EnvMap)

	cmd.Dir = dir

	if cmd_conf.StdoutStr != "" {
		cmd.Stdout = OpenFile(cmd_conf.StdoutStr, false)
	} else {
		cmd.Stdout = os.Stdout
	}
	if cmd_conf.StderrStr != "" {
		cmd.Stderr = OpenFile(cmd_conf.StderrStr, false)
	} else {
		cmd.Stderr = os.Stderr
	}

	cmd_conf.CmdInstance = *cmd

	// Store old umask and apply new one
	oldUmask := syscall.Umask(cmd_conf.Umask)

	// Start the command
	err := cmd_conf.CmdInstance.Start()
	cmd_conf.DateLaunched = time.Now().Unix()
	if err != nil {
		logging.Error(fmt.Sprintf("Error: %v", err))
	}
	cmd_conf.Status = STARTED

	// Restore umask
	syscall.Umask(oldUmask)

	go monitorCmd(cmd_conf, done)
}
