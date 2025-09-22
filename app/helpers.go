package main

import (
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
