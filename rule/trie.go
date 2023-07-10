package rule

import (
	"fmt"
	"net/url"
	"strings"
)

// types for a trie root node
type TrieNode struct {
	children map[string]*TrieNode
	rules    []*Rule
	// isWord   bool
}

// types for a trie node
type Trie struct {
	root *TrieNode
}

// NewTrie creates a new trie
func NewTrie() *Trie {
	return &Trie{
		root: &TrieNode{
			children: make(map[string]*TrieNode),
		},
	}
}

// Insert inserts a word into the trie
func (t *Trie) InsertString(word string) {
	node := t.root
	if _, ok := node.children[word]; !ok {
		node.children[word] = &TrieNode{
			children: make(map[string]*TrieNode),
		}
	}
	node = node.children[word]
}

// Insert a url host and paths into the trie along with the rule
func (t *Trie) InsertURL(u *url.URL, r *Rule) {
	// TODO: should this also handle scheme?

	node := t.root
	// insert the host into the trie
	if _, ok := node.children[u.Host]; !ok {
		node.children[u.Host] = &TrieNode{
			children: make(map[string]*TrieNode),
		}
	}
	node = node.children[u.Host]

	// remove the leading and trailing slash
	trimmedPath := strings.Trim(u.Path, "/")
	if len(trimmedPath) == 0 {
		node.rules = append(node.rules, r)
	} else {
		// insert the paths into the trie
		splitPaths := strings.Split(trimmedPath, "/")
		i := 0
		for _, path := range splitPaths {
			i++
			if _, ok := node.children[string(path)]; !ok {
				node.children[string(path)] = &TrieNode{
					children: make(map[string]*TrieNode),
				}
			}
			node = node.children[string(path)]
			if i == len(splitPaths) {
				node.rules = append(node.rules, r)
			}
		}
	}
}

// Insert a url host and paths into the trie
func (t *Trie) InsertRule(r *Rule) {
	node := t.root

	// TODO: handle error properly
	matchURL, err := url.Parse(r.Match.GetURL())
	if err != nil {
		fmt.Println("error parsing url")
	}

	// TODO: should this also handle scheme?

	// insert the host into the trie
	if _, ok := node.children[matchURL.Host]; !ok {
		node.children[matchURL.Host] = &TrieNode{
			children: make(map[string]*TrieNode),
		}
	}
	node = node.children[matchURL.Host]
	// remove the leading and trailing slash
	trimmedPath := strings.Trim(matchURL.Path, "/")

	if len(trimmedPath) == 0 {
		node.rules = append(node.rules, r)
	} else {

		// insert the paths into the trie
		splitPaths := strings.Split(trimmedPath, "/")
		i := 0
		for _, path := range splitPaths {
			i++
			if _, ok := node.children[string(path)]; !ok {
				node.children[string(path)] = &TrieNode{
					children: make(map[string]*TrieNode),
				}
			}
			node = node.children[string(path)]
			// if this is the last path, append the rule
			if i == len(splitPaths) {
				node.rules = append(node.rules, r)
			}
		}
	}
}

// return the longest prefix of the url that is in the trie
func (t *Trie) LongestPrefix(u *url.URL) string {
	node := t.root
	var prefix string
	// check the host
	if _, ok := node.children[u.Host]; !ok {
		return prefix
	}
	prefix += u.Host
	node = node.children[u.Host]
	// check the paths
	// remove the leading and trailing slash
	trimmedPath := strings.Trim(u.Path, "/")

	if len(trimmedPath) > 0 {
		splitPaths := strings.Split(trimmedPath, "/")
		for _, path := range splitPaths {
			if _, ok := node.children[string(path)]; !ok {
				break
			}
			prefix += "/" + string(path)
			node = node.children[string(path)]
		}
	}
	return prefix

}

// return the rules of the longest prefix of the url that is in the trie
func (t *Trie) Match(u *url.URL) []*Rule {
	node := t.root
	var rules []*Rule
	// check the host
	if _, ok := node.children[u.Host]; !ok {
		return rules
	}
	node = node.children[u.Host]
	// remove the leading and trailing slash
	trimmedPath := strings.Trim(u.Path, "/")
	if len(trimmedPath) == 0 {
		rules = node.rules
		return rules
	} else {
		// check the paths
		splitPaths := strings.Split(trimmedPath, "/")
		for _, path := range splitPaths {
			if _, ok := node.children[string(path)]; !ok {
				break
			}
			node = node.children[string(path)]
		}
		rules = node.rules
		return rules
	}
}
