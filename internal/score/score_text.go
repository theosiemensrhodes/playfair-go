package score

import (
	"math"
	"strings"
)

func ScoreTextFast(text []byte, sep byte) float64 {
	scorer := GetNgramScorerInstance()

	// Filter playfair separator
	sepFiltered := RemovePlayfairSep(text, sep)

	return scorer.bigrams.score(sepFiltered) + scorer.trigrams.score(sepFiltered) + scorer.quadgrams.score(sepFiltered)
}

func ScoreTextSlow(text []byte, sep byte, power float64) (float64, []string) {
	segmentor := GetSegmentorInstance()
	dict := GetDictionaryInstance()

	// Remove playfair separator
	filteredText := RemovePlayfairSep(text, sep)

	// Segment into words
	words := segmentor.Segment(filteredText)

	// Calculate avg score
	var english_char_count float64 = 0
	var total_char_count float64 = 0
	for _, word := range words {
		byte_word := []byte(strings.ToLower(word))

		if len(dict.Find(byte_word, 0)) > 0 {
			english_char_count += math.Pow(float64(len(word)), power)
		} else if len(dict.Find(byte_word, 1)) > 0 {
			english_char_count += math.Pow(float64(len(word)*2/3), power)
		} else if len(dict.Find(byte_word, 2)) > 0 {
			english_char_count += math.Pow(float64(len(word)*1/3), power)
		}

		total_char_count += math.Pow(float64(len(word)), power)
	}

	// Calculate the percentage
	return (english_char_count / total_char_count), words
}

func RemovePlayfairSep(text []byte, sep byte) string {
	// Filter playfair X pattern
	filtered := []byte{text[0]}
	for i := 1; i < len(text)-1; i++ {
		if text[i] == sep && text[i-1] == text[i+1] {
			continue
		}
		filtered = append(filtered, text[i])
	}

	// Append the last character
	filtered = append(filtered, text[len(text)-1])
	return string(filtered)
}
