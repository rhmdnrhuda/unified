// Package usecase implements application business logic. Each logic group in own file.
package usecase

import (
	"context"
	"time"

	"github.com/rhmdnrhuda/unified/core/entity"
)

//go:generate mockgen -source=interfaces.go -destination=./mocks_test.go -package=usecase_test

type (
	// Talent
	Talent interface {
		Create(ctx context.Context, req entity.TalentRequest) error
		GetTalent(ctx context.Context, university, major []string) (entity.Talent, error)
		Update(ctx context.Context, req entity.TalentRequest) error
	}

	TalentRepository interface {
		Create(ctx context.Context, data *entity.Talent) error
		Update(ctx context.Context, data *entity.Talent) error
		FindTalentByUniversityAndMajor(ctx context.Context, universities, majors []string) (entity.Talent, error)
	}

	// Vertex
	VertexOutBound interface {
		DoCallVertexAPIChat(ctx context.Context, request entity.BisonChatRequest, token string) (*entity.BisonChatResponse, error)
		DoCallVertexAPIText(ctx context.Context, request entity.BisonTextRequest, token string) (*entity.BisonTextResponse, error)
	}

	AdaOutBound interface {
		SendMessage(ctx context.Context, request entity.AdaRequest) error
		SendMessageButton(ctx context.Context, req entity.AdaButtonRequest) (string, error)
	}

	// Message
	Message interface {
		ProcessMessage(ctx context.Context, req entity.MessageRequest) (string, error)
	}

	UserRepository interface {
		Create(ctx context.Context, data *entity.User) error
		Update(ctx context.Context, data *entity.User) error
		FindUserByNumber(ctx context.Context, number string) (entity.User, error)
	}

	Cache interface {
		Set(key, token string, duration time.Duration)
		Get(key string) string
	}
)
