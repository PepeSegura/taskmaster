package execution

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"taskmaster/srcs/logging"
	"taskmaster/srcs/parser"
)

const (
	STARTED  uint8 = 1
	STOPPED  uint8 = 2
	FINISHED uint8 = 3
)

type Programs struct {
	Name        string
	CmdInstance exec.Cmd
	// autorestart
	// exitcodes
	// startTime
	// startRetries
	// stopTime
	DateLaunched string
	DateFinish   string
	StopSignal   string
	Umask        int
	Status       uint8
}

func (cmd_conf *Programs) ExecCmd(done chan int) {
	// Store old umask and apply new one
	oldUmask := syscall.Umask(cmd_conf.Umask)

	// Start the command
	err := cmd_conf.CmdInstance.Start()
	if err != nil {
		logging.Error(fmt.Sprintf("Error: %v", err))
	}
	cmd_conf.Status = STARTED

	// Restore umask
	syscall.Umask(oldUmask)

	go monitorCmd(&cmd_conf.CmdInstance, done)
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
		cmd.Stdout = OpenFile(program.Stdout, true)
	} else {
		cmd.Stdout = os.Stdout
	}
	if program.Stderr != "" {
		cmd.Stderr = OpenFile(program.Stderr, true)
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

func monitorCmd(cmd *exec.Cmd, done chan int) {
	// Wait for the command to complete
	if cmd == nil {
		done <- (-1)
	}
	err := cmd.Wait()
	if err != nil {
		logging.Warning(fmt.Sprintf("Command: [%s] PID: [%d] finished with error: %v", cmd.Path, cmd.Process.Pid, err))
	} else {
		logging.Info(fmt.Sprintf("Command: [%s] PID: [%d] finished successfully!", cmd.Path, cmd.Process.Pid))
	}
	// Send the PID to the channel
	done <- cmd.Process.Pid
}

func (cmd_conf *Programs) PrintStatus() {
	var statusstr string
	switch cmd_conf.Status {
	case 0:
		statusstr = "STOPPED"
	case 1:
		statusstr = "RUNNING"
	case 2:
		statusstr = "FINISHED"
	}
	if cmd_conf.CmdInstance.Process != nil {
		fmt.Printf("%s PID %d\n", statusstr, cmd_conf.CmdInstance.Process.Pid)
	} else {
		fmt.Printf("%s\n", statusstr)
	}
}
