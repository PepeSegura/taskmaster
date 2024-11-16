package parser

import (
	"fmt"
	"log"
	"os"

	"encoding/json"
	"gopkg.in/yaml.v3"
	// "github.com/chzyer/readline"
)

type program struct {
	Cmd          string            `yaml:"cmd"`
	Numprocs     int               `yaml:"numprocs"`
	Autostart    bool              `yaml:"autostart"`
	Autorestart  string            `yaml:"autorestart"`
	Exitcodes    []int             `yaml:"exitcodes"`
	Starttime    int               `yaml:"starttime"`
	Startretries int               `yaml:"startretries"`
	Stopsignal   string            `yaml:"stopsignal"`
	Stoptime     string            `yaml:"stoptime"`
	Stdout       string            `yaml:"stdout"`
	Stderr       string            `yaml:"stderr"`
	Env          map[string]string `yaml:"env"`
	Workingdir   string            `yaml:"workingdir"`
	Umask        string            `yaml:"umask"`
}

type configFile struct {
	Programs map[string]program `yaml:"programs"`
}

func ReadFile(filename string) (data []byte) {
	data, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("failed to open file: %v", err)
	}
	return
}

func PrintConfigFile(config configFile) {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		fmt.Println("Error marshalling config:", err)
		return
	}
	fmt.Println(string(data))
}

func Parser(filename string) (config configFile) {
	data := ReadFile(filename)
	err := yaml.Unmarshal([]byte(data), &config)
	if err != nil {
		log.Fatalf("cannot unmarshal data: %v", err)
	}
	return
}
