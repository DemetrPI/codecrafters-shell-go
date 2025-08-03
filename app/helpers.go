package main

import (
	"os"
	"path/filepath"
	"strings"
)

// parses user input
func parseArgs(input string) []string {
	var (
		args    []string
		current strings.Builder
		inSQuote, inDQuote, escaped bool
	)

	for _, char := range input {
		switch {
		case escaped:
			current.WriteRune(char)
			escaped = false

		case char == '\\' && !inSQuote && !inDQuote:
			escaped = true

		case char == '\'' && !inDQuote:
			inSQuote = !inSQuote

		case char == '"' && !inSQuote:
			inDQuote = !inDQuote

		case char == ' ' && !inSQuote && !inDQuote:
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}

		default:
			current.WriteRune(char)
		}
	}

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
