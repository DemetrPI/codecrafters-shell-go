package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// History stores the history of commands entered in the shell.
type History struct {
	lines []string
}

// cmdsMap maps built-in command names to their descriptions.
var cmdsMap = map[string]string{
	"echo":    "a shell builtin",
	"type":    "a shell builtin",
	"exit":    "a shell builtin",
	"pwd":     "a shell builtin",
	"cd":      "a shell builtin",
	"history": "a shell builtin",
}

// originalStdout and originalStderr store the original standard output and error streams.
var originalStdout *os.File

// path stores the directories from the PATH environment variable.
var path = strings.Split(os.Getenv("PATH"), ":")

// echo implements the "echo" built-in command.
func echo(args []string) {
	fmt.Println(strings.Join(args[1:], " "))
}

// pwd implements the "pwd" built-in command.
func pwd() {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting current directory: %v\n", err)
	} else {
		fmt.Println(dir)
	}
}

// cd implements the "cd" built-in command.
func cd(args []string) {
	if len(args) < 2 {
		fmt.Println("cd: not enough arguments")
		return
	}

	target := args[1]

	if target == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting home directory: %v\n", err)
			return
		}
		target = home
	}

	if err := os.Chdir(target); err != nil {
		fmt.Printf("cd: %v: No such file or directory\n", target)
	}
}

// type_ implements the "type" built-in command.
func type_(args []string) {

	if len(args) > 1 {
		target := args[1]
		if decs, ok := cmdsMap[target]; ok {
			fmt.Printf("%s is %s\n", target, decs)
		} else {
			filePath := findExecutable(target, path)
			if filePath != "" {
				fmt.Printf("%s is %s\n", target, filePath)
			} else {
				fmt.Printf("%s: not found\n", target)
			}
		}
	} else {
		fmt.Println("type: not enough arguments")
	}
}

// default_ handles execution of external commands.
func default_(args []string) {
	_, err := exec.LookPath(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: command not found\n", args[0])
		return
	}
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	_ = cmd.Run()
}

// storeHistory adds a line to the command history.
func (h *History) storeHistory(input string) {
	h.lines = append(h.lines, input)
}

// printHistory displays the command history with line numbers.
func (h *History) printHistory(args []string) {
	// If no arguments, print all history
	if len(args) == 1 {
		for i := 0; i < len(h.lines); i++ {
			fmt.Printf("%d %s\n", i+1, h.lines[i])
		}
		return
	}

	// Handle arguments
	switch args[1] {
	case "-r":
		// Expecting filename as args[2]
		if len(args) < 3 {
			fmt.Fprintln(os.Stderr, "history: -r: option requires an argument")
			return
		}
		filename := args[2]
		file, err := os.OpenFile(filename, os.O_RDONLY, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "history: %s: %v\n", filename, err)
			return
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			h.storeHistory(scanner.Text())
		}
		// Do not print history after reading from file
		return
	default:
		// Assume it's a numeric argument
		num, err := strconv.Atoi(args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "history: %s: numeric argument required\n", args[1])
			return
		}
		if num < 0 {
			fmt.Fprintf(os.Stderr, "history: %s: history position out of range\n", args[1])
			return
		}
		historyToShow := num
		start := max(len(h.lines)-historyToShow, 0)
		for i := start; i < len(h.lines); i++ {
			fmt.Printf("%d %s\n", i+1, h.lines[i])
		}
		return
	}
}

// init saves the original stdout file descriptors to restore them after redirection.
func init() {
	originalStdout = os.Stdout
}

func executeCommand(command string, cleanedArgs []string, history *History) {
	switch command {
	case "echo":
		echo(cleanedArgs)
	case "pwd":
		pwd()
	case "cd":
		cd(cleanedArgs)
	case "type":
		type_(cleanedArgs)
	case "history":
		history.printHistory(cleanedArgs)
	default:
		default_(cleanedArgs)
	}
}
