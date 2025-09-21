package main

import (
	"fmt"
	"github.com/chzyer/readline"
	"log"
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
	"history": "a shell builtin",
}

var originalStdout *os.File
var originalStderr *os.File

func init() {
	originalStdout = os.Stdout
	originalStderr = os.Stderr
}

func main() {
	// populating trie with built-ins...
	trie := NewTrie()
	for cmd := range cmdsMap {
		trie.Insert(cmd)
	}

	// ... and customs
	populateTrieCustomFiles(trie)

	line, err := readline.NewEx(&readline.Config{
		Prompt:       "$ ",
		AutoComplete: trie,
	})
	if err != nil {
		log.Fatal(err)
	}

	for {
		// Always print the prompt to the original stdout
		fmt.Fprint(originalStdout, "$ ")

		input, err := line.Readline()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading input:", err)
			os.Exit(1)
		}
		input = strings.TrimSpace(input)

		parsed := parseArgs(input)
		if len(parsed) == 0 {
			continue
		}

		var outputFile, errFile *os.File
		var cleanedArgs []string

		// Parse args for redirection
		for i := 0; i < len(parsed); i++ {
			arg := parsed[i]
			switch arg {
			case ">", "1>", "2>":
				if i+1 < len(parsed) {
					f, ferr := os.Create(parsed[i+1])
					if ferr != nil {
						fmt.Fprintf(originalStderr, "Error creating file: %v\n", ferr)
						return
					}
					if arg == ">" || arg == "1>" {
						outputFile = f
					} else {
						errFile = f
					}
					i++
				}
			case ">>", "1>>", "2>>":
				if i+1 < len(parsed) {
					f, ferr := os.OpenFile(parsed[i+1], os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
					if ferr != nil {
						fmt.Fprintf(originalStderr, "Error creating file: %v\n", ferr)
						return
					}
					if arg == ">>" || arg == "1>>" {
						outputFile = f
					} else {
						errFile = f
					}
					i++
				}
			default:
				cleanedArgs = append(cleanedArgs, arg)
			}
		}

		// After parsing, replace parsed with cleanedArgs
		parsed = cleanedArgs

		if len(cleanedArgs) == 0 {
			continue
		}

		command := cleanedArgs[0]
		args := cleanedArgs[1:]

		if errFile != nil {
			os.Stderr = errFile
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
		if errFile != nil {
			errFile.Close()
			os.Stderr = originalStderr
		}

		if outputFile != nil {
			outputFile.Close()
			os.Stdout = originalStdout
		}
	}
}
