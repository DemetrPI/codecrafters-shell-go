package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"github.com/chzyer/readline"
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
var originalStderr *os.File

// init saves the original stdout and stderr file descriptors to restore them after redirection.
func init() {
	originalStdout = os.Stdout
	originalStderr = os.Stderr
}

func main() {
	// populating trie with built-ins...
	trie := NewTrie()
	history := new(History)

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
		input, err := line.Readline()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading input:", err)
			os.Exit(1)
		}
		history.storeHistory(input)

		input = strings.TrimSpace(input)

		parsed := parseArgs(input)
		if len(parsed) == 0 {
			continue
		}

		var outputFile, errFile *os.File
		var cleanedArgs []string
		var redirectError bool

		// Parse args for redirection
		for i := 0; i < len(parsed); i++ {
			arg := parsed[i]
			isRedirect := true

			var isAppend, isStderr bool
			switch arg {
			case ">", "1>":
			case "2>":
				isStderr = true
			case ">>", "1>>":
				isAppend = true
			case "2>>":
				isAppend, isStderr = true, true
			default:
				isRedirect = false
				cleanedArgs = append(cleanedArgs, arg)
			}

			if isRedirect {
				if i+1 >= len(parsed) {
					fmt.Fprintln(originalStderr, "shell: syntax error near unexpected token `newline'")
					redirectError = true
					break
				}
				filename := parsed[i+1]
				flags := os.O_WRONLY | os.O_CREATE
				if isAppend {
					flags |= os.O_APPEND
				} else {
					flags |= os.O_TRUNC
				}

				f, ferr := os.OpenFile(filename, flags, 0644)
				if ferr != nil {
					fmt.Fprintf(originalStderr, "Error opening file: %v\n", ferr)
					redirectError = true
					break
				}

				targetFile := &outputFile
				if isStderr {
					targetFile = &errFile
				}
				if *targetFile != nil {
					(*targetFile).Close()
				}
				*targetFile = f
				i++ // Also skip the filename
			}
		}

		// After parsing, replace parsed with cleanedArgs
		parsed = cleanedArgs

		if redirectError || len(parsed) == 0 {
			if errFile != nil {
				errFile.Close()
			}
			if outputFile != nil {
				outputFile.Close()
			}
			continue
		}

		command := parsed[0]
		args := parsed[1:]

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
		case "history":
			history.printHistory()
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
