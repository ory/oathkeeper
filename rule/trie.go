// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package rule

import (
	"net/url"
	"strings"

	"github.com/dlclark/regexp2"
)

// types for a trie root node
type TrieNode struct {
	children map[string]*TrieNode
	rules    []Rule
	// isWord   bool
}

// types for a trie node
type Trie struct {
	root        *TrieNode
	pathCleaner *regexp2.Regexp
}

// NewTrie creates a new trie
func NewTrie() *Trie {
	return &Trie{
		root: &TrieNode{
			children: make(map[string]*TrieNode),
		},
		// if the path contains a regex, we don't need to insert it or anything after into the trie
		pathCleaner: regexp2.MustCompile(`<.*>.*`, 0),
	}
}

// Insert a url host and paths into the trie
func (t *Trie) InsertRule(r Rule) error {
	node := t.root

	matchURL, err := url.Parse(r.Match.GetURL())
	if err != nil {
		return err
	}

	// insert the protocol into the trie
	if _, ok := node.children[string(r.Match.Protocol())]; !ok {
		node.children[string(r.Match.Protocol())] = &TrieNode{
			children: make(map[string]*TrieNode),
		}
	}
	node = node.children[string(r.Match.Protocol())]

	// insert the methods into the trie
	for _, method := range r.Match.GetMethods() {
		// reset the node to the root, followed by the protocol
		node = t.root
		node = node.children[string(r.Match.Protocol())]
		if _, ok := node.children[method]; !ok {
			node.children[method] = &TrieNode{
				children: make(map[string]*TrieNode),
			}
		}
		node = node.children[method]

		// insert the scheme into the trie
		if _, ok := node.children[matchURL.Scheme]; !ok {
			node.children[matchURL.Scheme] = &TrieNode{
				children: make(map[string]*TrieNode),
			}
		}
		node = node.children[matchURL.Scheme]

		// insert the host into the trie
		if _, ok := node.children[matchURL.Host]; !ok {
			node.children[matchURL.Host] = &TrieNode{
				children: make(map[string]*TrieNode),
			}
		}
		node = node.children[matchURL.Host]

		// remove any regex from the path
		cleanPath, err := t.pathCleaner.Replace(matchURL.Path, "", 0, -1)
		if err != nil {
			return err
		}

		// remove the leading and trailing slash
		trimmedPath := strings.Trim(cleanPath, "/")

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
	return nil
}

// return the longest prefix of the url that is in the trie
func (t *Trie) LongestPrefix(u *url.URL) string {
	node := t.root
	var prefix string
	// check the scheme
	if _, ok := node.children[u.Scheme]; !ok {
		return prefix
	}
	prefix += u.Scheme
	node = node.children[u.Scheme]

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
func (t *Trie) Match(method string, u *url.URL, protocol Protocol) []Rule {
	node := t.root

	// check the protocol
	if _, ok := node.children[string(protocol)]; !ok {
		return nil
	}
	node = node.children[string(protocol)]

	// check the method
	if _, ok := node.children[method]; !ok {
		return nil
	}
	node = node.children[method]

	// check the scheme
	if _, ok := node.children[u.Scheme]; !ok {
		return nil
	}
	node = node.children[u.Scheme]

	// check the host
	if _, ok := node.children[u.Host]; !ok {
		return nil
	}
	node = node.children[u.Host]
	// remove the leading and trailing slash
	trimmedPath := strings.Trim(u.Path, "/")
	if len(trimmedPath) == 0 {
		return node.rules
	} else {
		// check the paths
		splitPaths := strings.Split(trimmedPath, "/")
		for _, path := range splitPaths {
			if _, ok := node.children[string(path)]; !ok {
				break
			}
			node = node.children[string(path)]
		}
		return node.rules
	}
}
