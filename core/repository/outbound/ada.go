package outbound

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/rhmdnrhuda/unified/config"
	"github.com/rhmdnrhuda/unified/core/entity"
	"io"
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

func (a *AdaOutbound) SendButton(ctx context.Context, req entity.AdaButtonRequest) error {
	_, err := a.doCallApi(ctx, entity.AdaButtonRequest{
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

	return err
}

func (a *AdaOutbound) doCallApi(ctx context.Context, request interface{}) (io.ReadCloser, error) {
	jsonRequest, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, a.cfg.AdaHostURL, bytes.NewReader(jsonRequest))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", a.cfg.AdaAPISecret))

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("%d", resp.StatusCode))
	}

	return resp.Body, nil
}
