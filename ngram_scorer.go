package main

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"os"
	"sync"
)

type NgramScore struct {
	ngrams map[string]float64
	L      int
	N      int
	floor  float64
}

func NewNgramScore(filename string, sep string) (*NgramScore, error) {
	ngrams := make(map[string]float64)
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var key string
	var count int
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		_, err := fmt.Sscanf(scanner.Text(), "%s %d", &key, &count)
		if err != nil {
			return nil, err
		}
		ngrams[key] = float64(count)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	scorer := &NgramScore{ngrams: ngrams, L: len(key)}
	scorer.N = 0
	for _, v := range ngrams {
		scorer.N += int(v)
	}
	for k, v := range ngrams {
		ngrams[k] = math.Log10(v / float64(scorer.N))
	}
	scorer.floor = math.Log10(0.01 / float64(scorer.N))

	return scorer, nil
}

func (scorer *NgramScore) score(text string) float64 {
	score := 0.0
	for i := 0; i <= len(text)-scorer.L; i++ {
		ngram := text[i : i+scorer.L]
		if val, ok := scorer.ngrams[ngram]; ok {
			score += val
		} else {
			score += scorer.floor
		}
	}
	return score
}

type NgramScorer struct {
	bigrams   *NgramScore
	trigrams  *NgramScore
	quadgrams *NgramScore
}

var ngramScoreInstance *NgramScorer
var ngramScoreOnce sync.Once

func GetNgramScorerInstance() *NgramScorer {
	ngramScoreOnce.Do(func() {
		ngramScoreInstance = &NgramScorer{}
		var err error
		ngramScoreInstance.bigrams, err = NewNgramScore("./resources/english_bigrams.txt", " ")
		if err != nil {
			log.Fatal(err)
		}
		ngramScoreInstance.trigrams, err = NewNgramScore("./resources/english_trigrams.txt", " ")
		if err != nil {
			log.Fatal(err)
		}
		ngramScoreInstance.quadgrams, err = NewNgramScore("./resources/english_quadgrams.txt", " ")
		if err != nil {
			log.Fatal(err)
		}
	})
	return ngramScoreInstance
}
