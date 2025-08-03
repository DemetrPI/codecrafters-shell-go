package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
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

// checks for executable command
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

// removes all single quotes from inside the string
func removeQuotes(s string) string {
	return strings.ReplaceAll(s, "'", "")
}

// splits the line into args, preserving quotes
func splitArgs(line string) []string {
	re := regexp.MustCompile(`(?:'[^']*')+|\S+`)
	parts := re.FindAllString(line, -1)
	var args []string
	for _, part := range parts {
		args = append(args, removeQuotes(part))
	}
	return args
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

		args := splitArgs(line)
		command := args[0]
		args = args[1:]

		if command == "exit" && len(args) > 0 && args[0] == "0" {
			os.Exit(0)
		}
		switch command {
		case "echo":
			fmt.Println(strings.Join(args, " "))
		case "pwd":
			dir, err := os.Getwd()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error getting current directory: %v\n", err)
			} else {
				fmt.Println(dir)
			}
		case "cd":
			if len(args) == 0 {
				fmt.Println("cd: not enough arguments")
			}
			target := args[0]
			if target == "~" {
				target, err = os.UserHomeDir()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error getting home directory: %v\n", err)
				}
			}
			if err := os.Chdir(target); err != nil {
				fmt.Printf("cd: %v: No such file or directory\n", target)
			}
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
