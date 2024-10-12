package main

func scoreText(text string) float64 {
	scorer := GetNgramScorerInstance()

	// Filter playfair separator
	sepFiltered := removePlayfairSep(text, 'X')

	return scorer.bigrams.score(sepFiltered) + scorer.trigrams.score(sepFiltered) + scorer.quadgrams.score(sepFiltered)
}

func removePlayfairSep(text string, sep byte) string {
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
