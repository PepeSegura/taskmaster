package execution

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"taskmaster/srcs/parser"
)

type Programs struct {
	Name         string
	CmdInstance  exec.Cmd
	Pid          int
	DateLaunched string
	DateFinish   string
	StopSignal   string
	Status       bool
}

func Cmd(cmd_conf Programs, done chan int) *exec.Cmd {
	args_command := []string{cmd_conf.Name, "4"}

	cmd_conf.CmdInstance = *exec.Command(args_command[0], args_command[1:]...)

	cmd_conf.CmdInstance.Env = append(os.Environ(), cmd_conf.CmdInstance.Env...)

	// cmd.Dir = "/"

	cmd_conf.CmdInstance.Stdout = os.Stdout
	cmd_conf.CmdInstance.Stderr = os.Stderr

	// Store old umask and apply new one
	newUmask := 022
	oldUmask := syscall.Umask(newUmask)

	// Start the command
	err := cmd_conf.CmdInstance.Start()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	// Restore umask
	syscall.Umask(oldUmask)
	// cmd.Wait()

	//monitor command

	go monitorCmd(&cmd_conf.CmdInstance, done)
	return &cmd_conf.CmdInstance
}

func setEnv(cmd *exec.Cmd, newEnv map[string]string) {
	env := os.Environ()

	for key, value := range newEnv {
		env = append(env, fmt.Sprintf("%s=%s", key, value))
	}

	cmd.Env = env
}

func openFile(path string) *os.File {
	outputFile, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Printf("Error opening output file: %v\n", err)
		return os.Stderr
	}
	return outputFile
}

func SetCmdInfo(program parser.Program) *exec.Cmd {
	var args_command []string = strings.Split(program.Cmd, " ")

	cmd := exec.Command(args_command[0], args_command[1:]...)

	setEnv(cmd, program.Env)

	cmd.Dir = program.Workingdir

	stdout := openFile(program.Stdout)
	stderr := openFile(program.Stderr)

	cmd.Stdout = stdout
	cmd.Stderr = stderr

	defer func() {
		if stdout != os.Stderr {
			stdout.Close()
		}
		if stderr != os.Stderr {
			stderr.Close()
		}
	}()

	oldUmask := syscall.Umask(program.Umask)

	err := cmd.Start()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	syscall.Umask(oldUmask)
	return cmd
}

func monitorCmd(cmd *exec.Cmd, done chan int) {
	// Wait for the command to complete
	if cmd == nil {
		done <- (-1)
	}
	err := cmd.Wait()
	if err != nil {
		fmt.Printf("Command: [%s] PID: [%d] finished with error: %v\n", cmd.Path, cmd.Process.Pid, err)
	} else {
		fmt.Printf("Command: [%s] PID: [%d] finished successfully!\n", cmd.Path, cmd.Process.Pid)
	}
	// Send the PID to the channel
	done <- cmd.Process.Pid
}
