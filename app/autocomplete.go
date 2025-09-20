package main

import (
	"fmt"
	"strings"
)

type Node struct {
	Children    map[string]*Node
	IsEndOfWord bool
}

// Trie  is our actual tree that will hold all of our nodes, root node will be nil
type Trie struct {
	RootNode *Node
}

// / NewNode this will be used to initialize a new node with 26 children
// /each child should first be initialized to nil
func NewNode() *Node {
	node := &Node{
		Children: make(map[string]*Node),
	}
	return node
}

// NewTrie Creates a new trie with a root('constructor')
func NewTrie() *Trie {
	root := NewNode()
	return &Trie{RootNode: root}
}

// Insert inserts a word into the trie.
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

	for char, childNode := range node.Children {
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
	if len(suggestions) == 0 {
		fmt.Fprint(originalStdout, "\x07")
	}

	// If there is only one completion, it's standard shell behavior
	// to append a space to indicate the command is complete.
	if len(completions) == 1 {
		suggestions[0] = append(suggestions[0], ' ')
	}

	return suggestions, len(prefix)
}
