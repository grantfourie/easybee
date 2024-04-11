package main

import (
	"bufio"
	"fmt"
	"github.com/grantfourie/easybee/src/wet"
	"strings"
	"syscall/js"
)

func main() {
	r, err := wet.FetchLexicon(wet.LEXICON_SOURCE)
	if err != nil {
		fmt.Println("Failed to fetch source: " + err.Error())
	}
	words, _ := wet.MakeWordList(bufio.NewScanner(r), wet.FilterWords)
	wet := wet.MakeWordExistenceTree(words)

	doc := js.Global().Get("document")
	solveBtn := doc.Call("getElementById", "solveBtn")
	resultList := doc.Call("getElementById", "resultList")

	solveBtn.Call("addEventListener", "click", js.FuncOf(
		func(this js.Value, args []js.Value) any {
			letters, requiredLetter := getLetters()
			results := wet.Solve(letters, requiredLetter)
			resultList.Set("innerHTML", "")
			for _, word := range results {
				listItem := doc.Call("createElement", "li")
				listItem.Set("textContent", word)
				resultList.Call("appendChild", listItem)
			}

			return nil
		},
	))

	select {}
}

func getLetters() ([]string, string) {
	letters := make([]string, 0, 7)
	hexagons := js.Global().Get("document").Call("getElementsByClassName", "hexagon")
	for i := 0; i < hexagons.Length(); i++ {
		hexagon := hexagons.Index(i)
		char := strings.ToLower(hexagon.Get("value").String())
		letters = append(letters, char)
	}

	middleHexagon := js.Global().Get("document").Call("getElementById", "center-letter")
	requiredLetter := strings.ToLower(middleHexagon.Get("value").String())

	fmt.Printf("input letters %v\n", letters)
	return letters, requiredLetter
}
