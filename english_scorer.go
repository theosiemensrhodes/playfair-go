package main

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"os"
	"strings"
	"sync"

	"github.com/agnivade/levenshtein"
	"github.com/antoineaugusti/wordsegmentation"
	"github.com/antoineaugusti/wordsegmentation/corpus"
	"github.com/theosiemensrhodes/go-bktree"
)

func NewBKTree(protoFile string, wordFile string) (*bktree.BKTree, error) {
	// Distance calculated by levenshtein distance
	tree := bktree.New(func(a, b []byte) int {
		return levenshtein.ComputeDistance(string(a), string(b))
	})

	// Best case, read tree from file
	err := tree.ReadFromFile(protoFile)
	if err != nil {
		file, err := os.Open(wordFile)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		var word string
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			_, err := fmt.Sscanf(scanner.Text(), "%s", &word)
			if err != nil {
				return nil, err
			}
			tree.Add([]byte(word))
		}

		// Try and save the created tree
		tree.SaveToFile(protoFile)
	}

	return tree, nil
}

type EnglishScorer struct {
	tree   *bktree.BKTree
	corpus corpus.EnglishCorpus
}

func (scorer *EnglishScorer) score(text string, power float64) float64 {
	// Remove playfair separator
	filteredText := removePlayfairSep(text, 'X')

	// Segment into words
	words := wordsegmentation.Segment(scorer.corpus, filteredText)

	// Calculate avg score
	var english_char_count float64 = 0
	var total_char_count float64 = 0
	for _, word := range words {
		byte_word := []byte(strings.ToLower(word))

		if len(scorer.tree.Find(byte_word, 0)) > 0 {
			english_char_count += math.Pow(float64(len(word)), power)
		} else if len(scorer.tree.Find(byte_word, 1)) > 0 {
			english_char_count += math.Pow(float64(len(word)*2/3), power)
		} else if len(scorer.tree.Find(byte_word, 2)) > 0 {
			english_char_count += math.Pow(float64(len(word)*1/3), power)
		}

		total_char_count += math.Pow(float64(len(word)), power)
	}

	// Calculate the percentage
	return (english_char_count / total_char_count)
}

var corpusInstance *corpus.EnglishCorpus
var corpusOnce sync.Once

func GetCorpusInstance() *corpus.EnglishCorpus {
	corpusOnce.Do(func() {
		corpus := corpus.NewEnglishCorpus()
		corpusInstance = &corpus
	})
	return corpusInstance
}

var bktreeInstance *bktree.BKTree
var bktreeOnce sync.Once

func GetDictionaryInstance() *bktree.BKTree {
	bktreeOnce.Do(func() {
		bktree, err := NewBKTree("./resources/bktree.bin", "./resources/words.txt")
		if err != nil {
			log.Fatal(err)
		}
		bktreeInstance = bktree
	})
	return bktreeInstance
}
