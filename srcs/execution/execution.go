package execution

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
	"unicode"

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
	ManuallyStopped  bool
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
		done <- -1
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
	var args_command []string = tokenize(program.Cmd)

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

	return cmd
}

func monitorCmd(cmd_conf *Programs, done chan int) {
	err := cmd_conf.CmdInstance.Wait()

	if err != nil {
		logging.Warning(fmt.Sprintf("Command: [%s] PID: [%d] finished with error: %v", cmd_conf.CmdInstance.Path, cmd_conf.CmdInstance.Process.Pid, err))
	}
	if cmd_conf.ManuallyStopped {
		cmd_conf.Status = STOPPED
		logging.Info(fmt.Sprintf("Command: [%s] PID: [%d] was manually stopped, signal not treated as error", cmd_conf.CmdInstance.Path, cmd_conf.CmdInstance.Process.Pid))
		done <- cmd_conf.CmdInstance.Process.Pid
		return
	}
	exitCode := cmd_conf.CmdInstance.ProcessState.ExitCode()

	for _, validCode := range cmd_conf.Exitcodes {
		if exitCode == validCode {
			if time.Now().Unix()-cmd_conf.DateLaunched >= int64(cmd_conf.StartTime) {
				logging.Info(fmt.Sprintf("Command: [%s] PID: [%d] finished successfully!", cmd_conf.CmdInstance.Path, cmd_conf.CmdInstance.Process.Pid))
				cmd_conf.Status = FINISHED
			} else {
				logging.Info(fmt.Sprintf("Command: [%s] PID: [%d] finished too early!", cmd_conf.CmdInstance.Path, cmd_conf.CmdInstance.Process.Pid))
				cmd_conf.Status = FAILED
			}
			done <- cmd_conf.CmdInstance.Process.Pid
			return
		}
	}
	logging.Info(fmt.Sprintf("Command: [%s] PID: [%d] finished with invalid exitcode", cmd_conf.CmdInstance.Path, cmd_conf.CmdInstance.Process.Pid))
	cmd_conf.Status = FAILED
	done <- cmd_conf.CmdInstance.Process.Pid
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
		if cmd_conf.RetryCtr <= cmd_conf.StartRetries {
			cmd_conf.Retry(done)
			return -1
		} else {
			cmd_conf.Status = FATAL
			logging.Info(fmt.Sprintf("[%s]: exceeded max retries; FATAL", cmd_conf.Name))
		}
	} else if cmd_conf.Status == FINISHED && cmd_conf.RestartCondition == ALWAYS { // solo always con exit code valido
		cmd_conf.RetryCtr++
		if cmd_conf.RetryCtr <= cmd_conf.StartRetries {
			cmd_conf.Retry(done)
			return -1
		} else {
			logging.Info(fmt.Sprintf("[%s]: exceeded max retries", cmd_conf.Name))
		}
	}
	return 0
}

func (cmd_conf *Programs) Retry(done chan int) {

	var args_command []string = tokenize(cmd_conf.CmdStr)

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

	if cmd_conf.CmdInstance.Process != nil {
		logging.Info(fmt.Sprintf("Retry #%d: Executing an instance of [%s] with pid %d", cmd_conf.RetryCtr, cmd_conf.Name, cmd_conf.CmdInstance.Process.Pid))
	} else {
		logging.Info(fmt.Sprintf("Retry #%d: Executing an instance of [%s]", cmd_conf.RetryCtr, cmd_conf.Name))
	}
	go monitorCmd(cmd_conf, done)
}

func tokenize(input string) []string {
	var tokens []string
	var currentToken strings.Builder
	var inQuote rune

	for _, r := range input {
		switch {
		case inQuote == 0:
			if unicode.IsSpace(r) {
				if currentToken.Len() > 0 {
					tokens = append(tokens, currentToken.String())
					currentToken.Reset()
				}
			} else if r == '\'' || r == '"' {
				if currentToken.Len() > 0 {
					tokens = append(tokens, currentToken.String())
					currentToken.Reset()
				}
				inQuote = r
			} else {
				currentToken.WriteRune(r)
			}

		default:
			if r == inQuote {
				if currentToken.Len() > 0 {
					tokens = append(tokens, currentToken.String())
					currentToken.Reset()
				}
				inQuote = 0
			} else {
				currentToken.WriteRune(r)
			}
		}
	}

	if currentToken.Len() > 0 {
		tokens = append(tokens, currentToken.String())
	}

	return tokens
}
