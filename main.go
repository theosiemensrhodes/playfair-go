package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/urfave/cli/v2"
)

// Global cli arguments
var key string
var filepath string
var text string
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
						Destination: &filepath,
						Usage:       "Load ciphertext from `FILE` instead",
						Action:      validateAndOpenFile,
					},
					&cli.BoolFlag{
						Name:        "verbose",
						Aliases:     []string{"v"},
						Destination: &logVerbose,
						Usage:       "Log the cracking process verbosely",
					},
				},
				Before: gatherInputText,
				Action: func(cCtx *cli.Context) error {
					err, ciphertext := validateAndTransformCiphertext(text, 'J')
					if err != nil {
						return err
					}

					return playfairCrack(ciphertext, 'J', 'X')
				},
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
					},
					&cli.StringFlag{
						Name:        "file",
						Aliases:     []string{"f"},
						Destination: &filepath,
						Usage:       "Load ciphertext from `FILE` instead",
						Action:      validateAndOpenFile,
					},
				},
				Before: gatherInputText,
				Action: func(cCtx *cli.Context) error {
					err, ciphertext := validateAndTransformCiphertext(text, 'J')
					if err != nil {
						return err
					}

					err, key = validateAndTransformKey(key, 'J')
					if err != nil {
						return err
					}

					fmt.Printf("Decrypting Text:\n%s\n\n", text)
					fmt.Printf("With Key: %s\n\n", key)

					plaintext := playfairDecrypt(ciphertext, []byte(strings.ToUpper(key)))

					fmt.Printf("Raw Plaintext:\n%s\n\n", plaintext)

					fmt.Printf("Segmented Plaintext:\n")
					printSegmentedPlaintext(plaintext, 'X')

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
					},
					&cli.StringFlag{
						Name:        "file",
						Aliases:     []string{"f"},
						Destination: &filepath,
						Usage:       "Load ciphertext from `FILE` instead",
						Action:      validateAndOpenFile,
					},
				},
				Before: gatherInputText,
				Action: func(cCtx *cli.Context) error {
					err, plainText := validateAndTransformPlaintext(text, 'J', 'I', 'X')
					if err != nil {
						return err
					}

					err, key = validateAndTransformKey(key, 'J')
					if err != nil {
						return err
					}

					fmt.Printf("Encrypting Text:\n%s\n\n", text)
					fmt.Printf("With Key: %s\n\n", key)

					ciphertext := playfairEncrypt(plainText, []byte(key))

					fmt.Printf("Ciphertext:\n%s\n", ciphertext)

					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
	}
}

func gatherInputText(cCtx *cli.Context) error {
	if text == "" {
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
		text = scanner.Text()

		if scanErr := scanner.Err(); scanErr != nil {
			return scanErr
		}
	}

	// Else we have read from file

	return nil
}
