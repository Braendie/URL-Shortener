package api

import (
	"errors"
	"fmt"
	"net/http"
)

var (
	ErrInvalidStatusCode = errors.New("invalid status code")
)

func GetRedirect(url string) (string, error) {
	const op = "api.GetRedirect"

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	if resp.StatusCode != http.StatusFound {
		return "", fmt.Errorf("%s: %w %v", op, ErrInvalidStatusCode, resp.StatusCode)
	}

	defer func() { _ = resp.Body.Close() }()

	return resp.Header.Get("Location"), nil
}
