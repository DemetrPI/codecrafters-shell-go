package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

var path = strings.Split(os.Getenv("PATH"), ":")

func echo(args []string) {
	output := ""
	for index, element := range args {
		if index > 0 {
			output += element + " "
		}
	}
	fmt.Println(output)
}

func pwd() {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting current directory: %v\n", err)
	} else {
		fmt.Println(dir)
	}
}

func cd(args []string) {
	if len(args) == 0 {
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

func type_(args []string) {

	if len(args) > 0 {
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

func default_(args []string) {
	filePath := findExecutable(args[0], path)
	if filePath != "" {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		_ = cmd.Run()
		
	} else {
		fmt.Println(args[0] + ": command not found")
	}
}
