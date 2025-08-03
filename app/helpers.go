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
		quoteChar rune
	)

	for _, char := range input {
		switch {
		case (char == '\'' || char == '"'):
			switch quoteChar {
			case 0:
				quoteChar = char
			case char:
				quoteChar = 0
			default:
				current.WriteRune(char)
			}
		case char == ' ' && quoteChar == 0:
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
