package main

import (
	"fmt"
	"sort"
	"strings"
)

type Node struct {
	Children    map[string]*Node
	IsEndOfWord bool
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
	// To ensure alphabetical order, we must iterate over the children in a sorted manner.
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
func (t *Trie) Do(line []rune, pos int) (newLine [][]rune, length int) {
	lineStr := string(line[:pos])
	lastSpace := strings.LastIndex(lineStr, " ")
	prefix := lineStr[lastSpace+1:]
	completions := t.FindCompletions(prefix)
	suggestions := make([][]rune, len(completions))

	for i, comp := range completions {
		suggestions[i] = []rune(strings.TrimPrefix(comp, prefix))
	}

	switch len(completions) {
	case 0:
		// No matches: ring the bell.
		fmt.Fprint(NewShell().originalStdout, "\a")
		return nil, len(prefix) // Return length of prefix to clear it
	case 1:
		suggestions[0] = append(suggestions[0], ' ')
		return suggestions, len(prefix)
	default:
		// Multi-match: use longest common prefix
		commonPrefix := longestCommonPrefix(completions)
		// If LCP is longer than current prefix, complete to LCP.
		if len(commonPrefix) > len(prefix) {
			t.lastPrefix = "" // Reset prefix state
			t.tabCount = 0    // Reset tab state
			return suggestions, len(prefix)
		}

		// Prefix is already the LCP. Handle multi-tab case.
		if t.lastPrefix != prefix {
			t.lastPrefix = prefix
			t.tabCount = 0
		}
		t.tabCount++

		if t.tabCount > 1 {
			// On second (or more) tab, show all options.
			fmt.Fprintf(NewShell().originalStdout, "\n%s\n", strings.Join(completions, "  "))
			fmt.Fprintf(NewShell().originalStdout, "$ %s", lineStr)
		} else {
			// On first tab, just beep.
			fmt.Fprint(NewShell().originalStdout, "\a")
		}
		return nil, len(prefix)
	}
}
