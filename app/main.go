package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/chzyer/readline"
)

func main() {
	// creating a History instance
	history := &History{}
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

		// Iterate through all parsed arguments to find and handle redirection operators.
		for i := 0; i < len(parsed); i++ {
			arg := parsed[i]
			isRedirect := true

			var isAppend, isStderr bool
			// Check if the current argument is a redirection operator.
			switch arg {
			case ">", "1>":
				// Standard output redirection (overwrite).
			case "2>":
				// Standard error redirection (overwrite).
				isStderr = true
			case ">>", "1>>":
				// Standard output redirection (append).
				isAppend = true
			case "2>>":
				// Standard error redirection (append).
				isAppend, isStderr = true, true
			default:
				// If the argument is not a redirection operator, it's part of the command itself.
				isRedirect = false
				cleanedArgs = append(cleanedArgs, arg)
			}

			// This block executes if a redirection operator was found.
			if isRedirect {
				// A redirection operator must be followed by a filename.
				if i+1 >= len(parsed) {
					fmt.Fprintln(originalStderr, "shell: syntax error near unexpected token `newline'")
					redirectError = true
					break
				}
				filename := parsed[i+1]
				// Set the file opening flags based on whether we are appending or truncating.
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

				// Determine whether to redirect stdout or stderr.
				targetFile := &outputFile
				if isStderr {
					targetFile = &errFile
				}
				// If a file is already open for this stream (e.g., `echo hi > a > b`), close the previous one.
				if *targetFile != nil {
					(*targetFile).Close()
				}
				*targetFile = f
				// Increment the loop counter to skip the filename argument in the next iteration.
				i++
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
			history.printHistory(parsed)
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
