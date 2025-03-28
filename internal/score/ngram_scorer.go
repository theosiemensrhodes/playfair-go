package score

import (
	"bufio"
	"log"
	"math"
	"playfaircrack/assets"
	"strconv"
	"strings"
	"sync"
)

type NgramScore struct {
	ngrams []float64
	L      int
	N      int
	floor  float64
}

func NewNgramScore(data string) (*NgramScore, error) {
	reader := strings.NewReader(data)
	scanner := bufio.NewScanner(reader)

	var ngramMap = make(map[string]float64)
	var totalCount int
	var ngramLength int
	var firstLine bool = true

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		key := fields[0]
		count, err := strconv.Atoi(fields[1])
		if err != nil {
			return nil, err
		}
		if firstLine {
			ngramLength = len(key)
			firstLine = false
		}
		ngramMap[key] = float64(count)
		totalCount += count
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if ngramLength < 2 || ngramLength > 4 {
		panic("ngram length is not 2, 3, or 4")
	}

	// Precompute size (26^L)
	var size int
	switch ngramLength {
	case 2:
		size = 26 * 26
	case 3:
		size = 26 * 26 * 26
	case 4:
		size = 26 * 26 * 26 * 26
	}

	scorer := &NgramScore{
		L: ngramLength,
		N: totalCount,
	}
	scorer.floor = math.Log10(0.01 / float64(scorer.N))

	scorer.ngrams = make([]float64, size)
	// Set all to floor
	for i := range scorer.ngrams {
		scorer.ngrams[i] = scorer.floor
	}

	// Fill known ngrams
	totalFloat := float64(scorer.N)
	for k, v := range ngramMap {
		logProb := math.Log10(v / totalFloat)
		idx := ngramToIndex(k, scorer.L)
		scorer.ngrams[idx] = logProb
	}

	return scorer, nil
}

func ngramToIndex(ngram string, L int) int {
	idx := 0
	for i := 0; i < L; i++ {
		idx = idx*26 + int(ngram[i]-'A')
	}
	return idx
}

func (scorer *NgramScore) score(text string) float64 {
	score := 0.0
	end := len(text) - scorer.L

	for i := 0; i <= end; i++ {
		var idx int
		switch scorer.L {
		case 2:
			idx = 26*int(text[i]-'A') + int(text[i+1]-'A')
		case 3:
			idx = 26*(26*int(text[i]-'A')+int(text[i+1]-'A')) + int(text[i+2]-'A')
		case 4:
			idx = 26*(26*(26*int(text[i]-'A')+int(text[i+1]-'A'))+int(text[i+2]-'A')) + int(text[i+3]-'A')
		default:
			panic("ngram length is not 2, 3, or 4")
		}
		score += scorer.ngrams[idx]
	}
	return score
}

type NgramScorer struct {
	bigrams   *NgramScore
	trigrams  *NgramScore
	quadgrams *NgramScore
}

var (
	ngramScoreInstance *NgramScorer
	ngramScoreOnce     sync.Once
)

func GetNgramScorerInstance() *NgramScorer {
	ngramScoreOnce.Do(func() {
		ngramScoreInstance = &NgramScorer{}

		// load bigrams, trigrams, and quadgrams in parallel
		var wg sync.WaitGroup
		wg.Add(3)

		go func() {
			defer wg.Done()
			var err error
			ngramScoreInstance.bigrams, err = NewNgramScore(assets.Bigrams)
			if err != nil {
				log.Fatal(err)
			}
		}()

		go func() {
			defer wg.Done()
			var err error
			ngramScoreInstance.trigrams, err = NewNgramScore(assets.Trigrams)
			if err != nil {
				log.Fatal(err)
			}
		}()

		go func() {
			defer wg.Done()
			var err error
			ngramScoreInstance.quadgrams, err = NewNgramScore(assets.Quadgrams)
			if err != nil {
				log.Fatal(err)
			}
		}()

		wg.Wait()
	})
	return ngramScoreInstance
}
