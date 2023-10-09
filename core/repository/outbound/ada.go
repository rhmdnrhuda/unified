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

type AdaOutbound struct {
	cfg    *config.Config
	client http.Client
}

func NewAdaOutbound(cfg *config.Config) *AdaOutbound {
	return &AdaOutbound{
		cfg: cfg,
		client: http.Client{
			Timeout: time.Second * 10,
		},
	}
}

func (a *AdaOutbound) SendMessage(ctx context.Context, request entity.AdaRequest) error {
	_, err := a.doCallApi(ctx, request)
	if err != nil {
		return err
	}

	return nil
}

func (a *AdaOutbound) SendMessageButton(ctx context.Context, req entity.AdaButtonRequest) (string, error) {
	res, err := a.doCallApi(ctx, entity.AdaButtonRequest{
		Platform:   req.Platform,
		From:       req.From,
		To:         req.To,
		Type:       "button",
		Text:       req.Text,
		HeaderType: req.HeaderType,
		Header:     req.Header,
		Footer:     req.Footer,
		Buttons:    req.Buttons,
	})
	if err != nil {
		return "", err
	}

	if len(res.Data) <= 0 {
		return "", err
	}

	return res.Data[0], nil
}

func (a *AdaOutbound) doCallApi(ctx context.Context, request interface{}) (entity.MessageResponse, error) {
	var response entity.MessageResponse

	jsonRequest, err := json.Marshal(request)
	if err != nil {
		return response, err
	}

	req, err := http.NewRequest(http.MethodPost, a.cfg.AdaHostURL, bytes.NewReader(jsonRequest))
	if err != nil {
		return response, err
	}

	req.Header.Set("Content-Type", "application/json")

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", a.cfg.AdaAPISecret))

	resp, err := a.client.Do(req)
	if err != nil {
		return response, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return response, errors.New(fmt.Sprintf("%d", resp.StatusCode))
	}

	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return response, err
	}

	return response, nil
}
