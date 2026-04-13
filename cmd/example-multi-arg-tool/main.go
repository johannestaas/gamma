package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/johannestaas/gamma"
	"github.com/johannestaas/gamma/internal/logx"
)

const (
	Reset  = "\033[0m"
	Purple = "\033[35m"
)

var logger *slog.Logger

func purple(s string) string {
	return Purple + s + Reset
}

var quitError = errors.New("quit")

func readMultilineInput() (string, error) {
	fmt.Fprintf(os.Stdout, "\nPrompt (press enter twice to send message, q to quit): ")

	scanner := bufio.NewScanner(os.Stdin)

	var lines []string
	emptyCount := 0

	for scanner.Scan() {
		line := scanner.Text()

		if strings.TrimSpace(line) == "" {
			emptyCount++
			if emptyCount >= 2 {
				break
			}
		} else {
			emptyCount = 0
		}

		lines = append(lines, line)
	}

	fmt.Fprintf(os.Stdout, "\nEnd of Message\n")
	if err := scanner.Err(); err != nil {
		return "", err
	}

	txt := strings.Join(lines, "\n")
	if strings.TrimSpace(txt) == "q" {
		return "", quitError
	}
	return txt, nil
}

func cardPullerToolWithArgs(args map[string]any) gamma.ToolResult {
	comment, cOk, cErr := gamma.GetOptionalArg(args, "dealer_comment", "")
	if cOk {
		fmt.Printf("\n*** Dealer made comment: %q\n", *comment)
	}
	if cErr != nil {
		return gamma.ErrorToToolError(cErr)
	}
	showCard, scErr := gamma.GetRequiredArg[bool](args, "show_card")
	if scErr != nil {
		return gamma.ErrorToToolError(scErr)
	}
	owner, oErr := gamma.GetRequiredArg[string](args, "owner")
	if oErr != nil {
		return gamma.ErrorToToolError(oErr)
	}
	ranks := []string{"ace", "2", "3", "4", "5", "6", "7", "8", "9", "10", "jack", "queen", "king"}
	suits := []string{"hearts", "diamonds", "clubs", "spades"}

	card := ranks[rand.Intn(len(ranks))] + " of " + suits[rand.Intn(len(suits))]

	if *showCard {
		fmt.Printf("\n*** %q picked up a card: %s\n", *owner, card)
	} else {
		fmt.Printf("\n*** %q picked up a card face down\n", *owner)
	}

	return gamma.ToolResult{
		Result: card,
	}
}

var cardPullerDef = gamma.CallableFunctionToolDef{
	Name:        "card_puller",
	Description: "pools a random card from a deck",
	Callable:    cardPullerToolWithArgs,
	Parameters: gamma.NewToolParameters(
		gamma.NamedProperty{
			Name:        "show_card",
			Type:        "boolean",
			Description: "whether to show the card to the user",
			Required:    true,
		},
		gamma.NamedProperty{
			Name:        "owner",
			Type:        "string",
			Description: "the role of the person taking the card, likely either user or dealer or player name",
			Required:    true,
		},
		gamma.NamedProperty{
			Name:        "dealer_comment",
			Type:        "string",
			Description: "this dealer sometimes makes a funny joke comment when he pulls a card, but not always",
			Required:    false,
		},
	),
}

var toolOpts = []gamma.Option{
	gamma.WithTool(cardPullerDef),
}

func main() {
	rootURL := flag.String("rootURL", "http://localhost:11434", "ollama root URL")
	model := flag.String("model", "gemma4", "the default LLM to use")
	prompt := flag.String("prompt", "You are a casino dealer with access to one tool to pull a random card. You play games with the user.", "the system prompt to use")
	flag.Parse()

	logger = logx.NewLogger(logx.Config{Level: "debug"})
	logger.Debug("example gamma client starting")

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Minute)
	defer cancel()

	logger.Info("testing gamma client from CLI",
		"root_url", *rootURL,
		"model", *model,
		"prompt", *prompt,
	)

	opts := append(toolOpts, gamma.WithRootURL(*rootURL))
	opts = append(opts, gamma.WithModel(*model))
	client := gamma.NewGammaClient(opts...)
	convo := client.NewConvo(*prompt)
	renderedThoughtLast := false
	for {
		var q string
		var err error
		q, err = readMultilineInput()
		if err != nil {
			if err == quitError {
				fmt.Fprintln(os.Stdout, "\nquitting")
				break
			} else {
				logger.Error("panicking due to error", "error", err)
				panic(err)
			}
		}
		ans := convo.Ask(ctx, q)
		thoughts := ""
		full := ""
		for chunk := range ans {
			if chunk.Err != nil {
				logger.Error("chat error", "chunk", chunk)
				continue
			}
			if chunk.Thinking != "" {
				if !renderedThoughtLast {
					fmt.Print("\n")
				}
				fmt.Print(purple(chunk.Thinking))
				renderedThoughtLast = true
				thoughts += chunk.Thinking
			}
			if chunk.Content != "" {
				if renderedThoughtLast {
					fmt.Print("\n")
				}
				fmt.Print(chunk.Content)
				renderedThoughtLast = false
				full += chunk.Content
			}
			/*
				if chunk.UsedTool {
					usedTool = true
				}
			*/
		}
	}
}
