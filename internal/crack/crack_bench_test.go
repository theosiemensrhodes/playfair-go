package crack

import (
	"math"
	"math/rand"
	"playfaircrack/internal/cipher"
	"playfaircrack/internal/score"
	"testing"
)

// Helper function to convert a key string to a [25]byte array
func stringTo25Byte(s string) [25]byte {
	var arr [25]byte
	copy(arr[:], s)
	return arr
}

// func BenchmarkPlayfairCrack(b *testing.B) {
// 	rand.Seed(42)

// 	// Load static scoring components
// 	score.GetNgramScorerInstance()
// 	score.GetSegmentorInstance()
// 	score.GetDictionaryInstance()

// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		for i, ct := range testdata.BenchCiphertexts {
// 			result := PlayfairCrack(ct, 'J', 'X', false)
// 			pt := testdata.BenchPlaintexts[i]
// 			if result.Plaintext != pt {
// 				b.Fatalf("Failed cracking:\ngot %v,\nexpected %q", result.Plaintext, pt)
// 			}
// 		}
// 	}
// }

func BenchmarkInnerLoop(b *testing.B) {
	rand.Seed(42)

	// Load static scoring components
	score.GetNgramScorerInstance()
	score.GetSegmentorInstance()
	score.GetDictionaryInstance()

	ciphertext := []byte("XCIEIQWXFYHMVHPNYHQFFVXWWPFCMZIOXHYHPNOUDLNZNOHEVQDZDCEPDUVIOBNZIQAGDFVNPAOBHOQREHDMVIPEDRXHXCHFVGBXFVCOGKEGEPAKCSCPGPTXVIVZEUXCCOXCFNHEHFEIDZVNLCUTDZYFIYVIEINIEBHFMCBIXIGVEPPKYCEHIQSVXCBEHIPAIPIKEYAOSIHYHXEPPRPQVOFMFHBCXYEPXCBYDRZDFCRIXCIQZDIKZXEPWBRFEHDRZCSODMVILBVNEGRPCXXQZPPAXCVNHEOIOECWFVAXHYCX")
	currentKey := stringTo25Byte("ABCDEFGHIKLMNOPQRSTUVWXYZ")
	currentScore := score.ScoreTextFast(cipher.PlayfairDecrypt(ciphertext, currentKey, 'J'), 'X')
	bestScore := currentScore
	curTemp := 25.0
	triesBeforeStagnation := 80000
	iterSinceBest := 0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Copied from simulated_annealing inner loop
		// We have stagnated, check if we are at solution
		if -3000 < bestScore && iterSinceBest > triesBeforeStagnation {
			// Ignore return
		}

		candidateKey := cipher.PermuteKey(currentKey, 'J')
		candidatePlaintext := cipher.PlayfairDecrypt(ciphertext, candidateKey, 'J')
		candidateScore := score.ScoreTextFast(candidatePlaintext, 'X')

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
			iterSinceBest = 0
		} else {
			iterSinceBest++
		}
	}
}
