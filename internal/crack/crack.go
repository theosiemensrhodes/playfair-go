package crack

import (
	"context"
	"fmt"
	"math"
	"playfaircrack/internal/cipher"
	"playfaircrack/internal/score"
	"runtime"
	"strings"
	"sync"
	"time"
)

const (
	THRESHHOLD_ENGLISH float64 = 0.9
)

type CrackResult struct {
	PercentEnglish float64
	Plaintext      string
	SegmentedText  []string
	Key            string
	ElapsedTime    time.Duration
}

type globalData struct {
	solutionChan    chan CrackResult
	chanLock        sync.Mutex
	cancel          context.CancelFunc
	ciphertext      []byte
	excludedLetter  byte
	separatorLetter byte
}

type poolData struct {
	bestScore   float64
	bestKey     [25]byte
	bestLock    sync.Mutex
	currentKeys []keyData
	currentLock sync.Mutex
	global      *globalData
}

type keyData struct {
	score float64
	key   [25]byte
}

func PlayfairCrack(ciphertext string, excludedLetter byte, separatorLetter byte, logVerbose bool) *CrackResult {
	numThreads := runtime.NumCPU()
	poolSize := 4
	numPools := numThreads / poolSize

	// Signal we are starting
	startTime := time.Now()
	if logVerbose {
		fmt.Printf("Cracking Cipher Text:\n%s\n\n", ciphertext)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var waitGroup sync.WaitGroup
	waitGroup.Add(numPools * poolSize)

	globalData := &globalData{
		solutionChan:    make(chan CrackResult),
		chanLock:        sync.Mutex{},
		cancel:          cancel,
		ciphertext:      []byte(ciphertext),
		excludedLetter:  excludedLetter,
		separatorLetter: separatorLetter,
	}

	// Start each pool
	for range numPools {
		poolData := &poolData{
			bestScore:   math.Inf(-1),
			bestLock:    sync.Mutex{},
			currentKeys: make([]keyData, poolSize),
			currentLock: sync.Mutex{},
			global:      globalData,
		}

		// Get a random starting keys
		for i := 0; i < poolSize; i++ {
			poolData.currentKeys[i].key = cipher.GenerateRandomKey(globalData.excludedLetter)
			plaintext := cipher.PlayfairDecrypt(globalData.ciphertext, poolData.currentKeys[i].key, globalData.excludedLetter)
			poolData.currentKeys[i].score = score.ScoreTextFast(plaintext, globalData.excludedLetter)
		}

		for i, keyData := range poolData.currentKeys {
			go processWorker(
				ctx,
				logVerbose,
				&waitGroup,
				poolData,
				i,
				keyData.key,
			)
		}
	}

	var solution CrackResult
	select {
	case solution = <-globalData.solutionChan:
		// A worker found the key, cancel other workers
		cancel()
	case <-ctx.Done():
		// Context canceled externally
	}

	// Wait for all goroutines to finish
	waitGroup.Wait()

	// Update elapsed time and return
	solution.ElapsedTime = time.Now().Sub(startTime)
	return &solution
}

func processWorker(
	ctx context.Context,
	logVerbose bool,
	waitGroup *sync.WaitGroup,
	poolData *poolData,
	pid int,
	initialKey [25]byte,
) {
	defer waitGroup.Done()

	// initialize search
	localBest := poolData.bestScore
	localBestKey := initialKey
	localSinceBest := 0

	temperature := 50.0
	coolingRate := 0.01
	epoch := 0

	for {
		// Check if other worker found solution
		select {
		case <-ctx.Done():
			// exit early
			return
		default:
		}

		// Run pool
		epoch++
		key, score := simulatedAnnealingCrack(
			ctx,
			logVerbose,
			poolData,
			pid,
			poolData.currentKeys[pid].key,
			temperature,
			0.1,
			coolingRate,
			1024,
			50000,
			5,
		)

		// compare to local best
		if score > localBest {
			localBest = score
			localBestKey = key
			localSinceBest = 0
		} else {
			localSinceBest++
		}

		// Check for solution
		if -3000 < localBest && checkForSolution(ctx, poolData.global, localBestKey) {
			return
		}
	}
}

func checkForSolution(
	ctx context.Context,
	globalData *globalData,
	key [25]byte,
) bool {
	globalData.chanLock.Lock()
	defer globalData.chanLock.Unlock()

	// Check if other already sent solution
	select {
	case <-ctx.Done():
		return false
	default:
	}

	// Check for solution
	solution := CrackResult{}
	plaintext := cipher.PlayfairDecrypt(globalData.ciphertext, key, globalData.excludedLetter)
	solution.PercentEnglish, solution.SegmentedText = score.ScoreTextSlow(plaintext, globalData.separatorLetter, 1.5)

	// We did not find solution
	if solution.PercentEnglish < THRESHHOLD_ENGLISH {
		return false
	}

	// Add result data
	var keyBuilder strings.Builder
	for _, b := range key {
		keyBuilder.WriteByte(b)
	}
	solution.Key = keyBuilder.String()
	solution.Plaintext = string(plaintext)

	select {
	case <-ctx.Done():
		// Context canceled, do not attempt to send
		return true
	case globalData.solutionChan <- solution:
		// Successfully sent the solution, cancel others
		globalData.cancel()
		return true
	}
}
