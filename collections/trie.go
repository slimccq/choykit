// Copyright © 2020 ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package collections

import (
	"bytes"
	"strings"
	"unicode"
)

const WildCardChar = '\u002A' // 通配符'*'

type TrieNode struct {
	children map[rune]*TrieNode
	isEnd    bool
}

func NewTrieNode() *TrieNode {
	return &TrieNode{
		children: make(map[rune]*TrieNode),
		isEnd:    false,
	}
}

type Trie struct {
	root *TrieNode
	size int
}

func NewTrie() *Trie {
	return &Trie{
		root: NewTrieNode(),
		size: 0,
	}
}

func (t *Trie) Init() error {
	t.root = NewTrieNode()
	t.size = 0
	return nil
}

func (t *Trie) WordsCount() int {
	return t.size
}

//插入需要过滤的敏感词
func (t *Trie) Insert(word string) {
	var node = t.root
	for _, r := range word {
		var ch = unicode.ToLower(r)
		if _, found := node.children[ch]; !found {
			node.children[ch] = NewTrieNode()
		}
		node = node.children[ch]
	}
	node.isEnd = true
	t.size++
}

//从指定位置开始匹配
func (t *Trie) MatchAt(words []rune, start int) int {
	node := t.root
	for start >= 0 && start < len(words) {
		var word = unicode.ToLower(words[start])
		child, found := node.children[word]
		if !found {
			child, found = node.children[WildCardChar] //匹配通配符
			if found {
				if child.isEnd {
					return start
				}
				node = child
				if child, found := node.children[word]; found {
					if child.isEnd {
						return start
					}
					node = child
				}
				start++
				continue
			}
			return -1
		}
		if child.isEnd {
			return start
		}
		node = child
		start++
	}
	if node.isEnd {
		return start
	}
	return -1
}

//匹配一串文字
func (t *Trie) MatchWords(words []rune) (int, int) {
	for i := 0; i < len(words); i++ {
		index := t.MatchAt(words, i)
		if index >= 0 {
			return i, index + 1
		}
	}
	return -1, -1
}

//是否有敏感词
func (t *Trie) MatchString(text string) bool {
	var words = []rune(text)
	start, end := t.MatchWords(words)
	return start < 0 && end < 0
}

//将敏感字符替换为*
func (t *Trie) FilterText(text string) string {
	var words = []rune(text)
	var start, end = t.MatchWords(words)
	if start < 0 && end < 0 { // common case
		return text
	}
	var buf bytes.Buffer
	for end > 0 {
		buf.WriteString(string(words[:start]))
		buf.WriteString(strings.Repeat("*", end-start))
		words = words[end:]
		start, end = t.MatchWords(words)
	}
	buf.WriteString(string(words))
	return buf.String()
}
