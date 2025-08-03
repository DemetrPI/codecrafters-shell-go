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
	}
}
