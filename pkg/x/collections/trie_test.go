// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

// +build !ignore

package collections

import (
	"strings"
	"testing"
)

func TestSensitiveSimple(t *testing.T) {
	var words = []string{
		"fuck you",
		"damn shit",
		"我日",
	}
	var trie = NewTrie()
	for _, text := range words {
		var filtered = trie.FilterText(text)
		if strings.Index(filtered, "*") < 0 {
			t.Fatalf("filter failed: %s, %s", text, filtered)
		}
	}
}

func TestSensitiveExample(t *testing.T) {
	var names []string
	var trie = NewTrie()
	trie.Insert("习*近*平")
	trie.Insert("毛*泽*东")
	trie.Insert("法*轮*功")

	matchAll := func(names []string) {
		for _, name := range names {
			begin, end := trie.MatchWords([]rune(name))
			if begin < 0 || end < begin {
				t.Fatalf("MatchWords failed")
			}
		}
	}

	matchNone := func(names []string) {
		for _, name := range names {
			begin, end := trie.MatchWords([]rune(name))
			if begin < 0 || end < begin {
				t.Fatalf("MatchWords failed, %d, %d, %s", begin, end, name)
			}
		}
	}

	names = []string{
		"近平",
		"习近",
		"习平",
		"习一一近平",
		"零零习近二二平",
		"习近二二平",
		"零零习一一近平",
		"习近二二平三三",
	}
	matchNone(names)

	names = []string{
		"习近平",
		"一一习近平",
		"习近平四四",
		"一一习近平四四",
		"一习近平",
		"习近平四",
		"习二近平",
		"习近三平",
		"一习二近平",
		"一习近三平",
		"一习近平四",
		"习二近三平",
		"习二近平四",
		"习近三平四",
		"一习二近三平",
		"一习近三平四",
		"一习二近平四",
		"习二近三平四",
		"一习二近三平四",
	}
	matchAll(names)

	names = []string{
		"习毛近平",
		"习毛近泽平",
		"习毛近泽东",
		"习毛近泽平东",
		"毛习泽近东",
	}
	matchAll(names)
}

func TestSensitiveFilter(t *testing.T) {
	var trie = NewTrie()
	trie.Insert("干死")
	var text = "干死AA干死BB"
	var filtered = trie.FilterText(text)
	var replaced = strings.Replace(text, "干死", "**", -1)
	if replaced != filtered {
		t.Fatalf("sensitive filter failure: expect: %s, got: %s", replaced, filtered)
	}
	t.Logf("filtered text: %s\n", filtered)
}
