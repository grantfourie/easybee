package main

import (
	"bufio"
	"fmt"
	"github.com/grantfourie/easybee/src/wet"
)

func main() {
	r, err := wet.FetchLexicon(wet.LEXICON_SOURCE)
	if err != nil {
		fmt.Println("Failed to fetch source: " + err.Error())
	}
	words, _ := wet.MakeWordList(bufio.NewScanner(r), wet.FilterWords)
	wet := wet.MakeWordExistenceTree(words)
	fmt.Println(wet.Solve([]string{"r", "u", "c", "a", "n", "t", "i"}, "r"))
	return
}
