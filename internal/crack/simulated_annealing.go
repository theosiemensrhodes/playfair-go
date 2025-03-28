package crack

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"os"
	"playfaircrack/internal/cipher"
	"playfaircrack/internal/score"
	"time"

	"golang.org/x/term"
)

func simulatedAnnealingCrack(
	ctx context.Context,
	logVerbose bool,
	poolData *poolData,
	pid int,
	startingKey [25]byte,
	initialTemp float64,
	floorTemp float64,
	coolingRate float64,
	triesPerEpoch int,
	triesBeforeStagnation int,
	geneticTempMultiplier float64,
) ([25]byte, float64) {
	ciphertext := poolData.global.ciphertext
	excludedLetter := poolData.global.excludedLetter
	separatorLetter := poolData.global.separatorLetter

	currentKey := startingKey
	currentScore := score.ScoreTextFast(cipher.PlayfairDecrypt(ciphertext, currentKey, excludedLetter), separatorLetter)

	bestKey := currentKey
	bestScore := currentScore

	iterSinceBest := 0
	iter := 0
	for curTemp := initialTemp; curTemp >= floorTemp; curTemp *= (1 - coolingRate) {
		// Check for other found solution or is solving
		select {
		case <-ctx.Done():
			// exit early
			return bestKey, bestScore
		default:
			// Wait if some thread is checking for solution
			poolData.global.chanLock.Lock()
			poolData.global.chanLock.Unlock()
		}

		for index := 0; index < triesPerEpoch; index++ {
			// We have stagnated, check if we are at solution
			if -3000 < bestScore && iterSinceBest > triesBeforeStagnation {
				return bestKey, bestScore
			}

			candidateKey := cipher.PermuteKey(currentKey, excludedLetter)
			candidatePlaintext := cipher.PlayfairDecrypt(ciphertext, candidateKey, excludedLetter)
			candidateScore := score.ScoreTextFast(candidatePlaintext, separatorLetter)

			// Calculate acceptance rate as function of current temperature
			delta := candidateScore - currentScore
			deltaRatio := delta / curTemp
			acceptanceRate := math.Exp(deltaRatio)

			if rand.Float64() < acceptanceRate {
				currentScore = candidateScore
				currentKey = candidateKey
			}

			// New global best, report it
			if currentScore > bestScore {
				bestScore = currentScore
				bestKey = currentKey
				iterSinceBest = 0
			} else {
				iterSinceBest++
			}

			iter++
		}

		// update global solution
		poolData.bestLock.Lock()
		if bestScore > poolData.bestScore {
			poolData.bestScore = bestScore
			poolData.bestKey = bestKey

			bestPlaintext := cipher.PlayfairDecrypt(ciphertext, bestKey, excludedLetter)

			if logVerbose {
				timestamp := time.Now().Format("15:04:05")
				width, _, err := term.GetSize(int(os.Stdout.Fd()))
				if err != nil {
					fmt.Printf("%d %s %s %f %s\n", iter, timestamp, bestKey, bestScore, bestPlaintext[:50])
				} else {
					fmt.Printf("%10d %2.4f %s %s %-4.4f %s\n", iter, curTemp, timestamp, bestKey, bestScore, bestPlaintext[:width-65])
				}
			}
		}
		poolData.bestLock.Unlock()

		// Step annealing genetic algo with prob e^-temp/max_temp
		// acceptanceRate := math.Exp(-curTemp / initialTemp)
		if rand.Float64() < 0.5 {
			newKeyData := geneticSimulatedAnnealingStep(poolData, pid, geneticTempMultiplier*curTemp, currentKey, currentScore)
			currentKey, currentScore = newKeyData.key, newKeyData.score
		}
	}

	return bestKey, bestScore
}

func geneticSimulatedAnnealingStep(
	poolData *poolData,
	pid int,
	temp float64,
	bestKey [25]byte,
	bestScore float64,
) keyData {
	poolData.currentLock.Lock()
	defer poolData.currentLock.Unlock()

	poolData.currentKeys[pid].key = bestKey
	poolData.currentKeys[pid].score = bestScore

	// Generate energies
	pdf := make([]float64, len(poolData.currentKeys))
	sum := 0.0
	for i, keyData := range poolData.currentKeys {
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
			return poolData.currentKeys[i]
		}
	}

	// If float error, fallback on last
	return poolData.currentKeys[len(poolData.currentKeys)-1]
}
