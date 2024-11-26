package signals

import (
	"fmt"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
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

	for i := 1; i < 65; i++ {
		sig := syscall.Signal(i)
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

// func sendSignal(signal_name string, pid int) error {
// 	var sig syscall.Signal

// 	switch signal_name {
// 	case "SIGTERM":
// 		sig = syscall.SIGTERM
// 	case "SIGKILL":
// 		sig = syscall.SIGKILL
// 	case "SIGINT":
// 		sig = syscall.SIGINT
// 	case "SIGSTOP":
// 		sig = syscall.SIGSTOP
// 	case "SIGUSR1":
// 		sig = syscall.SIGUSR1
// 	case "SIGUSR2":
// 		sig = syscall.SIGUSR2
// 	default:
// 		return fmt.Errorf("invalid signal: %s", signal_name)
// 	}

// 	err := syscall.Kill(pid, sig)
// 	if err != nil {
// 		return fmt.Errorf("failed to send signal: %v", err)
// 	}
// 	return nil
// }
