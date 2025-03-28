package score

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"playfaircrack/assets"
	"strings"
	"sync"

	"github.com/agnivade/levenshtein"
	"github.com/theosiemensrhodes/go-bktree"
	"github.com/theosiemensrhodes/wordsegmentation"
	"github.com/theosiemensrhodes/wordsegmentation/corpus"
)

func NewBKTree(protoFile string, words string) (*bktree.BKTree, error) {
	// Distance calculated by levenshtein distance
	tree := bktree.New(func(a, b []byte) int {
		return levenshtein.ComputeDistance(string(a), string(b))
	})

	// Best case, read tree from file
	err := tree.ReadFromFile(protoFile)
	if err != nil {
		reader := strings.NewReader(words)
		scanner := bufio.NewScanner(reader)

		var word string
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
	tree      *bktree.BKTree
	segmentor wordsegmentation.Segmentor
}

func (scorer *EnglishScorer) score(text []byte, power float64) float64 {
	// Remove playfair separator
	filteredText := RemovePlayfairSep(text, 'X')

	// Segment into words
	words := scorer.segmentor.Segment(filteredText)

	// Ignore last word, could be cut off
	words = words[:len(words)-2]

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

var segmentorInstance *wordsegmentation.Segmentor
var segmentorOnce sync.Once

func GetSegmentorInstance() *wordsegmentation.Segmentor {
	segmentorOnce.Do(func() {
		corpus := corpus.NewEnglishCorpus()
		segmentorInstance = wordsegmentation.NewSegmentor(corpus)
	})
	return segmentorInstance
}

var bktreeInstance *bktree.BKTree
var bktreeOnce sync.Once

func GetDictionaryInstance() *bktree.BKTree {
	bktreeOnce.Do(func() {
		bktree, err := NewBKTree("./assets/bktree.bin", assets.Dictionary)
		if err != nil {
			log.Fatal(err)
		}
		bktreeInstance = bktree
	})
	return bktreeInstance
}
