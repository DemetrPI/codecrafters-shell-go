package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"
)

// maps command names and description
var cmdsMap = map[string]string{
	"echo": "a shell builtin",
	"type": "a shell builtin",
	"exit": "a shell builtin",
	"pwd":  "a shell builtin",
	"cd":   "a shell builtin",
}

func main() {
	for {

		fmt.Fprint(os.Stdout, "$ ")

		input, err := bufio.NewReader(os.Stdin).ReadString('\n')
		input = strings.TrimSpace(input)

		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading input:", err)
			os.Exit(1)
		}

		args := parseArgs(input)

		if args[0] == "exit" && len(args) > 0 && args[1] == "0" {
			os.Exit(0)
		}

		var outputFile *os.File
		for i, arg := range args[1:] {
			if (arg == ">" || arg == "1>") && i+1 < len(args) {
				if outputFile, err = os.Create(args[i+1]); err != nil {
					fmt.Fprintf(os.Stderr, "Error creating file: %v\n", err)
					continue
				}
				args = args[:i]
				break
			}
		}

		if outputFile != nil {
			defer outputFile.Close()
			os.Stdout = outputFile
		}

		switch args[0] {
		case "echo":
			echo(args)
		case "pwd":
			pwd()
		case "cd":
			cd(args)
		case "type":
			type_(args)
		default:
			default_(args)
		}
		if outputFile != nil {
			os.Stdout = os.NewFile(uintptr(syscall.Stdout), "/dev/stdout")
		}
	}
}
