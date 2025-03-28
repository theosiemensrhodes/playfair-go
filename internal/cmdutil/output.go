package cmdutil

import (
	"fmt"

	"playfaircrack/internal/score"
)

func PrintSegmentedPlaintext(plaintext []byte, sep byte) {
	segmentor := score.GetSegmentorInstance()

	// Remove playfair separator
	filteredText := score.RemovePlayfairSep(plaintext, sep)

	// Segment into words and print
	words := segmentor.Segment(filteredText)
	for _, word := range words {
		fmt.Printf("%s ", word)
	}
}
