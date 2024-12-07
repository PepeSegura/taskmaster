package signals

import (
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"sync/atomic"
	"syscall"
	"taskmaster/srcs/controller"
	"taskmaster/srcs/parser"
)

var ReloadProgram int32 = 0
var FinishProgram int32 = 0
var signalChannel chan os.Signal = make(chan os.Signal, 1)

func signalHandler(sig os.Signal) {
	fmt.Printf("Signal (%d) received: %s\n", sig.(syscall.Signal), sig.String())

	if sig == syscall.SIGHUP {
		fmt.Println("Reloading config...")
		atomic.StoreInt32(&ReloadProgram, 1)
	} else if sig == syscall.SIGKILL || sig == syscall.SIGSTOP || sig == syscall.SIGINT {
		fmt.Println("Closing program...")
		atomic.StoreInt32(&FinishProgram, 1)
	} else {
		fmt.Println("Ignoring signal...")
	}
}

func setupSignalHandler() {
	fmt.Println("Setting signalHandler")
	arr := [4]int{1, 2, 9, 19}

	for _, sigNum := range arr {
		sig := syscall.Signal(sigNum)
		signal.Notify(signalChannel, sig)
	}
}

func Init() {
	setupSignalHandler()
	go func() {
		for {
			signal := <-signalChannel
			signalHandler(signal)
			if atomic.LoadInt32(&ReloadProgram) == 1 {
				fmt.Println("Realoading input file")
			}
		}
	}()
}

func addGroup() {
	/*
		Creates program group and executes all the procs :)
	*/
	fmt.Println("Adding a new group!")
}

func DiffConfigs(oldConfig, newConfig parser.ConfigFile) {

	oldPrograms := oldConfig.Programs
	newPrograms := newConfig.Programs

	for programName := range oldPrograms {
		if _, exists := newPrograms[programName]; !exists {
			fmt.Printf("Program '%s' removed.\n", programName)
			controller.KillGroup(programName)
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
