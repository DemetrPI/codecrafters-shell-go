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
	lines         []string
	savedLinesLen int
}

// cmdsMap maps built-in command names to their descriptions.
var cmdsMap = map[string]string{
	"echo":    "a shell builtin",
	"type":    "a shell builtin",
	"exit":    "a shell builtin",
	"pwd":     "a shell builtin",
	"cd":      "a shell builtin",
	"history": "a shell builtin",
	"jobs":    "a shell builtin",
}

var path = strings.Split(os.Getenv("PATH"), ":")

// echo implements the "echo" built-in command.
func echo(args []string) {
	fmt.Println(strings.Join(args[1:], " "))
}

//jobs implements the "jobs" built-in command.
func jobs() {
	fmt.Println("")
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

func exit_(args []string, h *History) {
	h.save()
	if len(args) > 1 {
		code, err := strconv.Atoi(args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "exit: %s: numeric argument required\n", args[1])
			return
		}
		os.Exit(code)
	}
	os.Exit(0)
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

// load reads the history from the file specified by the HISTFILE environment variable.
func (h *History) load() {
	histfilePath := os.Getenv("HISTFILE") // Get the latest value
	if histfilePath == "" {
		return
	}
	data, err := os.ReadFile(histfilePath)
	if err != nil {
		return // Not an error if the file doesn't exist
	}
	for line := range strings.SplitSeq(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			h.storeHistory(line)
		}
	}
	h.savedLinesLen = len(h.lines)
}

// save writes the history to the file specified by the HISTFILE environment variable.
func (h *History) save() {
	histfilePath := os.Getenv("HISTFILE") // Get the latest value
	if histfilePath == "" {
		return
	}

	// Open the file for writing, create if not exists, truncate if exists.
	file, err := os.OpenFile(histfilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "history: error saving to %s: %v\n", histfilePath, err)
		return
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, line := range h.lines {
		fmt.Fprintln(writer, line)
	}
	err = writer.Flush()
	if err != nil {
		fmt.Fprintf(os.Stderr, "history: error flushing buffer to %s: %v\n", histfilePath, err)
		return
	}

	h.savedLinesLen = len(h.lines) // All lines are now saved
}

// handleHistoryCommand processes the `history` built-in command and its arguments.
func (h *History) handleHistoryCommand(args []string) {
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
		// After reading, all lines are considered "saved".
		h.savedLinesLen = len(h.lines)
		// Do not print history after reading from file
		return
	case "-w", "-a": // Handle both write and append
		if len(args) < 3 {
			fmt.Fprintf(os.Stderr, "history: %s: option requires an argument\n", args[1])
			return
		}
		filename := args[2]

		// Determine file opening flags based on the option
		flags := os.O_WRONLY | os.O_CREATE
		var linesToWrite []string

		if args[1] == "-w" {
			flags |= os.O_TRUNC // Overwrite for -w
			linesToWrite = h.lines
		} else { // -a
			flags |= os.O_APPEND // Append for -a
			// Only write new lines for -a
			linesToWrite = h.lines[h.savedLinesLen:]
		}

		file, err := os.OpenFile(filename, flags, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "history: %s: %v\n", filename, err)
			return
		}
		defer file.Close()

		// Write each history line to the file
		writer := bufio.NewWriter(file)
		for _, line := range linesToWrite {
			fmt.Fprintln(writer, line)
		}
		err = writer.Flush() // Ensure all buffered content is written
		if err != nil {
			fmt.Fprintf(os.Stderr, "error writing buffer: %s\n", err)
		}
		// After writing, all lines are now considered "saved".
		h.savedLinesLen = len(h.lines)
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
		history.handleHistoryCommand(cleanedArgs)
	case "exit":
		exit_(cleanedArgs, history)
	case "jobs":
		
	default:
		default_(cleanedArgs)
	}
}
