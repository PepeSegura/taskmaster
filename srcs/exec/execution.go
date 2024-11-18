package exec

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

func (e *Execution) add(name string, program parser.Program) {
	fmt.Println("Adding: ", program)
	newCmdInstance := *Cmd(program.Cmd)

	newProgram := Programs{
		Name:         name,
		CmdInstance:  newCmdInstance,
		Pid:          newCmdInstance.Process.Pid,
		DateLaunched: "25/12/2024",
		DateFinish:   "26/12/2024",
		StopSignal:   "SIGTERM",
	}

	// Append to the slice; initialize slice if nil
	e.Programs[name] = append(e.Programs[name], newProgram)
}

func Init(config parser.ConfigFile) {
	CMDs := Execution{
		Programs: make(map[string][]Programs),
	}

	for name, program := range config.Programs {
		fmt.Println("Name: ", name)
		fmt.Println("Cmd: ", program.Cmd)
		if CMDs.Programs[name] != nil {
			fmt.Println("Ya existe en la execucion el comando " + name)
		} else {
			fmt.Println("Aun no existe en la execucion el comando " + name)
			CMDs.add(name, program)
			// fmt.Println("Despues de añadir el comando "+name+"\n", CMDs.Programs[name][0])
		}
	}
}
