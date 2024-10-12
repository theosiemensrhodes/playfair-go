package main

import (
	"bufio"
	"bytes"
	"fmt"
	"math"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/urfave/cli/v2"
)

func err(args ...any) {
	fmt.Fprintln(os.Stderr, args...)
	os.Exit(1)
}

func tryOpenFile(_ *cli.Context, filepath string) error {
	if filetext != "" {
		panic("Called try open file multiple times")
	}

	file, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("The file %v could not be opened", filepath)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return fmt.Errorf("The file %v could not be read", filepath)
		}
		return fmt.Errorf("The file %v contains no text", filepath)
	}

	filetext = scanner.Text()
	return nil
}

func verifyKey(_ *cli.Context, key string) error {
	if len(key) != 25 {
		return fmt.Errorf("key value %v not of length 25", key)
	}
	upper_string := strings.ToUpper(key)
	for _, l := range strings.Split("ABCDEFGHIJKLMNOPQRSTUVWXYZ", "") {
		if l != "x" && !strings.Contains(upper_string, l) {
			return fmt.Errorf("key value %v missing letter %v", key, l)
		}
	}
	return nil
}

// Global cli arguments
var key string
var file string
var filetext string
var logVerbose bool

func main() {
	app := &cli.App{
		Name:     "Playfair Cipher Crack",
		Usage:    "Cracking sufficiently long (200+ character) Playfair encrypted ciphertext",
		Version:  "v19.99.0",
		Compiled: time.Now(),
		Authors: []*cli.Author{
			{
				Name:  "Theo Siemens-Rhodes",
				Email: "theo.siemensrhodes@gmail.com",
			},
		},
		Commands: []*cli.Command{
			{
				Name:        "crack",
				Aliases:     []string{"c"},
				Usage:       "Crack the ciphertext given to stdin",
				Description: "*describe how to crack*",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "file",
						Aliases:     []string{"f"},
						Destination: &file,
						Usage:       "Load ciphertext from `FILE` instead",
						Action:      tryOpenFile,
					},
					&cli.BoolFlag{
						Name:        "verbose",
						Aliases:     []string{"v"},
						Destination: &logVerbose,
						Usage:       "Log the cracking process verbosely",
					},
				},
				Action: crackCommand,
			},
			{
				Name:        "decrypt",
				Aliases:     []string{"d"},
				Usage:       "Decrypt the ciphertext given to stdin with the key passed in",
				Description: "*desc*",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "key",
						Aliases:     []string{"k"},
						Destination: &key,
						Usage:       "Decrypt with `KEY`",
						Required:    true,
						Action:      verifyKey,
					},
					&cli.StringFlag{
						Name:        "file",
						Aliases:     []string{"f"},
						Destination: &file,
						Usage:       "Load ciphertext from `FILE` instead",
						Action:      tryOpenFile,
					},
				},
				Action: func(cCtx *cli.Context) error {
					fmt.Println(file, filetext, key)
					return nil
				},
			},
			{
				Name:        "encrypt",
				Aliases:     []string{"e"},
				Usage:       "Encrypt the plaintext given to stdin",
				Description: "*desc*",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "key",
						Aliases:     []string{"k"},
						Destination: &key,
						Usage:       "Encrypt with `KEY`",
						Action:      verifyKey,
					},
					&cli.StringFlag{
						Name:        "file",
						Aliases:     []string{"f"},
						Destination: &file,
						Usage:       "Load ciphertext from `FILE` instead",
						Action:      tryOpenFile,
					},
				},
				Action: func(cCtx *cli.Context) error {
					fmt.Println(file, filetext, key)
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
	}
}

func crackCommand(cCtx *cli.Context) error {
	var ciphertext string
	if filetext == "" {
		// Read from stdin
		// Check if the input is a pipe or file (not a terminal)
		stat, statErr := os.Stdin.Stat()
		if statErr != nil {
			return statErr
		}
		if (stat.Mode() & os.ModeCharDevice) != 0 {
			return fmt.Errorf("No input provided, please use input redirection")
		}

		// Read ciphertext
		scanner := bufio.NewScanner(os.Stdin)
		if !scanner.Scan() {
			return fmt.Errorf("No input provided")
		}
		ciphertext = scanner.Text()

		if scanErr := scanner.Err(); scanErr != nil {
			return scanErr
		}
	} else {
		// Read from file
		ciphertext = filetext
	}

	// Get a random starting key
	bestKey := []byte("ABCDEFGHIKLMNOPQRSTUVWXYZ")
	rand.Shuffle(len(bestKey), func(i, j int) { bestKey[i], bestKey[j] = bestKey[j], bestKey[i] })

	bestScore := math.Inf(-1)
	i := 0

	fmt.Println("Cracking Playfair")
	fmt.Printf("    Initial Key: %s\n", bestKey)
	fmt.Printf("    Ciphertext: %s\n", ciphertext)

	for {
		i++
		key, score, success := simulatedAnnealingCrack(
			ciphertext,
			1024,
			50.0,
			1.0,
			0.01,
			80000,
			bestKey,
		)

		if success {
			os.Exit(1)
		}

		if score > bestScore {
			bestScore = score
			bestKey = bytes.Clone(key)
			fmt.Printf("Best score so far: %f, on iteration %d\n", score, i)
			fmt.Printf("    Key: %s\n", key)
			plainText := playfairDecrypt(ciphertext, key)
			fmt.Printf("    Plaintext: %s\n", plainText)
		}
	}
}
