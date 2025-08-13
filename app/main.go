package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// maps command names and description
var cmdsMap = map[string]string{
	"echo": "a shell builtin",
	"type": "a shell builtin",
	"exit": "a shell builtin",
	"pwd":  "a shell builtin",
	"cd":   "a shell builtin",
}

var originalStdout *os.File
var originalStderr *os.File

func init() {
	originalStdout = os.Stdout
	originalStderr = os.Stderr
}

func main() {
	for {
		// Always print the prompt to the original stdout
		fmt.Fprint(originalStdout, "$ ")

		input, err := bufio.NewReader(os.Stdin).ReadString('\n')
		input = strings.TrimSpace(input)

		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading input:", err)
			os.Exit(1)
		}

		parsed := parseArgs(input)
		if len(parsed) == 0 {
			continue
		}

		var outputFile, errFile *os.File
		var cleanedArgs []string

		// Parse args for redirection

		for i := 0; i < len(parsed); i++ {
			arg := parsed[i]

			switch arg {
			case ">", "1>":
				if i+1 < len(parsed) {
					f, ferr := os.Create(parsed[i+1])
					if ferr != nil {
						fmt.Fprintf(originalStderr, "Error creating file: %v\n", ferr)
						return
					}
					outputFile = f
					i++ // skip the filename
				}
			case "2>":
				if i+1 < len(parsed) {
					f, ferr := os.Create(parsed[i+1])
					if ferr != nil {
						fmt.Fprintf(originalStderr, "Error creating file: %v\n", ferr)
						return
					}
					errFile = f
					i++ // skip the filename
				}
			default:
				cleanedArgs = append(cleanedArgs, arg)
			}
		}

		// After parsing, replace parsed with cleanedArgs
		parsed = cleanedArgs

		if len(cleanedArgs) == 0 {
			continue
		}

		command := cleanedArgs[0]
		args := cleanedArgs[1:]

		if errFile != nil {
			os.Stderr = errFile
		}
		if outputFile != nil {
			os.Stdout = outputFile
		}

		if command == "exit" && len(args) > 0 && args[0] == "0" {
			os.Exit(0)
		}

		switch command {
		case "echo":
			echo(parsed)
		case "pwd":
			pwd()
		case "cd":
			cd(parsed)
		case "type":
			type_(parsed)
		default:
			default_(parsed)
		}

		// Restore os.Stdout and os.Stderr to their original values after the command executes
		if errFile != nil {
			errFile.Close()
			os.Stderr = originalStderr
		}

		if outputFile != nil {
			outputFile.Close()
			os.Stdout = originalStdout
		}
	}
}
