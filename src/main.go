package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

const (
	LEXICON_SOURCE     = "http://www.gwicks.net/textlists/english3.zip"
	MIN_WORD_LENGTH    = 4 // words this length or longer are included
	MAX_UNIQUE_LETTERS = 7 // the maximum number of unique letters a word can include
)

func main() {

	defer TrackTime("main")()

	r, err := FetchLexicon(LEXICON_SOURCE)
	if err != nil {
		panic("Failed to fetch source: " + err.Error())
	}

	words, _ := MakeWordList(bufio.NewScanner(r), FilterWords)

	// words = words[2000:2020]
	wet := MakeWordExistenceTree(words)
	//tree.Print(wet)

	results := wet.Solve([]string{"b", "u", "l", "f", "r", "o", "g"}, "o")
	fmt.Printf("Found (%v) words\n", len(results))
}

func FetchLexicon(src string) (io.Reader, error) {
	defer TrackTime("FetchLexicon")()

	response, err := http.Get(src)
	defer response.Body.Close()
	if err != nil || response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Failed to GET %v, status: %v, error: %v", LEXICON_SOURCE, response.Status, err.Error())
	}
	responseBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to read response bytes: %v", err.Error())
	}

	zipReader, err := zip.NewReader(bytes.NewReader(responseBytes), response.ContentLength)
	if err != nil {
		return nil, fmt.Errorf("Failed to create zip reader: %v", err.Error())
	}

	unzippedReader, err := zipReader.File[0].Open()
	if err != nil {
		return nil, fmt.Errorf("Failed to create unzipped reader: %v", err.Error())
	}
	return unzippedReader, nil
}

func MakeWordList(src *bufio.Scanner, filter func(string) bool) (out []string, srcCount int) {
	defer TrackTime("MakeWordList")()
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

func TrackTime(name string) func() {
	start := time.Now()
	return func() {
		log.Printf("%s [%vms]\n", name, time.Since(start).Milliseconds())
	}
}
