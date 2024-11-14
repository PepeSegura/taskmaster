package main

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

func readFile(filename string) (data []byte) {
	data, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("failed to open file: %v", err)
	}
	return
}

func printConfigFile(config configFile) {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		fmt.Println("Error marshalling config:", err)
		return
	}
	fmt.Println(string(data))
}

func main() {
	var config configFile

	data := readFile("/workspaces/taskmaster/srcs/config.yml")

	err := yaml.Unmarshal([]byte(data), &config)
	if err != nil {
		log.Fatalf("cannot unmarshal data: %v", err)
	}
	printConfigFile(config)
}

// func main() {
//     rl, err := readline.New("readline> ")
//     if err != nil {
//         panic(err)
//     }
//     defer rl.Close()

//     for {
//         line, err := rl.Readline()
//         if err != nil { // io.EOF on Ctrl+D
//             break
// 		}
// 		fmt.Println(err)
//         fmt.Println("line:", line)
//     }
// }
