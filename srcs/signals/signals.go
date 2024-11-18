package signals

import (
	"fmt"
	"syscall"
)

func sendSignal(signal_name string, pid int) error {
	var sig syscall.Signal

	switch signal_name {
	case "SIGTERM":
		sig = syscall.SIGTERM
	case "SIGKILL":
		sig = syscall.SIGKILL
	case "SIGINT":
		sig = syscall.SIGINT
	case "SIGSTOP":
		sig = syscall.SIGSTOP
	case "SIGUSR1":
		sig = syscall.SIGUSR1
	case "SIGUSR2":
		sig = syscall.SIGUSR2
	default:
		return fmt.Errorf("invalid signal: %s", signal_name)
	}

	err := syscall.Kill(pid, sig)
	if err != nil {
		return fmt.Errorf("failed to send signal: %v", err)
	}
	return nil
}
