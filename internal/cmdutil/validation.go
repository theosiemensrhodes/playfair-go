package cmdutil

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"unicode"

	"github.com/divan/num2words"
)

func GatherInput(filepath string) (error, string) {
	var scanner *bufio.Scanner

	if filepath != "" {
		file, err := os.Open(filepath)
		if err != nil {
			return fmt.Errorf("The file %v could not be opened", filepath), ""
		}
		defer file.Close()

		// Read file
		scanner = bufio.NewScanner(file)
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				return fmt.Errorf("The file %v could not be read", filepath), ""
			}
			return fmt.Errorf("The file %v contains no text", filepath), ""
		}
	} else {
		// Read from stdin
		// Check if the input is a pipe or file (not a terminal)
		stat, statErr := os.Stdin.Stat()
		if statErr != nil {
			return statErr, ""
		}
		if (stat.Mode() & os.ModeCharDevice) != 0 {
			return fmt.Errorf("No input provided, please use input redirection"), ""
		}

		// Read ciphertext
		scanner = bufio.NewScanner(os.Stdin)
		if !scanner.Scan() {
			return fmt.Errorf("No input provided"), ""
		}
	}

	text := scanner.Text()

	if scanErr := scanner.Err(); scanErr != nil {
		return scanErr, ""
	}

	return nil, text
}

func ValidateAndTransformKey(key string, excludedLetter rune) (error, [25]byte) {
	outKey := new([25]byte)

	if key != "" {
		key = strings.ToUpper(key)
		var transformed strings.Builder
		for _, l := range key {
			if l == excludedLetter || strings.ContainsRune(transformed.String(), l) {
				continue
			}
			transformed.WriteRune(l)
		}

		for _, l := range "ABCDEFGHIJKLMNOPQRSTUVWXYZ" {
			if l == excludedLetter || strings.ContainsRune(transformed.String(), l) {
				continue
			}

			transformed.WriteRune(l)
		}

		key = transformed.String()
	} else {
		var byteKey []byte
		letters := []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
		for _, l := range letters {
			if l != byte(excludedLetter) {
				byteKey = append(byteKey, l)
			}
		}
		rand.Shuffle(len(byteKey), func(i, j int) { byteKey[i], byteKey[j] = byteKey[j], byteKey[i] })

		key = string(byteKey)
	}

	if len(key) != 25 {
		panic("Transformed key is not of length 25")
	}

	for i, letter := range key {
		outKey[i] = byte(letter)
	}

	return nil, *outKey
}

func ValidateAndTransformCiphertext(ciphertext string, excludedLetter rune) (error, string) {
	if len(ciphertext)%2 != 0 {
		return fmt.Errorf("Ciphertexts must be aligned on the two letter boundary"), ""
	}

	// Ensure uppercase
	ciphertext = strings.ToUpper(ciphertext)

	// Preallocate space for the transformed string
	transformed := make([]rune, 0, len(ciphertext))
	var prevL rune
	for i, l := range ciphertext {
		// Ensure letter and not excluded
		if !unicode.IsLetter(l) {
			return fmt.Errorf("Ciphertexts must not contain any non-letters, %c", l), ""
		}
		if l == excludedLetter {
			return fmt.Errorf("Ciphertexts must not contain %c, the excluded letter", l), ""
		}

		// Ensure no pairs
		if i%2 == 1 && l == prevL {
			return fmt.Errorf("Ciphertexts must not contain letter pairs aligned on the two letter boundary, %c%c", l, l), ""
		}

		transformed = append(transformed, l)
		prevL = l
	}

	return nil, string(transformed)
}

func ValidateAndTransformPlaintext(plaintext string, exc, rep, sep rune) (error, string) {
	// Ensure uppercase
	plaintext = strings.ToUpper(plaintext)

	// Handle converting all to letters
	var builder strings.Builder
	var number []rune
	for _, l := range plaintext {
		// Skip whitespace
		if unicode.IsSpace(l) || unicode.IsPunct(l) {
			continue
		}

		// Handle numbers
		if unicode.IsNumber(l) {
			number = append(number, l)
		} else if len(number) != 0 {
			// We have a full text number
			num, err := strconv.Atoi(string(number))
			if err != nil {
				number = number[:0]
				continue
			}

			// Convert to string
			numberString := num2words.Convert(num)

			// Remove whitespace and uppercase
			numberString = strings.ReplaceAll(numberString, " ", "")
			numberString = strings.ToUpper(numberString)

			// Write and reset
			builder.WriteString(numberString)
			number = number[:0]
		}

		// Ensure letter
		if !unicode.IsLetter(l) {
			return fmt.Errorf("Ciphertexts must not contain any non-letters, %c", l), ""
		}

		// Replace excludedLetters with replacements
		if l == exc {
			l = rep
		}

		builder.WriteRune(l)
	}

	// Playfair transformation
	var prevL rune
	var i int = 0
	var transformed string = builder.String()
	builder.Reset()
	for _, l := range transformed {
		// Ensure no pairs
		if i%2 == 1 && l == prevL {
			builder.WriteRune(sep)
			i++
		}

		builder.WriteRune(l)
		prevL = l
		i++
	}

	// Ensure alignment
	if builder.Len()%2 != 0 {
		builder.WriteRune(sep)
	}

	return nil, builder.String()
}
