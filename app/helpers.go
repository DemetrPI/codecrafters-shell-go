package main

import (
	"os"
	"path/filepath"
	"strings"
)

// parses user input
func parseArgs(input string) []string {
    var (
        args      []string
        current   strings.Builder
        quoteChar rune // Use a rune (' or ") to track active quote. 0 means none.
        escaped   bool
    )

    for _, char := range input {
        // Handle an escaped character
        if escaped {
            current.WriteRune(char)
            escaped = false
            continue
        }

        // Set escaped flag if '\' is found outside of single quotes
        if char == '\\' && quoteChar != '\'' {
            escaped = true
            continue
        }

        // If we are inside a quoted section
        if quoteChar != 0 {
            if char == quoteChar {
                quoteChar = 0 // End of quoted section
            } else {
                current.WriteRune(char) // Add character to the current argument
            }
            continue
        }

        // If we are not in a quoted section
        switch char {
        case '\'', '"':
            quoteChar = char // Start of a new quoted section
        
        case ' ':
            if current.Len() > 0 {
                args = append(args, current.String())
                current.Reset()
            }
        
        default:
            current.WriteRune(char)
        }
    }

    // Add the final argument to the list
    if current.Len() > 0 {
        args = append(args, current.String())
    }

    return args
}


// checks for executable command
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
