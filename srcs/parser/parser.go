package parser

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"syscall"
	"taskmaster/srcs/logging"

	"gopkg.in/yaml.v3"
)

type Program struct {
	Cmd          string            `yaml:"cmd"`
	Numprocs     int               `yaml:"numprocs"`
	Autostart    bool              `yaml:"autostart"`
	Autorestart  string            `yaml:"autorestart"`
	Exitcodes    []int             `yaml:"exitcodes"`
	Starttime    int               `yaml:"starttime"`
	Startretries int               `yaml:"startretries"`
	Stopsignal   string            `yaml:"stopsignal"`
	Stoptime     int               `yaml:"stoptime"`
	Stdout       string            `yaml:"stdout"`
	Stderr       string            `yaml:"stderr"`
	Env          map[string]string `yaml:"env"`
	Workingdir   string            `yaml:"workingdir"`
	Umask        int               `yaml:"umask"`
}

type ConfigFile struct {
	Programs map[string]Program `yaml:"programs"`
}

var SignalTypes = map[string]syscall.Signal{
	"SIHUP":     1,
	"SIGINT":    2,
	"SIGQUIT":   3,
	"SIGILL":    4,
	"SIGTRAP":   5,
	"SIGABRT":   6,
	"SIGBUS":    7,
	"SIGFPE":    8,
	"SIGKILL":   9,
	"SIGUSR1":   10,
	"SIGSSEGV":  11,
	"SIGUSR2":   12,
	"SIGPIPE":   13,
	"SIGALRM":   14,
	"SIGTERM":   15,
	"SIGCHLD":   17,
	"SIGCONT":   18,
	"SIGSTOP":   19,
	"SIGTSTP":   20,
	"SIGTTIN":   21,
	"SIGTTOU":   22,
	"SIGURG":    23,
	"SIGXCPU":   24,
	"SIGXFSZ":   25,
	"SIGVTALRM": 26,
	"SIGPROF":   27,
	"SIGWINCH":  28,
	"SIGIO":     29,
	"SIGPWR":    30,
	"SIGSYS":    31,
}

func (config *ConfigFile) Print() {
	data, err := yaml.Marshal(config)
	if err != nil {
		logging.Error(fmt.Sprintf("Error marshalling config to YAML: %v", err))
		return
	}
	fmt.Println(string(data))
}

func validateSignal(name string) (string, error) {
	name = strings.ToUpper(name)
	_, exists := SignalTypes[name]
	if !exists {
		return "", errors.New("invalid signal name: [" + name + "]")
	}
	return name, nil
}

func (p *Program) validate() error {
	var err error

	if strings.TrimSpace(p.Cmd) == "" {
		return fmt.Errorf("Cmd is empty")
	}
	if p.Numprocs < 0 {
		return fmt.Errorf("Invalid NumProcs: [%d]", p.Numprocs)
	}
	if p.Autorestart != "always" && p.Autorestart != "never" && p.Autorestart != "unexpected" {
		return fmt.Errorf("Invalid option in AutoRestart [%s]", p.Autorestart)
	}
	if p.Starttime < 0 {
		return fmt.Errorf("Invalid StartTime: [%d]", p.Starttime)
	}
	if p.Startretries < 0 {
		return fmt.Errorf("Invalid StartRetries: [%d]", p.Startretries)
	}
	if p.Stoptime < 0 {
		return fmt.Errorf("Invalid StopTime: [%d]", p.Stoptime)
	}
	p.Stopsignal, err = validateSignal(p.Stopsignal)
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	return nil
}

func (c *ConfigFile) validate() {
	for name, program := range c.Programs {
		err := program.validate()
		if err != nil {
			fmt.Printf("Parsing error: [%s] -> %v", name, err)
			logging.Fatal(fmt.Sprintf("Parsing error: [%s] -> %v\n", name, err))
		}
	}
}

func readFile(filename string) (data []byte) {
	data, err := os.ReadFile(filename)
	if err != nil {
		logging.Fatal(fmt.Sprint("Error: %v", err))
	}
	return
}

func checkArgs(default_file string) (string, error) {
	if len(os.Args) == 2 {
		return os.Args[1], nil
	}
	if len(os.Args) > 2 {
		return default_file, errors.New("Invalid number of arguments.")
	}
	fmt.Println("No arguments passed, using [" + default_file + "]")
	return default_file, nil
}

func (c *ConfigFile) load(filename string) {
	var err error

	data := readFile(filename)

	err = yaml.Unmarshal([]byte(data), &c)
	if err != nil {
		logging.Fatal(fmt.Sprintf("cannot unmarshal data: %v", err))
	}
}

func Init(filename string) (config ConfigFile) {
	var err error

	filename, err = checkArgs(filename)
	if err != nil {
		logging.Fatal(fmt.Sprintf("Error: %v", err))
	}

	config.load(filename)
	config.validate()
	config.Print()
	return
}
