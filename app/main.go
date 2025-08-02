package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var cmdsMap = map[string]string{
	"echo": "a shell builtin",
	"type": "a shell builtin",
	"exit": "a shell builtin",
}

func findExecutable(command string, paths []string) string {
	for _, path := range paths {
		fullPath := filepath.Join(path, command)
		fileInfo, err := os.Stat(fullPath)
		if err == nil && fileInfo.Mode().Perm()&0111 != 0 {
			return fullPath
		}
	}
	return ""
}

func main() {
	for {
		paths := strings.Split(os.Getenv("PATH"), ":")
		fmt.Fprint(os.Stdout, "$ ")
		line, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading input:", err)
			os.Exit(1)
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Fields(line) 
		command := parts[0]
		args := parts[1:]
		if command == "exit" && len(args) > 0 && args[0] == "0" {
			os.Exit(0)
		}

		switch command {
		case "echo":
			fmt.Println(strings.Join(args, " "))

		case "type":
			if len(args) > 0 {
				target := args[0]
				if decs, ok := cmdsMap[target]; ok {
					fmt.Printf("%s is %s\n", target, decs)
				} else {
					filePath := findExecutable(target, paths)
					if filePath != "" {
						fmt.Printf("%s is %s\n", target, filePath)
					} else {
						fmt.Printf("%s: not found\n", target)
					}
				}
			} else {
				fmt.Println("type: not enough arguments")
			}

		default:
			filePath := findExecutable(command, paths)
			if filePath != "" {
				cmd := exec.Command(command, args...)
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				cmd.Stdin = os.Stdin
				err := cmd.Run()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error running command: %v\n", err)
				}
			} else {
				fmt.Println(command + ": command not found")
			}
		}
	}
}
