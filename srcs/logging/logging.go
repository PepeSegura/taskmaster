package logging

import (
	"fmt"
	"os"
	"time"
)

var Filename string

func Init(name string) {
	Filename = name
}

func openFile(path string, truncate bool) *os.File {
	var outputFile *os.File
	var err error
	if truncate {
		outputFile, err = os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	} else {
		outputFile, err = os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	}
	if err != nil {
		fmt.Printf("Error opening output file: %v", err)
		return os.Stderr
	}
	return outputFile
}

func printMsg(mode, message string) {
	file := openFile(Filename, false)
	defer func() {
		file.Close()
	}()
	dt := time.Now()
	buffer := dt.Format("[01-02-2006 15:04:05] : ")
	buffer += fmt.Sprintf("%-8s - %s\n", mode, message)
	file.Write([]byte(buffer))
}

func Debug(message string) {
	printMsg("Debug", message)
}

func Info(message string) {
	printMsg("Info", message)
}

func Warning(message string) {
	printMsg("Warning", message)
}

func Error(message string) {
	printMsg("Error", message)
}

func Fatal(message string) {
	printMsg("Fatal", message)
	os.Exit(1)
}
