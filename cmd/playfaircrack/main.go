package main

import (
	"fmt"
	"os"
	"playfaircrack/internal/cipher"
	"playfaircrack/internal/cmdutil"
	"playfaircrack/internal/crack"
	"playfaircrack/internal/score"
	"time"

	"github.com/urfave/cli/v2"
)

// Global cli arguments
var key string
var filepath string
var logVerbose bool

func main() {
	app := &cli.App{
		Name:     "Playfair Cipher Crack",
		Usage:    "Cracking sufficiently long (200+ character) Playfair encrypted ciphertext",
		Version:  "v0.0",
		Compiled: time.Now(),
		Authors: []*cli.Author{
			{
				Name:  "Theo Siemens-Rhodes",
				Email: "theo.siemensrhodes@gmail.com",
			},
		},
		Commands: []*cli.Command{
			{
				Name:    "crack",
				Aliases: []string{"c"},
				Usage:   "Crack the ciphertext given to stdin",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "file",
						Aliases:     []string{"f"},
						Destination: &filepath,
						Usage:       "Load ciphertext from `FILE` instead",
					},
					&cli.BoolFlag{
						Name:        "verbose",
						Aliases:     []string{"v"},
						Destination: &logVerbose,
						Usage:       "Log the cracking process verbosely",
					},
				},
				Action: func(cCtx *cli.Context) error {
					err, text := cmdutil.GatherInput(filepath)
					if err != nil {
						return err
					}

					err, ciphertext := cmdutil.ValidateAndTransformCiphertext(text, 'J')
					if err != nil {
						return err
					}

					result := crack.PlayfairCrack(ciphertext, 'J', 'X', logVerbose)

					// Log result
					if logVerbose {
						fmt.Printf("\nSolution found, in %v!\n", result.ElapsedTime)
						fmt.Printf("Key: %s\n", result.Key)
						fmt.Printf("English Word Score %2.2f\n\n", 100.0*result.PercentEnglish)

						fmt.Printf("Raw Plaintext:\n%s\n\n", result.Plaintext)

						fmt.Printf("Segmented Plaintext:\n")
						for _, word := range result.SegmentedText {
							fmt.Printf("%s ", word)
						}
						fmt.Printf("\n")
					} else {
						fmt.Printf("%f, %f, %s", score.ScoreTextFast([]byte(result.Plaintext), 'X'), result.PercentEnglish, result.Plaintext)
						for _, word := range result.SegmentedText {
							fmt.Printf("%s ", word)
						}
						// fmt.Printf("%-4.4f ", score.ScoreTextFast([]byte(result.Plaintext), 'X'))
						// fmt.Printf("%s\n", result.Plaintext)
					}

					return nil
				},
			},
			{
				Name:    "decrypt",
				Aliases: []string{"d"},
				Usage:   "Decrypt the ciphertext given to stdin with the key passed in",
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
					},
				},
				Action: func(cCtx *cli.Context) error {
					err, text := cmdutil.GatherInput(filepath)
					if err != nil {
						return err
					}

					err, ciphertext := cmdutil.ValidateAndTransformCiphertext(text, 'J')
					if err != nil {
						return err
					}

					var validKey [25]byte
					err, validKey = cmdutil.ValidateAndTransformKey(key, 'J')
					if err != nil {
						return err
					}

					fmt.Printf("Decrypting Text:\n%s\n\n", text)
					fmt.Printf("With Key: %s\n\n", key)

					plaintext := cipher.PlayfairDecrypt([]byte(ciphertext), validKey, 'J')

					fmt.Printf("Raw Plaintext:\n%s\n\n", plaintext)

					fmt.Printf("Segmented Plaintext:\n")
					cmdutil.PrintSegmentedPlaintext(plaintext, 'X')

					return nil
				},
			},
			{
				Name:    "encrypt",
				Aliases: []string{"e"},
				Usage:   "Encrypt the plaintext given to stdin",
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
					},
				},
				Action: func(cCtx *cli.Context) error {
					err, text := cmdutil.GatherInput(filepath)
					if err != nil {
						return err
					}

					err, plainText := cmdutil.ValidateAndTransformPlaintext(text, 'J', 'I', 'X')
					if err != nil {
						return err
					}

					var validKey [25]byte
					err, validKey = cmdutil.ValidateAndTransformKey(key, 'J')
					if err != nil {
						return err
					}

					fmt.Printf("Encrypting Text:\n%s\n\n", text)
					fmt.Printf("With Key: %s\n\n", key)

					ciphertext := cipher.PlayfairEncrypt([]byte(plainText), validKey, 'J')

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
