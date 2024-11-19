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

type Execution struct {
	Programs map[string][]Programs
}

func Cmd(path string) *exec.Cmd {
	var args_command []string = strings.Split(path, " ")

	cmd := exec.Command(args_command[0], args_command[1:]...)

	cmd.Env = append(os.Environ(), "MY_VAR=some_value")

	// cmd.Dir = "/"

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Store old umask and apply new one
	newUmask := 022
	oldUmask := syscall.Umask(newUmask)

	// Start the command
	err := cmd.Start()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	// Restore umask
	syscall.Umask(oldUmask)
	// cmd.Wait()
	return cmd
}

func setEnv(cmd *exec.Cmd, newEnv map[string]string) {
	env := os.Environ()

    for key, value := range newEnv {
        env = append(env, fmt.Sprintf("%s=%s", key, value))
    }

    cmd.Env = env
}

func openFile(path string) (*os.File) {
	outputFile, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
    if err != nil {
        fmt.Printf("Error opening output file: %v\n", err)
        return os.Stderr
    }
	return outputFile
}

func setCmdInfo(program parser.Program) *exec.Cmd {
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

func (e *Execution) add(name string, program parser.Program) {
	newCmdInstance := *setCmdInfo(program)

	newProgram := Programs{
		Name:         name,
		CmdInstance:  newCmdInstance,
		Pid:          newCmdInstance.Process.Pid,
		DateLaunched: "25/12/2024",
		DateFinish:   "26/12/2024",
		StopSignal:   "SIGTERM",
	}

	e.Programs[name] = append(e.Programs[name], newProgram)
}

func Init(config parser.ConfigFile) {
	CMDs := Execution{
		Programs: make(map[string][]Programs),
	}

	for name, program := range config.Programs {
		fmt.Println("Name: ", name)
		for index := range program.Numprocs {
			fmt.Printf("[%s] Executing %d/%d\n", strings.ToUpper(name), index + 1, program.Numprocs)
			CMDs.add(name, program)
		}
	}
}
