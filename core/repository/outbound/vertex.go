package outbound

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/temukan-co/monolith/config"
	"github.com/temukan-co/monolith/core/entity"
	"net/http"
)

type VertexOutbound struct {
	cfg *config.Config
}

func NewVertexOutbound(cfg *config.Config) *VertexOutbound {
	return &VertexOutbound{
		cfg: cfg,
	}
}

func (v *VertexOutbound) DoCallVertexAPIChat(ctx context.Context, request entity.BisonChatRequest) (*entity.BisonChatResponse, error) {
	jsonRequest, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	// Create a new HTTP client.
	client := http.Client{}

	// Create a new HTTP post request.
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, v.cfg.VertexURL+v.cfg.ChatModel, bytes.NewReader(jsonRequest))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", v.cfg.GoogleSecret))

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("Unexpected HTTP status code: %d", resp.StatusCode))
	}

	var response entity.BisonChatResponse
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}
