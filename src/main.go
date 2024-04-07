// I use a lexicon from http://www.gwicks.net/textlists/english3.zip
// This list of words does not match the official NYT Spelling Bee set, but is good enough.

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"syscall/js"
)

const (
	LEXICON_SOURCE     = "https://grantfourie.github.io/easybee/wordlist.txt"
	MIN_WORD_LENGTH    = 4 // words this length or longer are included
	MAX_UNIQUE_LETTERS = 7 // the maximum number of unique letters a word can include
)

func main() {

	r, err := FetchLexicon(LEXICON_SOURCE)
	if err != nil {
		fmt.Println("Failed to fetch source: " + err.Error())
	}
	words, _ := MakeWordList(bufio.NewScanner(r), FilterWords)
	wet := MakeWordExistenceTree(words)

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
