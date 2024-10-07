package main

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"os"
	"sync"
)

type NgramScorer struct {
	ngrams map[string]float64
	L      int
	N      int
	floor  float64
}

func NewNgramScorer(filename string, sep string) (*NgramScorer, error) {
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

	scorer := &NgramScorer{ngrams: ngrams, L: len(key)}
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

func (scorer *NgramScorer) score(text string) float64 {
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

type NgramScore struct {
	bigrams   *NgramScorer
	trigrams  *NgramScorer
	quadgrams *NgramScorer
}

var instance *NgramScore
var once sync.Once

func GetNgramScoreInstance() *NgramScore {
	once.Do(func() {
		instance = &NgramScore{}
		var err error
		instance.bigrams, err = NewNgramScorer("./resources/english_bigrams.txt", " ")
		if err != nil {
			log.Fatal(err)
		}
		instance.trigrams, err = NewNgramScorer("./resources/english_trigrams.txt", " ")
		if err != nil {
			log.Fatal(err)
		}
		instance.quadgrams, err = NewNgramScorer("./resources/english_quadgrams.txt", " ")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Loading ngram files")
	})
	return instance
}
