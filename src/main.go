// I use a lexicon from http://www.gwicks.net/textlists/english3.zip
// This list of words does not match the official NYT Spelling Bee set, but is good enough.

package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	LEXICON_SOURCE     = "https://grantfourie.github.io/easybee/wordlist.txt"
	MIN_WORD_LENGTH    = 4 // words this length or longer are included
	MAX_UNIQUE_LETTERS = 7 // the maximum number of unique letters a word can include
)

func main() {

	letters := flag.String("letters", "finalty", "the set of letters in the puzzle, center letter first")
	flag.Parse()

	r, err := FetchLexicon(LEXICON_SOURCE)
	if err != nil {
		panic("Failed to fetch source: " + err.Error())
	}

	words, _ := MakeWordList(bufio.NewScanner(r), FilterWords)
	wet := MakeWordExistenceTree(words)
	letterList := strings.Split(*letters, "")

	results := wet.Solve(letterList, letterList[0])
	fmt.Printf("Found (%v) words\n", len(results))
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
