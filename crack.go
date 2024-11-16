package main

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"os"
	"sync"
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

type GlobalData struct {
	bestScore    float64
	bestKey      []byte
	bestLock     sync.Mutex
	currentKeys  []KeyData
	currentLock  sync.Mutex
	solutionChan chan SolutionData
	chanLock     sync.Mutex
}

type KeyData struct {
	score float64
	key   []byte
}

func playfairCrack(ciphertext string, excludedLetter byte, separatorLetter byte) error {
	poolSize := 8

	// Signal we are starting
	fmt.Printf("Cracking Cipher Text:\n%s\n\n", ciphertext)
	startTime := time.Now()

	globalData := &GlobalData{
		bestScore:    math.Inf(-1),
		bestLock:     sync.Mutex{},
		currentKeys:  make([]KeyData, poolSize),
		currentLock:  sync.Mutex{},
		solutionChan: make(chan SolutionData),
		chanLock:     sync.Mutex{},
	}

	// Get a random starting keys
	for i := 0; i < poolSize; i++ {
		globalData.currentKeys[i].key = generateRandomKey(excludedLetter)
		plaintext := playfairDecrypt(ciphertext, globalData.currentKeys[i].key)
		globalData.currentKeys[i].score = scoreTextFast(plaintext, excludedLetter)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var waitGroup sync.WaitGroup
	waitGroup.Add(poolSize)

	for i, keyData := range globalData.currentKeys {
		go processWorker(
			ctx,
			&waitGroup,
			globalData,
			i,
			keyData.key,
			separatorLetter,
			ciphertext,
		)
	}

	var solution SolutionData
	select {
	case solution = <-globalData.solutionChan:
		// A worker found the key, cancel other workers
		cancel()
	case <-ctx.Done():
		// Context canceled externally
	}

	// Wait for all goroutines to finish
	waitGroup.Wait()

	// Log result
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

func processWorker(
	ctx context.Context,
	waitGroup *sync.WaitGroup,
	globalData *GlobalData,
	pid int,
	initialKey []byte,
	separatorLetter byte,
	ciphertext string,
) {
	defer waitGroup.Done()

	// initialize search
	localBest := globalData.bestScore
	localBestKey := initialKey

	key := initialKey
	temperature := 50.0
	epoch := 0

	for {
		epoch++
		key, score := simulatedAnnealingCrack(
			ctx,
			globalData,
			pid,
			key,
			separatorLetter,
			ciphertext,
			temperature,
			1.0,
			0.01,
			256,
			80000,
		)

		// Check for other found solution
		select {
		case <-ctx.Done():
			// exit early
			return
		default:
		}

		// compare to local best
		if score > localBest {
			localBest = score
			localBestKey = bytes.Clone(key)
		}

		// update global solution
		globalData.bestLock.Lock()
		if localBest > globalData.bestScore {
			globalData.bestScore = localBest
			globalData.bestKey = bytes.Clone(localBestKey)

			bestPlaintext := playfairDecrypt(ciphertext, globalData.bestKey)

			if logVerbose {
				timestamp := time.Now().Format("15:04:05")
				width, _, err := term.GetSize(int(os.Stdout.Fd()))
				if err != nil {
					fmt.Printf("%d %s %s %f %s\n", epoch, timestamp, localBestKey, localBest, bestPlaintext[:50])
				} else {
					fmt.Printf("%10d %s %s %-4.4f %s\n", epoch, timestamp, localBestKey, localBest, bestPlaintext[:width-57])
				}
			}
		}
		globalData.bestLock.Unlock()

		// Check for solution
		if -3000 < localBest {
			globalData.chanLock.Lock()
			defer globalData.chanLock.Unlock()
			solution := SolutionData{}
			solution.plaintext = playfairDecrypt(ciphertext, localBestKey)
			solution.percentEnglish, solution.segmentedText = scoreTextSlow(solution.plaintext, separatorLetter, 1.5)

			select {
			case <-ctx.Done():
				// Context canceled, do not attempt to send
				return
			case globalData.solutionChan <- solution:
				// Successfully sent the solution
				return
			}
		}

		// Random steps
		// global_gap := (globalData.bestScore + 5000) / 2000
		// if rand.Float64() > global_gap {
		// 	key = generateRandomKey('X')
		// 	fmt.Printf("Restarting with random key at %f: %s\n", global_gap, key)
		// } else {
		// 	gap := (globalData.bestScore - localBest) / globalData.bestScore
		// 	if rand.Float64() < gap {
		// 		key = globalData.bestKey
		// 		fmt.Printf("Restarting with global best key at %f: %s\n", gap, key)
		// 	} else {
		// 		key = localBestKey
		// 		fmt.Printf("Restarting with own best key at %f: %s\n", gap, key)
		// 	}
		// 	// since we're close, start at a lower temperature
		// 	temperature = 25.0
		// }
	}
}
