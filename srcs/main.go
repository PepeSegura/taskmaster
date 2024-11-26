package main

import (
	"fmt"
	"reflect"
	"sync/atomic"
	"taskmaster/srcs/controller"
	_ "taskmaster/srcs/input"
	"taskmaster/srcs/parser"
	"taskmaster/srcs/signals"
	"time"
	// _ "github.com/chzyer/readline"
)

func removeGroup() {
	/*
		Receives program group and kills all the procs from it :(
	*/
	fmt.Println("Killing the group!")
}

func addGroup() {
	/*
		Creates program group and executes all the procs :)
	*/
	fmt.Println("Adding a new group!")
}


func diffConfigs(oldConfig, newConfig parser.ConfigFile) {

	oldPrograms := oldConfig.Programs
	newPrograms := newConfig.Programs

	for programName := range oldPrograms {
		if _, exists := newPrograms[programName]; !exists {
			fmt.Printf("Program '%s' removed.\n", programName)
			removeGroup()
		}
	}
	for programName := range newPrograms {
		if _, exists := oldPrograms[programName]; !exists {
			fmt.Printf("Program '%s' added.\n", programName)
			addGroup()
		}
	}

	for programName, oldProgram := range oldPrograms {
		if newProgram, exists := newPrograms[programName]; exists {
			oldVal := reflect.ValueOf(oldProgram)
			newVal := reflect.ValueOf(newProgram)

			for i := 0; i < oldVal.NumField(); i++ {
				fieldName := oldVal.Type().Field(i).Name
				oldField := oldVal.Field(i).Interface()
				newField := newVal.Field(i).Interface()

				if !reflect.DeepEqual(oldField, newField) {
					fmt.Printf("Program '%s', field '%s' changed: '%v' -> '%v'\n",
						programName, fieldName, oldField, newField)
				}
			}
		}
	}
}

func main() {
	config := parser.Init("configs/basic.yml")

	controller.Init(config)
	controller.Controller(config)

	signals.Init()

	for signals.FinishProgram == 0 {
		if atomic.LoadInt32(&signals.ReloadProgram) == 1 {
			newConfig := parser.Init("configs/basic.yml")
			diffConfigs(config, newConfig)
			config = newConfig
			atomic.StoreInt32(&signals.ReloadProgram, 0)
		}
		fmt.Println("Doing things...")
		time.Sleep(time.Second / 2)
	}
	fmt.Println("Closing program...")
}
