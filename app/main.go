package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)


var cmdsMap = map[string]string{
	"echo": "a shell builtin",
	"type": "a shell builtin",
	"exit": "a shell builtin",
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
				if desc, ok := cmdsMap[subcommandName]; ok {
					fmt.Printf("%s is %s\n", subcommandName, desc)
				} else if path, err := exec.LookPath(subcommandName); err == nil {
					fmt.Printf("%s is %s\n", subcommandName, path)
				} else {
					fmt.Printf("%s: not found\n", subcommandName)
				}
			default:
				fmt.Println(command[:len(command)-1] + ": command not found")
			}
		}
	}
}
