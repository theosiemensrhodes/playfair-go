package main

import (
	"bytes"
	"fmt"
	"math"
	"math/rand"
	"time"
)

func playfairCrack(ciphertext string, excludedLetter byte, separatorLetter byte) error {
	// Get a random starting key
	var bestKey []byte
	letters := []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	for _, l := range letters {
		if l != excludedLetter {
			bestKey = append(bestKey, l)
		}
	}
	rand.Shuffle(len(bestKey), func(i, j int) { bestKey[i], bestKey[j] = bestKey[j], bestKey[i] })

	bestScore := math.Inf(-1)
	i := 0

	fmt.Printf("Cipher Text:\n%s\n\n", ciphertext)

	startTime := time.Now()
	for {
		i++
		key, score, solution := simulatedAnnealingCrack(
			ciphertext,
			separatorLetter,
			1024,
			50.0,
			1.0,
			0.01,
			80000,
			bestKey,
		)

		if solution != nil {
			elapsedTime := time.Now().Sub(startTime)

			fmt.Printf("Solution found, in %v!\n", elapsedTime)
			fmt.Printf("Key: %s\n", key)
			fmt.Printf("English Word Score %2.2f%%\n\n", 100.0*solution.percentEnglish)

			fmt.Printf("Raw Plaintext:\n%s\n\n", solution.plaintext)

			fmt.Printf("Segmented Plaintext:\n")
			for _, word := range solution.segmentedText {
				fmt.Printf("%s ", word)
			}

			return nil
		}

		if score > bestScore {
			bestScore = score
			bestKey = bytes.Clone(key)
		}
	}
}
