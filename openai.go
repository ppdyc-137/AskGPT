package main

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type gptModel struct {
	client openai.Client
	model  string
	askCh  chan string
	respCh chan respMsg
}

type respMsg struct {
	content  string
	finished bool
}

func newClient(baseurl string, apikey string) openai.Client {
	return openai.NewClient(
		option.WithBaseURL(baseurl),
		option.WithAPIKey(apikey),
	)
}

func (m gptModel) newConversation() tea.Cmd {
	return func() tea.Msg {
		param := openai.ChatCompletionNewParams{
			Messages: []openai.ChatCompletionMessageParamUnion{},
			Seed:     openai.Int(1),
			Model:    m.model,
		}

		for {
			question := <-m.askCh
			param.Messages = append(param.Messages, openai.UserMessage(question))

			stream := m.client.Chat.Completions.NewStreaming(context.TODO(), param)

			// optionally, an accumulator helper can be used
			acc := openai.ChatCompletionAccumulator{}

			for stream.Next() {
				chunk := stream.Current()
				acc.AddChunk(chunk)

				if len(chunk.Choices) > 0 {
					m.respCh <- respMsg{content: chunk.Choices[0].Delta.Content}
				}
			}

			if stream.Err() != nil {
				panic(stream.Err())
			}
			m.respCh <- respMsg{finished: true}

			param.Messages = append(param.Messages, acc.Choices[0].Message.ToParam())
		}
	}
}

func (m gptModel) waitForResponse() tea.Cmd {
	return func() tea.Msg {
		return <-m.respCh
	}
}
