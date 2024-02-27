package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type GeminiModel struct {
	ctx context.Context
	client *genai.Client
	chatSession *genai.ChatSession
}

func NewGemini (ctx context.Context) *GeminiModel {
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	model := client.GenerativeModel("gemini-1.0-pro")
	cs := model.StartChat()
	return &GeminiModel{
		ctx: ctx,
		client: client,
		chatSession : cs,
	}
}

func (g *GeminiModel) SendMessage (msg string) string{
	res, err := g.chatSession.SendMessage(g.ctx, genai.Text(msg))
	if err != nil {
		log.Fatal(err)
	}
	for _, cand := range res.Candidates {
		if cand.Content != nil {
			for _, part := range cand.Content.Parts {
				return fmt.Sprint(part) 
			}
		}
	}
	return "" 
}