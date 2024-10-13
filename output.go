package main

import (
	"fmt"

	"github.com/antoineaugusti/wordsegmentation"
)

func printKey(key string) {
	// Print key in lines
	for i := 0; i < 25; i += 5 {
		fmt.Printf("%c %c %c %c %c\n", key[i+0], key[i+1], key[i+2], key[i+3], key[i+4])
	}
}

func printSegmentedPlaintext(plaintext string, sep byte) {
	corpus := GetCorpusInstance()

	// Remove playfair separator
	filteredText := removePlayfairSep(plaintext, sep)

	// Segment into words and print
	words := wordsegmentation.Segment(corpus, filteredText)
	for _, word := range words {
		fmt.Printf("%s ", word)
	}
}
