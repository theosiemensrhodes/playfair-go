package main

import (
	"bytes"
	"fmt"
	"math"
	"math/rand"
	"os"
	"time"

	"golang.org/x/term"
)

const (
	THRESHHOLD_ENGLISH float64 = 0.9
)

type SolutionData struct {
	percentEnglish float64
	plaintext      string
	segmentedText  []string
}

func simulatedAnnealingCrack(
	encryptedText string,
	separatorLetter byte,
	triesPerEpoch int,
	initialTemp float64,
	floorTemp float64,
	coolingRate float64,
	triesBeforeStagnation int,
	startingKey []byte,
) ([]byte, float64, *SolutionData) {
	currentKey := bytes.Clone(startingKey)
	currentScore := scoreTextFast(playfairDecrypt(encryptedText, currentKey), separatorLetter)

	bestKey := bytes.Clone(currentKey)
	bestPlaintext := playfairDecrypt(encryptedText, bestKey)
	bestScore := currentScore
	iter := 0
	iterSinceBest := 0

	for curTemp := initialTemp; curTemp >= floorTemp; curTemp *= (1 - coolingRate) {
		for index := 0; index < triesPerEpoch; index++ {
			// We have stagnated, check if we are at solution
			if iterSinceBest > triesBeforeStagnation {
				solution := &SolutionData{}
				solution.plaintext = bestPlaintext
				solution.percentEnglish, solution.segmentedText = scoreTextSlow(bestPlaintext, separatorLetter, 1.5)
				if solution.percentEnglish > THRESHHOLD_ENGLISH {
					if logVerbose {
						fmt.Printf("\n")
					}

					return bestKey, bestScore, solution
				}

				iterSinceBest = 0
			}

			// We are not at solution, keep checking
			candidateKey := bytes.Clone(currentKey)
			permuteKey(candidateKey)
			candidatePlaintext := playfairDecrypt(encryptedText, candidateKey)
			candidateScore := scoreTextFast(candidatePlaintext, separatorLetter)

			// Calculate acceptance rate as function of current temperature
			delta := candidateScore - currentScore
			deltaRatio := delta / curTemp
			acceptanceRate := math.Exp(deltaRatio)

			if rand.Float64() < acceptanceRate {
				currentScore = candidateScore
				currentKey = bytes.Clone(candidateKey)
			}

			// New global minimum, report it
			if currentScore > bestScore {
				bestScore = currentScore
				bestKey = bytes.Clone(currentKey)
				bestPlaintext = candidatePlaintext

				timestamp := time.Now().Format("15:04:05")
				width, _, err := term.GetSize(int(os.Stdout.Fd()))
				if logVerbose {
					if err != nil {
						fmt.Printf("%d %s %s %f %f %s\n", iter, timestamp, bestKey, bestScore, curTemp, candidatePlaintext[:50])
					} else {
						fmt.Printf("%10d %s %s %-4.4f %-4.4f %s\n", iter, timestamp, bestKey, bestScore, curTemp, candidatePlaintext[:width-65])
					}
				}
				iterSinceBest = 0
			}

			iterSinceBest++
			iter++
		}
	}

	return bestKey, bestScore, nil
}
