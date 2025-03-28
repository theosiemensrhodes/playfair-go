package cipher

import (
	"math/rand"
)

func GenerateRandomKey(excludedLetter byte) [25]byte {
	var key [25]byte
	idx := 0
	for l := byte('A'); l <= 'Z'; l++ {
		if l != excludedLetter {
			key[idx] = l
			idx++
		}
	}
	rand.Shuffle(25, func(i, j int) { key[i], key[j] = key[j], key[i] })

	return key
}

func PermuteKey(key [25]byte, excludedLetter byte) [25]byte {
	if len(key) != 25 {
		panic("Key length must be 25")
	}
	r := rand.Uint32() % 100
	if r < 2 {
		for i := 0; i < rand.Intn(25)+1; i++ {
			key = swapChars(key)
		}
	} else if r < 5 {
		key = GenerateRandomKey(excludedLetter)
	} else if r < 10 {
		key = swapRows(key)
	} else if r < 16 {
		key = swapCols(key)
	} else {
		key = swapChars(key)
	}
	return key
}

func swapRows(key [25]byte) [25]byte {
	row1, row2 := rand.Intn(5), rand.Intn(5)
	base1, base2 := row1*5, row2*5
	for i := 0; i < 5; i++ {
		key[base1+i], key[base2+i] = key[base2+i], key[base1+i]
	}
	return key
}

func swapCols(key [25]byte) [25]byte {
	col1, col2 := rand.Intn(5), rand.Intn(5)
	for i := 0; i < 5; i++ {
		rowBase := i * 5
		key[rowBase+col1], key[rowBase+col2] = key[rowBase+col2], key[rowBase+col1]
	}
	return key
}

func swapChars(key [25]byte) [25]byte {
	idx1, idx2 := rand.Intn(25), rand.Intn(25)
	key[idx1], key[idx2] = key[idx2], key[idx1]
	return key
}

// PlayfairDecrypt decrypts the ciphertext using the given key.
func PlayfairDecrypt(ciphertext []byte, key [25]byte, excludedLetter byte) []byte {
	var position [26]int
	position[excludedLetter-'A'] = -1
	for i, char := range key {
		position[char-'A'] = i
	}

	ctlen := len(ciphertext)
	decryptedText := make([]byte, ctlen)

	for i := 1; i < ctlen; i += 2 {
		char1, char2 := ciphertext[i-1], ciphertext[i]
		pos1 := position[char1-'A']
		pos2 := position[char2-'A']
		row1, col1 := pos1/5, pos1%5
		row2, col2 := pos2/5, pos2%5

		if row1 == row2 {
			// Same row: shift left
			newCol1 := col1 - 1
			if newCol1 < 0 {
				newCol1 = 4
			}
			newCol2 := col2 - 1
			if newCol2 < 0 {
				newCol2 = 4
			}
			decryptedText[i-1] = key[row1*5+newCol1]
			decryptedText[i] = key[row2*5+newCol2]
		} else if col1 == col2 {
			// Same column: shift up
			newRow1 := row1 - 1
			if newRow1 < 0 {
				newRow1 = 4
			}
			newRow2 := row2 - 1
			if newRow2 < 0 {
				newRow2 = 4
			}
			decryptedText[i-1] = key[newRow1*5+col1]
			decryptedText[i] = key[newRow2*5+col2]
		} else {
			// Rectangle swap
			decryptedText[i-1] = key[row1*5+col2]
			decryptedText[i] = key[row2*5+col1]
		}
	}

	return decryptedText
}

// PlayfairEncrypt encrypts the plaintext using the given key.
func PlayfairEncrypt(plaintext []byte, key [25]byte, excludedLetter byte) []byte {
	var position [26]int
	position[excludedLetter-'A'] = -1
	for i, char := range key {
		position[char-'A'] = i
	}

	ptlen := len(plaintext)
	encryptedText := make([]byte, ptlen)

	for i := 1; i < ptlen; i += 2 {
		char1, char2 := plaintext[i-1], plaintext[i]
		pos1 := position[char1-'A']
		pos2 := position[char2-'A']
		row1, col1 := pos1/5, pos1%5
		row2, col2 := pos2/5, pos2%5

		if row1 == row2 {
			// Same row: shift right
			newCol1 := col1 + 1
			if newCol1 == 5 {
				newCol1 = 0
			}
			newCol2 := col2 + 1
			if newCol2 == 5 {
				newCol2 = 0
			}
			encryptedText[i-1] = key[row1*5+newCol1]
			encryptedText[i] = key[row2*5+newCol2]
		} else if col1 == col2 {
			// Same column: shift down
			newRow1 := row1 + 1
			if newRow1 == 5 {
				newRow1 = 0
			}
			newRow2 := row2 + 1
			if newRow2 == 5 {
				newRow2 = 0
			}
			encryptedText[i-1] = key[newRow1*5+col1]
			encryptedText[i] = key[newRow2*5+col2]
		} else {
			// Rectangle: swap columns
			encryptedText[i-1] = key[row1*5+col2]
			encryptedText[i] = key[row2*5+col1]
		}
	}

	return encryptedText
}
