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

func init() {
	originalStdout = os.Stdout
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

        command := parsed[0]
        args := parsed[1:]
        
        var outputFile *os.File
        var redirectIndex int = -1

        // Find the redirection operator and filename
        for i, arg := range parsed {
            if arg == "1>" || arg == ">" {
                if i+1 < len(parsed) {
                    redirectIndex = i
                    break
                }
            }
        }
        
        if redirectIndex != -1 {
            outputFileName := parsed[redirectIndex+1]
            if outputFile, err = os.Create(outputFileName); err != nil {
                fmt.Fprintf(os.Stderr, "Error creating file: %v\n", err)
                continue
            }
            parsed = parsed[:redirectIndex]
            args = parsed[1:]
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
        if outputFile != nil {
            outputFile.Close()
            os.Stdout = originalStdout
        }
    }
}
