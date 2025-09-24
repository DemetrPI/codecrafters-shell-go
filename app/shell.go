package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/chzyer/readline"
)

// Shell holds the state of the shell application.
type Shell struct {
	history        *History
	trie           *Trie
	line           *readline.Instance
	originalStdout *os.File
	originalStderr *os.File
}

// NewShell creates and initializes a new Shell instance.
func NewShell() *Shell {
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

	history := &History{}
	history.load()

	s := &Shell{
		history:        history,
		trie:           trie,
		line:           line,
		originalStdout: os.Stdout,
		originalStderr: os.Stderr,
	}

	return s
}

// Run starts the main read-eval-print loop of the shell.
func (s *Shell) Run() {
	defer s.line.Close()
	for {
		input, err := s.line.Readline()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading input:", err)
			os.Exit(1)
		}
		s.history.storeHistory(input)

		input = strings.TrimSpace(input)

		parsed := parseArgs(input)
		if len(parsed) == 0 {
			continue
		}
		cleanedArgs, outputFile, errFile, err := handleRedirections(parsed)
		if err != nil {
			fmt.Fprintln(s.originalStderr, err)
			// clean up any open files and continue
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

		executeCommand(command, cleanedArgs, s.history)

		// Restore os.Stdout and os.Stderr to their original values after the command executes
		if errFile != nil {
			errFile.Close()
			os.Stderr = s.originalStderr
		}

		if outputFile != nil {
			outputFile.Close()
			os.Stdout = s.originalStdout
		}
	}
}
