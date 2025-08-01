package main

import (
	"bufio"
	"fmt"
	"os"
)

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
			fmt.Println(command[:len(command)-1] + ": command not found")
		}
	}

}
