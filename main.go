package main

import (
	"fmt"
	"log"
	"context"
	"os"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()
	model := client.GenerativeModel("gemini-1.0-pro")
	cs := model.StartChat()

	send := func(msg string) *genai.GenerateContentResponse {
		fmt.Printf("== Me: %s\n== Model:\n", msg)
		res, err := cs.SendMessage(ctx, genai.Text(msg))
		if err != nil {
			log.Fatal(err)
		}
		return res
	}

	res := send("Can you name some brands of air fryer?")
	printResponse(res)
	// iter := cs.SendMessageStream(ctx, genai.Text("Which one of those do you recommend?"))
	// for {
	// 	res, err := iter.Next()
	// 	if err == iterator.Done {
	// 		break
	// 	}
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	printResponse(res)
	// }

	// for i, c := range cs.History {
	// 	log.Printf("    %d: %+v", i, c)
	// }
	// res = send("Why do you like the Philips?")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// printResponse(res)

	// bubble tea
	p := tea.NewProgram(InitialModel())

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

func printResponse(resp *genai.GenerateContentResponse) {
	for _, cand := range resp.Candidates {
		if cand.Content != nil {
			for _, part := range cand.Content.Parts {
				fmt.Println(part)
			}
		}
	}
	fmt.Println("---")
}

type (
	errMsg error
)
