package main

import (
	"fmt"
	"github.com/shivamMg/ppds/tree"
	"slices"
	"sort"
	"strings"
	"sync/atomic"
)

func MakeWordExistenceTree(words []string) *WordExistenceTreeNode {
	defer TrackTime("MakeWordExistenceTree")()
	headNode := &WordExistenceTreeNode{IsHead: true, children: make(map[string]*WordExistenceTreeNode, 26)}
	currNode := headNode
	for _, w := range words {
		currNode = headNode
		for i, c := range w {
			l := string(c)
			// if we are discovering a new tree node, populate it
			if currNode.children[l] == nil {
				currNode.children[l] = &WordExistenceTreeNode{Word: currNode.Word + l, children: make(map[string]*WordExistenceTreeNode, 26)}
			}

			// if we are at the end of this word, indicate we have found a word and break
			if i == len(w)-1 {
				currNode.children[l].IsDictWord = true
				break
			}

			// move to the next node
			currNode = currNode.children[l]
		}
	}

	return headNode
}

type WordExistenceTreeNode struct {
	children   map[string]*WordExistenceTreeNode
	IsHead     bool
	IsDictWord bool
	Word       string
}

func (wet *WordExistenceTreeNode) Data() interface{} {
	if wet.IsHead {
		return "*"
	}
	return string(wet.Word[len(wet.Word)-1])
}

func (wet *WordExistenceTreeNode) Children() []tree.Node {
	childSlice := make([]tree.Node, 0, len(wet.children))
	for _, c := range wet.children {
		childSlice = append(childSlice, c)
	}

	sort.Slice(childSlice, func(i, j int) bool {
		return childSlice[i].Data().(string) < childSlice[j].Data().(string)
	})

	return childSlice
}

func (wet *WordExistenceTreeNode) Solve(letters []string, required string) []string {

	defer TrackTime("Solve")()

	resultsChan := make(chan string)
	resultsSlice := make([]string, 0)
	workerCounter := int64(len(letters))

	for _, c := range letters {
		go wet.solve(c, letters, required, &workerCounter, resultsChan)
	}

	for atomic.LoadInt64(&workerCounter) != 0 {
		select {
		case r := <-resultsChan:
			resultsSlice = append(resultsSlice, r)
			fmt.Printf("Found Word: [%v]\n", strings.ToUpper(r))
		default:
			continue
		}
	}
	return resultsSlice
}

func (wet *WordExistenceTreeNode) solve(start string, letters []string, required string, workerCounter *int64, results chan<- string) {
	currNode, found := wet.children[start]
	if !found {
		atomic.AddInt64(workerCounter, -1)
		return
	}

	for r, child := range currNode.children {
		// is this path allowed ?
		if slices.Contains(letters, r) {

			// have we found an allowed word
			if child.IsDictWord && strings.Contains(child.Word, string(required)) {
				results <- child.Word
			}

			// continue down the tree
			for _, l := range letters {
				atomic.AddInt64(workerCounter, 1)
				go child.solve(l, letters, required, workerCounter, results)
			}
		}
	}
	atomic.AddInt64(workerCounter, -1)
}
