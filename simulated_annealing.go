package main

import (
	"fmt"
	"math"
	"math/rand"
	"time"
	"golang.org/x/term"
	"os"
)

func simulatedAnnealingCrack(encryptedText string, triesPerEpoch int, initialTemp float64, coolingRate float64, startingKey []byte) ([]byte, float64) {
	currentKey := copyByteArr(startingKey)
	currentScore := scoreText(playfairDecrypt(encryptedText, currentKey))

	bestKey := copyByteArr(currentKey)
	bestScore := currentScore
	iter := 0

	for curTemp := initialTemp; curTemp >= 1; curTemp *= (1 - coolingRate) {
		for index := 0; index < triesPerEpoch; index++ {
			candidateKey := copyByteArr(currentKey)
			permuteKey(candidateKey)
			candidatePlaintext := playfairDecrypt(encryptedText, candidateKey)
			candidateScore := scoreText(candidatePlaintext)

			delta := candidateScore - currentScore
			deltaRatio := delta / curTemp
			acceptanceRate := math.Exp(deltaRatio)

			if rand.Float64() < acceptanceRate {
				currentScore = candidateScore
				currentKey = copyByteArr(candidateKey)
			}

			if currentScore > bestScore {
				bestScore = currentScore
				bestKey = copyByteArr(currentKey)

				timestamp := time.Now().Format("15:04:05")
				width, _, err := term.GetSize(int(os.Stdout.Fd()))
				if err != nil {
					fmt.Printf("%d %s %s %f %f %s\n", iter, timestamp, bestKey, bestScore, curTemp, candidatePlaintext[:50])
				} else {
					fmt.Printf("%10d %s %s %-4.4f %4.4f %s\n", iter, timestamp, bestKey, bestScore, curTemp, candidatePlaintext[:width - 65])
				}
			}
			iter++
		}
	}

	return bestKey, bestScore
}

func copyByteArr(src []byte) []byte {
	dst := make([]byte, len(src))
	copy(dst, src)
	return dst
}
