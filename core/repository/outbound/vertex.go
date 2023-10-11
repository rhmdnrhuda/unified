package outbound

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/rhmdnrhuda/unified/config"
	"github.com/rhmdnrhuda/unified/core/entity"
	"net/http"
	"time"
)

type VertexOutbound struct {
	cfg    *config.Config
	client http.Client
}

func NewVertexOutbound(cfg *config.Config) *VertexOutbound {
	return &VertexOutbound{
		cfg: cfg,
		client: http.Client{
			Timeout: time.Second * 60,
		},
	}
}

func (v *VertexOutbound) DoCallVertexAPIChat(ctx context.Context, request entity.BisonChatRequest, token string) (*entity.BisonChatResponse, error) {
	ctx = context.Background()
	jsonRequest, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	fmt.Println(string(jsonRequest))

	req, err := http.NewRequest(http.MethodPost, v.cfg.VertexURL+v.cfg.ChatModel, bytes.NewReader(jsonRequest))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	var resp *http.Response

	// Retry max 5 times when error != nil
	for i := 0; i < 5; i++ {
		resp, err = v.client.Do(req)
		if err == nil {
			break
		}
	}

	// If the request failed after 5 retries, return the error
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

func (v *VertexOutbound) DoCallVertexAPIText(ctx context.Context, request entity.BisonTextRequest,
	token string) (*entity.BisonTextResponse, error) {
	jsonRequest, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	fmt.Println(string(jsonRequest))

	req, err := http.NewRequest(http.MethodPost, v.cfg.VertexURL+v.cfg.TextModel, bytes.NewReader(jsonRequest))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	var resp *http.Response

	// Retry max 5 times when error != nil
	for i := 0; i < 5; i++ {
		resp, err = v.client.Do(req)
		if err == nil {
			break
		}
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("Unexpected HTTP status code: %d", resp.StatusCode))
	}

	var response entity.BisonTextResponse
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}
