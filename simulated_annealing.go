package main

import (
	"bytes"
	"context"
	"math"
	"math/rand"
)

func simulatedAnnealingCrack(
	ctx context.Context,
	globalData *GlobalData,
	pid int,
	startingKey []byte,
	separatorLetter byte,
	encryptedText string,
	initialTemp float64,
	floorTemp float64,
	coolingRate float64,
	triesPerEpoch int,
	triesBeforeStagnation int,
) ([]byte, float64) {
	currentKey := bytes.Clone(startingKey)
	currentScore := scoreTextFast(playfairDecrypt(encryptedText, currentKey), separatorLetter)

	bestKey := bytes.Clone(currentKey)
	bestScore := currentScore

	iterSinceBest := 0
	for curTemp := initialTemp; curTemp >= floorTemp; curTemp *= (1 - coolingRate) {
		for index := 0; index < triesPerEpoch; index++ {
			select {
			case <-ctx.Done():
				// exit early
				return bestKey, bestScore
			default:
			}

			// We have stagnated, check if we are at solution
			if (-3000 < bestScore && iterSinceBest > (triesBeforeStagnation/10)) || iterSinceBest > triesBeforeStagnation {
				return bestKey, bestScore
			}

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

			// New global best, report it
			if currentScore > bestScore {
				bestScore = currentScore
				bestKey = bytes.Clone(currentKey)
				iterSinceBest = 0
			} else {
				iterSinceBest++
			}
		}

		// Step annealing genetic algo
		newKeyData := geneticSimulatedAnnealingStep(globalData, pid, curTemp, currentKey, currentScore)
		currentKey, currentScore = newKeyData.key, newKeyData.score
	}

	return bestKey, bestScore
}

func geneticSimulatedAnnealingStep(
	globalData *GlobalData,
	pid int,
	temp float64,
	bestKey []byte,
	bestScore float64,
) KeyData {
	globalData.currentLock.Lock()
	defer globalData.currentLock.Unlock()

	globalData.currentKeys[pid].key = bestKey
	globalData.currentKeys[pid].score = bestScore

	// Generate energies
	pdf := make([]float64, len(globalData.currentKeys))
	sum := 0.0
	for i, keyData := range globalData.currentKeys {
		pdf[i] = math.Exp(keyData.score / temp) // e^E_i/T
		sum += pdf[i]
	}

	// Normalize and integrate
	for i, energy := range pdf {
		normalized := energy / sum

		if i == 0 {
			pdf[i] = normalized
		} else {
			pdf[i] = normalized + pdf[i-1]
		}
	}

	// Pick next key
	point := rand.Float64()
	for i, energy := range pdf {
		if point < energy {
			return globalData.currentKeys[i]
		}
	}

	// If float error, fallback on last
	return globalData.currentKeys[len(globalData.currentKeys)-1]
}
