package main

func scoreText(text string) float64 {
	scorer := GetNgramScoreInstance()
	return scorer.bigrams.score(text) + scorer.trigrams.score(text) + scorer.quadgrams.score(text)
}
