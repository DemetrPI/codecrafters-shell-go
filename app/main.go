package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type command struct {
	name        string
	description string
}
type commands []command

var cmds = commands{
	
		{
			name:        "echo",
			description: "is a shell builtin",
		},
		{
			name:        "type",
			description: "is a shell builtin",
		},
		{
			name:        "exit",
			description: "is a shell builtin",
		},
	
}

func getCommandDescription(name string) string {
	for _, cmd := range cmds {
		if cmd.name == name {
			return cmd.description
		}
	}
	return ""
}

func main() {

	for {
		fmt.Fprint(os.Stdout, "$ ")
		command, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading input:", err)
			os.Exit(1)
		}
		if command == "exit 0\n" {
			os.Exit(0)
		} else {
			subcommands := strings.Split(command, " ")
			switch subcommands[0] {
			case "echo":
				fmt.Printf("%s", strings.Join(subcommands[1:], " "))
			case "type":
				subcommandName := strings.TrimSpace(subcommands[1])
				commandDescription := getCommandDescription(subcommandName)
				if commandDescription != "" {
					fmt.Printf("%s %s\n", subcommandName, commandDescription)
				} else {
					fmt.Printf("%s: command not found\n", subcommandName)
				}
			default:
				fmt.Println(command[:len(command)-1] + ": command not found")
			}
		}
	}
}
