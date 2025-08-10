package main

import (
	"os"
	"path/filepath"
	"strings"
)

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
