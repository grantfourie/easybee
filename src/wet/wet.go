package wet

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"slices"
	"sort"
	"strings"
)

const (
	LEXICON_SOURCE     = "https://grantfourie.github.io/easybee/wordlist.txt"
	MIN_WORD_LENGTH    = 4 // words this length or longer are included
	MAX_UNIQUE_LETTERS = 7 // the maximum number of unique letters a word can include =
)

func MakeWordExistenceTree(words []string) *WordExistenceTreeNode {
	headNode := &WordExistenceTreeNode{IsHead: true, Children: make(map[string]*WordExistenceTreeNode, 26)}
	currNode := headNode
	for _, w := range words {
		currNode = headNode
		for i, c := range w {
			l := string(c)
			// if we are discovering a new tree node, populate it
			if currNode.Children[l] == nil {
				currNode.Children[l] = &WordExistenceTreeNode{Word: currNode.Word + l, Children: make(map[string]*WordExistenceTreeNode, 26)}
			}

			// if we are at the end of this word, indicate we have found a word and break
			if i == len(w)-1 {
				currNode.Children[l].IsDictWord = true
				break
			}

			// move to the next node
			currNode = currNode.Children[l]
		}
	}

	return headNode
}

type WordExistenceTreeNode struct {
	Children   map[string]*WordExistenceTreeNode
	IsHead     bool
	IsDictWord bool
	Word       string
}

func (wet *WordExistenceTreeNode) Solve(letters []string, required string) []string {
	results := make([]string, 0)

	if wet.IsDictWord && strings.Contains(wet.Word, required) {
		results = append(results, wet.Word)
	}

	for r, child := range wet.Children {
		if slices.Contains(letters, r) {

			results = append(results, child.Solve(letters, required)...)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i] < results[j]
	})
	return results

}

func FetchLexicon(src string) (io.Reader, error) {

	response, err := http.Get(src)
	defer response.Body.Close()
	if err != nil || response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Failed to GET %v, status: %v, error: %v", src, response.Status, err.Error())
	} else if responseBytes, err := io.ReadAll(response.Body); err != nil {
		return nil, fmt.Errorf("Failed to read response body when fetching lexicon: %v", err.Error())
	} else {
		return bytes.NewReader(responseBytes), nil
	}
}

func MakeWordList(src *bufio.Scanner, filter func(string) bool) (out []string, srcCount int) {
	for src.Scan() {
		srcCount++
		if !filter(src.Text()) {
			out = append(out, src.Text())
		}
	}

	return
}

func FilterWords(w string) bool {

	if len(w) < MIN_WORD_LENGTH {
		return true
	}

	uniqueLetters := make(map[rune]bool, MAX_UNIQUE_LETTERS)
	for _, char := range w {
		uniqueLetters[char] = true
		if len(uniqueLetters) > MAX_UNIQUE_LETTERS {
			return true
		}
	}

	return false
}
