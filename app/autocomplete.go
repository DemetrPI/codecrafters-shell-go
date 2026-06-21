package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Node is a single node in the trie.
// Each node maps a character to the next node in the path.
type Node struct {
	Children    map[string]*Node
	IsEndOfWord bool // true if this node marks the end of a complete word
}

// Trie represents a trie data structure for autocomplete.
type Trie struct {
	RootNode   *Node
	lastPrefix string // Stores the prefix from the last tab completion attempt.
	tabCount   int    // Counts consecutive tab presses for the same prefix.
}

// NewNode initializes a new Trie node.
func NewNode() *Node {
	node := &Node{
		Children: make(map[string]*Node),
	}
	return node
}

// NewTrie creates and returns a new Trie.
func NewTrie() *Trie {
	root := NewNode()
	return &Trie{RootNode: root}
}

// Insert adds a word to the trie.
func (t *Trie) Insert(word string) {
	current := t.RootNode
	for _, r := range word {
		char := string(r)
		node, ok := current.Children[char]
		if !ok {
			node = NewNode()
			current.Children[char] = node
		}
		current = node
	}
	// Mark the end of the word on the *last* node for the word.
	current.IsEndOfWord = true
}

// findAllWords is a recursive helper to find all words from a given node.
func (t *Trie) findAllWords(node *Node, prefix string, words *[]string) {
	if node.IsEndOfWord {
		*words = append(*words, prefix)
	}
	keys := make([]string, 0, len(node.Children))
	for char := range node.Children {
		keys = append(keys, char)
	}
	sort.Strings(keys)
	for _, char := range keys {
		childNode := node.Children[char]
		t.findAllWords(childNode, prefix+char, words)
	}
}

// FindCompletions finds all words in the trie with a given prefix.
func (t *Trie) FindCompletions(prefix string) []string {
	current := t.RootNode
	for _, r := range prefix {
		char := string(r)
		node, ok := current.Children[char]
		if !ok {
			// No words with this prefix
			return []string{}
		}
		current = node
	}
	var completions []string
	t.findAllWords(current, prefix, &completions)
	return completions
}

// Do implements the chzyer/readline.AutoCompleter interface.
// It is called automatically by the readline library when the user presses Tab.
// It returns a list of completion suffixes and the length of the prefix to replace.
func (t *Trie) Do(line []rune, pos int) (newLine [][]rune, length int) {
	lineStr := string(line[:pos])
	var fileCompletions []string
	lastSpace := strings.LastIndex(lineStr, " ")
	// prefix is the current word being typed (after the last space)
	prefix := lineStr[lastSpace+1:]
	// Look up command completions from the trie
	completions := t.FindCompletions(prefix)
	suggestions := make([][]rune, len(completions))
	// If there's no space yet, the user is still typing the command name
	isFirstWord := lastSpace == -1

	// Pre-compute suffix suggestions for command completions (strip the already-typed prefix)
	for i, comp := range completions {
		suggestions[i] = []rune(strings.TrimPrefix(comp, prefix))
	}

	if isFirstWord {
		switch len(completions) {
		case 0:
			// No matches: ring the bell.
			fmt.Fprint(NewShell().originalStdout, "\a")
			return nil, len(prefix)
		case 1:
			// Unique match: complete and add a trailing space
			suggestions[0] = append(suggestions[0], ' ')
			return suggestions, len(prefix)
		default:
			// Multiple matches: complete up to the longest common prefix (LCP)
			commonPrefix := longestCommonPrefix(completions)
			// If LCP is longer than current prefix, complete to LCP.
			if len(commonPrefix) > len(prefix) {
				t.lastPrefix = "" // Reset prefix state
				t.tabCount = 0    // Reset tab state
				return suggestions, len(prefix)
			}
		}
	} else {
		// --- Filename/directory completion (arguments after the command) ---
		var dir, filePrefix string
		if before, ok := strings.CutSuffix(prefix, "/"); ok {
			// Prefix ends with "/": user wants to list contents of that directory
			dir = before
			filePrefix = ""
		} else {
			// Split into the directory to search and the partial filename to match
			dir = filepath.Dir(prefix)
			filePrefix = filepath.Base(prefix)
			if filePrefix == "." {
				// filepath.Base returns "." for empty prefix — treat as match-all
				filePrefix = ""
			}
		}

		files, err := os.ReadDir(dir)
		if err != nil {
			fmt.Fprint(NewShell().originalStdout, "\a")
			return nil, len(prefix)
		}

		// Collect all entries whose name starts with filePrefix
		for _, file := range files {
			if strings.HasPrefix(file.Name(), filePrefix) {
				// Store the full relative path so we can trim the prefix correctly later
				fullPath := filepath.Join(dir, file.Name())
				fileCompletions = append(fileCompletions, fullPath)
			}
		}
		sort.Strings(fileCompletions)
		suggestions = make([][]rune, len(fileCompletions))
		switch len(fileCompletions) {
		case 0:
			// No file matches: ring the bell
			fmt.Fprint(NewShell().originalStdout, "\a")
			return nil, len(prefix)
		case 1:
			// Unique match: append "/" for directories, space for files
			info, err := os.Stat(fileCompletions[0])
			if err == nil && info.IsDir() {
				suggestions[0] = []rune(strings.TrimPrefix(fileCompletions[0], prefix))
				suggestions[0] = append(suggestions[0], '/')
				return suggestions, len(prefix)
			} else {
				suggestions[0] = []rune(strings.TrimPrefix(fileCompletions[0], prefix))
				suggestions[0] = append(suggestions[0], ' ')
				return suggestions, len(prefix)
			}
		default:
			commonPrefix := longestCommonPrefix(fileCompletions)
			if len(commonPrefix) > len(prefix) {
				// If the LCP is longer than the current prefix, complete to the LCP.
				t.lastPrefix = "" // Reset prefix state
				t.tabCount = 0    // Reset tab state
				suggestions = make([][]rune, len(fileCompletions))
				for i, comp := range fileCompletions {
					suggestions[i] = []rune(strings.TrimPrefix(comp, prefix))
				}
				return suggestions, len(prefix)
			}
		}
		// Multiple matches: build suffix suggestions for the multi-tab display
		for i, comp := range fileCompletions {
			suggestions[i] = []rune(strings.TrimPrefix(comp, prefix))
		}
		if len(suggestions) == 0 {
			fmt.Fprint(NewShell().originalStdout, "\a")
			return nil, len(prefix)
		}
	}

	// Track consecutive tab presses to decide whether to beep or show all options.
	if t.lastPrefix != prefix {
		// New prefix: reset the tab counter
		t.lastPrefix = prefix
		t.tabCount = 0
	}
	t.tabCount++

	if t.tabCount > 1 {
		displayNames := make([]string, len(fileCompletions))
		for i, path := range fileCompletions {
			info, err := os.Stat(path)
			if err == nil && info.IsDir() {
				displayNames[i] = path + "/"
			} else {
				displayNames[i] = path
			}
		}
		// Second (or more) tab: print all matching options
		var allCompletions []string
		if isFirstWord {
			allCompletions = completions
		} else {
			allCompletions = displayNames
		}
		fmt.Fprintf(NewShell().originalStdout, "\n%s\n", strings.Join(allCompletions, " "))
		fmt.Fprintf(NewShell().originalStdout, "$ %s", lineStr)
	} else {
		// First tab with multiple matches: just beep
		fmt.Fprint(NewShell().originalStdout, "\a")
	}
	return nil, len(prefix)
}
