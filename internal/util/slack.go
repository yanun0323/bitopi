package util

import (
	"bytes"
	"context"
	"io"
	"net/http"

	"github.com/pkg/errors"
)

type NotifierType uint8

func NewSlackNotifier(token string) *SlackNotifier {
	return &SlackNotifier{
		token:      token,
		httpClient: http.DefaultClient,
	}
}

type SlackNotifier struct {
	token      string
	httpClient *http.Client
}

func (s SlackNotifier) Send(ctx context.Context, url Url, msg Messenger) ([]byte, int, error) {
	if len(url) == 0 {
		return nil, 0, errors.New("empty url")
	}

	if len(s.token) == 0 {
		return nil, 0, errors.New("empty token")
	}

	req, err := s.postRequest(s.token, msg, url)
	if err != nil {
		return nil, 0, err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, 0, err
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}

	return respBody, resp.StatusCode, nil
}

func (notifier SlackNotifier) postRequest(token string, msg Messenger, url Url) (*http.Request, error) {
	reqBody, err := msg.Marshal()
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, url.String(), bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("Authorization", "Bearer "+token)

	return req, nil
}
