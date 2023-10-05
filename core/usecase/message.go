package usecase

import (
	"context"
	"fmt"
	"github.com/temukan-co/monolith/core/entity"
)

var UniBuddy = make(map[string][]entity.Message)

type MessageUseCase struct {
	vertex VertexOutBound
}

func NewMessageUseCase(vertex VertexOutBound) *MessageUseCase {
	return &MessageUseCase{
		vertex: vertex,
	}
}

func (m *MessageUseCase) ProcessMessage(ctx context.Context, req entity.MessageRequest) (string, error) {
	bisonChatReq := initBisonChatUniBuddyRequest()
	var messages []entity.Message
	if val, ok := UniBuddy[req.UserID]; ok {
		messages = val
	}

	fmt.Println("current message", messages)

	messages = append(messages, entity.Message{
		Author:  "user",
		Content: req.Message,
	})

	bisonChatReq.Instances[0].Messages = messages
	res, err := m.vertex.DoCallVertexAPIChat(ctx, bisonChatReq)
	if err != nil {
		return "", err
	}

	messages = append(messages, res.Predictions[0].Candidates[0])
	UniBuddy[req.UserID] = messages

	return res.Predictions[0].Candidates[0].Content, nil
}

func initBisonChatUniBuddyRequest() entity.BisonChatRequest {
	return entity.BisonChatRequest{
		Instances: []entity.Instance{
			{
				Context: "I am Unified, a student personal assistant. I help students choose universities and majors by providing recommendations and answering their questions. If the student has already chosen a university and major, I will respond with a JSON object in the format {university: UGM, major: law}",
				Examples: []entity.Example{
					{
						Input: entity.Content{
							Content: "join uni-buddy",
						},
						Output: entity.Content{
							Content: "Do you have any university preference?",
						},
					},
					{
						Input: entity.Content{
							Content: "hi",
						},
						Output: entity.Content{
							Content: "Hi, Do you have any university preference?",
						},
					},
				},
			},
		},
		Parameters: entity.Parameter{
			Temperature:     0.3,
			MaxOutputTokens: 200,
			TopP:            0.8,
			TopK:            40,
		},
	}
}
