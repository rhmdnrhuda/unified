// Package usecase implements application business logic. Each logic group in own file.
package usecase

import (
	"context"

	"github.com/temukan-co/monolith/core/entity"
)

//go:generate mockgen -source=interfaces.go -destination=./mocks_test.go -package=usecase_test

type (
	// Talent
	Talent interface {
		Create(ctx context.Context, req entity.TalentRequest) error
		GetTalent(ctx context.Context, university, major string) (entity.Talent, error)
		Update(ctx context.Context, req entity.TalentRequest) error
	}

	TalentRepository interface {
		Create(ctx context.Context, data *entity.Talent) error
		Update(ctx context.Context, data *entity.Talent) error
		FindTalentByUniversityAndMajor(ctx context.Context, university, major string) (entity.Talent, error)
	}

	// Vertex
	VertexOutBound interface {
		DoCallVertexAPIChat(ctx context.Context, request entity.BisonChatRequest) (*entity.BisonChatResponse, error)
	}

	// Message
	Message interface {
		ProcessMessage(ctx context.Context, req entity.MessageRequest) (string, error)
	}
)
