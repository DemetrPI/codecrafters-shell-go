package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// parseArgs splits a command line string into arguments, handling quotes and escapes.
func parseArgs(input string) []string {
	var (
		args      []string
		current   strings.Builder
		quoteChar rune // 0 indicates not inside quotes, '\'' or '"' indicates the active quote type
		escaped   bool
	)

	for _, char := range input {
		// If the previous character was an escape character...
		if escaped {
			// Inside double quotes, '\' only escapes certain characters.
			// For others, the backslash is treated as a literal character.
			if quoteChar == '"' {
				if char == '$' || char == '"' || char == '\\' {
					current.WriteRune(char)
				} else {
					current.WriteRune('\\')
					current.WriteRune(char)
				}
			} else {
				// Outside of double quotes, '\' escapes the next character literally.
				current.WriteRune(char)
			}
			escaped = false
			continue
		}

		switch char {
		// An escape character is only special if not inside single quotes.
		case '\\':
			if quoteChar != '\'' {
				escaped = true
			} else {
				current.WriteRune(char)
			}
		// Handle the start and end of quoted sections.
		case '\'', '"':
			switch quoteChar {
			case 0:
				quoteChar = char // Start quoting
			case char:
				quoteChar = 0 // Stop quoting
			default:
				current.WriteRune(char) // Other quote type, treat as literal
			}
		// Spaces are delimiters only when not in quotes.
		case ' ':
			if quoteChar == 0 {
				if current.Len() > 0 {
					args = append(args, current.String())
					current.Reset()
				}
			} else {
				current.WriteRune(' ') // It's a literal space
			}
		// Any other character is part of the argument.
		default:
			current.WriteRune(char)
		}
	}

	// Add the last argument if it exists.
	if current.Len() > 0 {
		args = append(args, current.String())
	}

	return args
}

// findExecutable checks for an executable command in the given paths.
func findExecutable(command string, paths []string) string {
	for _, path := range paths {
		fullPath := filepath.Join(path, command)
		fileInfo, err := os.Stat(fullPath)
		if err == nil && fileInfo.Mode().Perm()&0111 != 0 {
			return fullPath
		}
	}
	return ""
}

// populateTrieCustomFiles populates the Trie with executable file names from the system's PATH.
func populateTrieCustomFiles(t *Trie) {
	pathEnv := os.Getenv("PATH")
	paths := strings.SplitSeq(pathEnv, ":")
	for path := range paths {
		files, err := os.ReadDir(path)
		if err != nil {
			continue
		}
		for _, file := range files {
			if !file.IsDir() {
				fullPath := filepath.Join(path, file.Name())
				info, err := os.Stat(fullPath)
				if err == nil && info.Mode().Perm()&0111 != 0 {
					t.Insert(file.Name())
				}
			}
		}
	}
}

// longestCommonPrefix finds the longest common prefix string amongst a slice of strings.
func longestCommonPrefix(strs []string) string {
	// no LCP
	if len(strs) == 0 {
		return ""
	}
	// Use the first string as the reference.
	for i := 0; i < len(strs[0]); i++ {
		char := strs[0][i]
		// Check this character against all other strings.
		for j := 1; j < len(strs); j++ {
			// If a string is shorter or the character doesn't match, we've found the LCP.
			if i >= len(strs[j]) || strs[j][i] != char {
				return strs[0][:i]
			}
		}
	}
	// If the loop completes, the entire first string is the common prefix.
	return strs[0]
}

func handleRedirections(parsed []string) (
	cleanedArgs []string,
	outputFile *os.File,
	errFile *os.File,
	err error) {

	// Iterate through all parsed arguments to find and handle redirection operators.
	for i := 0; i < len(parsed); i++ {
		arg := parsed[i]
		isRedirect := true

		var isAppend, isStderr bool
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
				return nil, outputFile, errFile, fmt.Errorf("shell: syntax error near unexpected token `newline'")
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
				return nil, outputFile, errFile, fmt.Errorf("error opening file: %v", ferr)
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
	return cleanedArgs, outputFile, errFile, nil
}

