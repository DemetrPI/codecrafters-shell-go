package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
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
		// \033[33m = Yellow
		// \033[0m  = Reset (so your typed commands aren't also yellow)
		Prompt:       "\033[33m$ \033[0m",
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
		commandSegments := splitCommandsByPipe(input)

		if len(commandSegments) == 0 {
			continue
		}

		// If there is a pipeline, execute it.
		if len(commandSegments) > 1 {
			if err := s.executePipeline(commandSegments); err != nil {
				fmt.Fprintln(s.originalStderr, err)
			}
		} else { // Otherwise, execute a single command.
			parsed := parseArgs(input)
			if len(parsed) == 0 {
				continue
			}
			cleanedArgs, outputFile, errFile, err := handleRedirections(parsed)
			if err != nil {
				fmt.Fprintln(s.originalStderr, err)
				// clean up any open files and continue
				if outputFile != nil {
					outputFile.Close()
				}
				if errFile != nil {
					errFile.Close()
				}
				continue
			}

			command := cleanedArgs[0]

			// Redirect stdout/stderr if needed
			if errFile != nil {
				os.Stderr = errFile
			}
			if outputFile != nil {
				os.Stdout = outputFile
			}

			executeCommand(command, cleanedArgs, s.history)

			// Restore stdout/stderr and close files
			s.restoreStdio(outputFile, errFile)
		}
	}
}

// executePipeline handles the execution of a series of commands linked by pipes.
func (s *Shell) executePipeline(commandSegments []string) error {
	var inputPipe io.Reader = os.Stdin
	var runningCommands []*exec.Cmd
	var pipesToClose []io.Closer

	// Variables to hold the final command's redirection files for cleanup
	var finalOutputFile *os.File
	var finalErrFile *os.File

	// 1. Setup and Start All Commands
	for i, segment := range commandSegments {
		parsed := parseArgs(strings.TrimSpace(segment))
		if len(parsed) == 0 {
			continue
		}

		// handleRedirections parses redirection operators and opens files
		cleanedArgs, outputFile, errFile, err := handleRedirections(parsed)
		if err != nil {
			// Ensure any files opened by handleRedirections are closed on parse error
			if outputFile != nil {
				outputFile.Close()
			}
			if errFile != nil {
				errFile.Close()
			}
			return err
		}

		command := cleanedArgs[0]
		isLastCommand := i == len(commandSegments)-1

		// --- Built-in Check ---
		if _, ok := cmdsMap[command]; ok {
			var r *os.File
			var w *os.File
			var err error

			// Save original stdout/stderr before redirection
			oldStdout := os.Stdout
			oldStderr := os.Stderr

			// 1. Setup Output Pipe/Redirection
			if !isLastCommand {
				// Not the last command: built-in output must go to a new pipe
				// Create a pipe for the built-in's output to be consumed by the next command
				r, w, err = os.Pipe()
				if err != nil {
					return fmt.Errorf("shell: pipe error: %v", err)
				}
				os.Stdout = w // Built-in output goes to the pipe's write end

				// We track the read end for closing and set it as the next input
				pipesToClose = append(pipesToClose, r)
				inputPipe = r // Input for the next command
			} else {
				// Last command: built-in output goes to specified file or original stdout
				if outputFile != nil {
					os.Stdout = outputFile
					finalOutputFile = outputFile // Track for final cleanup
				} else {
					os.Stdout = s.originalStdout
				}
			}

			// 2. Setup Stderr Redirection
			if errFile != nil {
				os.Stderr = errFile
				if isLastCommand {
					finalErrFile = errFile // Track for final cleanup
				}
			} else {
				os.Stderr = s.originalStderr
			}

			// 3. Execution: executeCommand uses the currently set os.Stdout/os.Stderr.
			// Note: For 'echo' and 'type', we don't need to explicitly read from 'inputPipe'.
			executeCommand(command, cleanedArgs, s.history)

			// 4. Cleanup
			// Close the write end of the pipe immediately if one was created
			if w != nil {
				w.Close()
			}

			// Restore original stdout/stderr
			os.Stdout = oldStdout
			os.Stderr = oldStderr

			// For non-last commands, we need to close any redirection files opened
			// by handleRedirections, as their output went to a pipe instead.
			if !isLastCommand {
				s.restoreStdio(outputFile, errFile) // This closes files if they were opened
			}

			continue // Move to the next command in the pipeline
		}

		// --- External Command Execution Setup ---

		cmd := exec.Command(command, cleanedArgs[1:]...)
		cmd.Stdin = inputPipe
		cmd.Stderr = os.Stderr // Default stderr

		if !isLastCommand {
			// Current command is not the last: its output must go to a pipe.
			pipe, err := cmd.StdoutPipe()
			if err != nil {
				// Close files if error occurs before start
				if outputFile != nil {
					outputFile.Close()
				}
				if errFile != nil {
					errFile.Close()
				}
				return fmt.Errorf("shell: pipe error: %v", err)
			}

			pipesToClose = append(pipesToClose, pipe) // Track pipe
			inputPipe = pipe                          // Input for the next command

			// FIX 2: Ignore redirection files for non-last commands (output must go to pipe)
			// Close them immediately if handleRedirections found them.
			if outputFile != nil {
				outputFile.Close()
			}
			if errFile != nil {
				errFile.Close()
			}

		} else {
			// Current command is the last: its output goes to file or os.Stdout.
			if outputFile != nil {
				cmd.Stdout = outputFile
				finalOutputFile = outputFile // Store for final cleanup
			} else {
				cmd.Stdout = os.Stdout
			}

			if errFile != nil {
				cmd.Stderr = errFile
				finalErrFile = errFile // Store for final cleanup
			}
		}

		if err := cmd.Start(); err != nil {
			return fmt.Errorf("shell: error starting command %s: %v", command, err)
		}
		runningCommands = append(runningCommands, cmd)
	}

	// 2. Wait for All Commands to Finish
	for _, cmd := range runningCommands {
		if err := cmd.Wait(); err != nil {
			return fmt.Errorf("shell: command exited with error: %v", err)
		}
	}

	// 3. Close Pipes
	for _, pipe := range pipesToClose {
		if err := pipe.Close(); err != nil {
			errorMsg := err.Error()

			// Check for known non-fatal errors caused by OS cleanup
			if strings.Contains(errorMsg, "file already closed") || strings.Contains(errorMsg, "close |0") {
				// Suppress printing of this specific, non-fatal error to pass the test
				continue
			}

			// Print only genuine, unexpected errors
			fmt.Fprintf(os.Stderr, "error closing pipe: %v\n", err)
		}
	}

	// s.restoreStdio is called here to close the final output/error files if they were opened.
	s.restoreStdio(finalOutputFile, finalErrFile)

	return nil
}

// restoreStdio restores stdout and stderr and closes the provided files.
func (s *Shell) restoreStdio(out, err *os.File) {
	if out != nil {
		out.Close()
		os.Stdout = s.originalStdout
	}
	if err != nil {
		err.Close()
		os.Stderr = s.originalStderr
	}
}
