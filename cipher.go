package main

import "math/rand"

func generateRandomKey(excludedLetter byte) []byte {
	var key []byte
	letters := []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	for _, l := range letters {
		if l != excludedLetter {
			key = append(key, l)
		}
	}
	rand.Shuffle(len(key), func(i, j int) { key[i], key[j] = key[j], key[i] })
	return key
}

func permuteKey(key []byte) {
	if len(key) != 25 {
		panic("Key length must be 25")
	}
	r := rand.Uint32() % 100
	if r < 2 {
		for i := 0; i < rand.Intn(25)+1; i++ {
			swapChars(key)
		}
	} else if r < 6 {
		swapRows(key)
	} else if r < 10 {
		swapCols(key)
	} else {
		swapChars(key)
	}
}

func swapRows(key []byte) {
	row1, row2 := rand.Intn(5), rand.Intn(5)
	for i := 0; i < 5; i++ {
		key[row1*5+i], key[row2*5+i] = key[row2*5+i], key[row1*5+i]
	}
}

func swapCols(key []byte) {
	col1, col2 := rand.Intn(5), rand.Intn(5)
	for i := 0; i < 5; i++ {
		key[i*5+col1], key[i*5+col2] = key[i*5+col2], key[i*5+col1]
	}
}

func swapChars(key []byte) {
	idx1, idx2 := rand.Intn(25), rand.Intn(25)
	key[idx1], key[idx2] = key[idx2], key[idx1]
}

func playfairDecrypt(ciphertext string, key []byte) string {
	if len(key) != 25 {
		panic("Key length must be 25")
	}
	position := make(map[byte]int, 25)
	for i, char := range key {
		position[char] = i
	}
	decryptedText := make([]byte, len(ciphertext))
	for i := 1; i < len(ciphertext); i += 2 {
		char1, char2 := ciphertext[i-1], ciphertext[i]
		pos1, pos2 := position[char1], position[char2]
		row1, col1 := pos1/5, pos1%5
		row2, col2 := pos2/5, pos2%5
		if row1 == row2 {
			decryptedText[i-1] = key[row1*5+(col1-1+5)%5]
			decryptedText[i] = key[row2*5+(col2-1+5)%5]
		} else if col1 == col2 {
			decryptedText[i-1] = key[((row1-1+5)%5)*5+col1]
			decryptedText[i] = key[((row2-1+5)%5)*5+col2]
		} else {
			decryptedText[i-1] = key[row1*5+col2]
			decryptedText[i] = key[row2*5+col1]
		}
	}
	return string(decryptedText)
}

func playfairEncrypt(plaintext string, key []byte) string {
	if len(key) != 25 {
		panic("Key length must be 25")
	}
	position := make(map[byte]int, 25)
	for i, char := range key {
		position[char] = i
	}
	encryptedText := make([]byte, len(plaintext))
	for i := 1; i < len(plaintext); i += 2 {
		char1, char2 := plaintext[i-1], plaintext[i]
		pos1, pos2 := position[char1], position[char2]
		row1, col1 := pos1/5, pos1%5
		row2, col2 := pos2/5, pos2%5
		if row1 == row2 {
			// Same row: shift right
			encryptedText[i-1] = key[row1*5+(col1+1)%5]
			encryptedText[i] = key[row2*5+(col2+1)%5]
		} else if col1 == col2 {
			// Same column: shift down
			encryptedText[i-1] = key[((row1+1)%5)*5+col1]
			encryptedText[i] = key[((row2+1)%5)*5+col2]
		} else {
			// Rectangle: swap columns
			encryptedText[i-1] = key[row1*5+col2]
			encryptedText[i] = key[row2*5+col1]
		}
	}
	return string(encryptedText)
}
