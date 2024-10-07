package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
)

func readNthLine(filename string, n int) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for i := 0; i < n; i++ {
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				return "", err
			}
			return "", fmt.Errorf("file has fewer than %d lines", n)
		}
	}

	return scanner.Text(), nil
}

func main() {
	ciphertext, err := readNthLine("../ciphertexts_playfair.txt", 30)
	if err != nil {
		fmt.Println("Error reading file:", err)
		os.Exit(1)
	}
	ciphertext = "RFKPQAZXFHDHGIWHEFXHHKXFXHGRHWSWEXKHTDADZXFOKDWTWGWZGYFZDYRUXKYDWOQKXQXPEHXFWZIGSPKIMFMWAOBQOFSWEQELRHQKDYHKXYFDDGKGVRFAXHIZEXQATHRZKIYTUEPTMWZXHRXQPSUAQXZXYXQATXXHDOFQFYBQWXWGRMURTXDIDQTPKIFDBDVRCGHKXHEOFDDGXHXETHKIUDIZHRFDQGTXDFWOUDLQCXSPGQDYANXKHKGFLXVEWFXBRUZRYDDGZAANEXKUAVMFBQIXDKDHDNWOCBHKDYCK"

	bestScore := math.Inf(-1)
	bestKey := []byte("ABCDEFGHIKLMNOPQRSTUVWXYZ")
	i := 0

	fmt.Println("Cracking Playfair")

	for {
		i++
		key, score := simulatedAnnealingCrack(
			ciphertext,
			4096,
			30.0,
			0.03,
			bestKey,
		)

		if score > bestScore {
			bestScore = score
			bestKey = copyByteArr(key)
			fmt.Printf("Best score so far: %f, on iteration %d\n", score, i)
			fmt.Printf("    Key: %s\n", key)
			plainText := playfairDecrypt(ciphertext, key)
			fmt.Printf("    Plaintext: %s\n", plainText)
		}
	}
}
