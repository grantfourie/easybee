package wet

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"
	"sync"
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
	Children   map[string]*WordExistenceTreeNode `json:"Children"`
	IsHead     bool                              `json:"-"`
	IsDictWord bool                              `json:"-"`
	Word       string                            `json:"value"`
}

func (wet *WordExistenceTreeNode) Solve(letters []string, required string) []string {

	resultsChan := make(chan string)
	done := make(chan error)
	resultsSlice := make([]string, 0)
	wg := &sync.WaitGroup{}
	wg.Add(len(letters))

	for _, c := range letters {
		c := c
		go wet.solve(c, letters, required, resultsChan, wg)
	}

	go func() {
		wg.Wait()
		done <- errors.New("all done")
	}()

	fmt.Println("solvers started")

	for {
		select {
		case r := <-resultsChan:
			resultsSlice = append(resultsSlice, r)
		case err := <-done:
			fmt.Println("all Children finished: ", err.Error())
			return resultsSlice
		}
	}
}

func (wet *WordExistenceTreeNode) solve(start string, letters []string, required string, results chan<- string, wg *sync.WaitGroup) {
	currNode, found := wet.Children[start]
	if !found {
		wg.Done()
		return
	}

	for r, child := range currNode.Children {
		// is this path allowed ?
		if slices.Contains(letters, r) {

			// have we found an allowed word
			if child.IsDictWord && strings.Contains(child.Word, string(required)) {
				results <- child.Word
			}

			// continue down the tree
			wg.Add(len(letters))
			for _, l := range letters {
				go child.solve(l, letters, required, results, wg)
			}
		}
	}
	wg.Done()
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
