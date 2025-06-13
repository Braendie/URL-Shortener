package save_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Braendie/url-shortener/internal/http-server/handlers/url/save"
	"github.com/Braendie/url-shortener/internal/http-server/handlers/url/save/mocks"
	"github.com/Braendie/url-shortener/internal/lib/logger/handlers/slogdiscard"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type Response struct {
	Alias string `json:"alias"`
}

func TestSaveHandler(t *testing.T) {
	cases := []struct {
		name      string
		alias     string
		url       string
		respError string
		mockError error
		code      int
	}{
		{
			name:  "Success",
			alias: "test_alias",
			url:   "https://google.com",
			code:  http.StatusOK,
		},
		{
			name:  "Empty alias",
			alias: "",
			url:   "https://google.com",
			code:  http.StatusOK,
		},
		{
			name:      "Empty URL",
			url:       "",
			alias:     "some_alias",
			respError: "field URL is a required field URL",
			code:      http.StatusBadRequest,
		},
		{
			name:      "Invalid URL",
			url:       "some invalid URL",
			alias:     "some_alias",
			respError: "field URL is not a valid URL",
			code:      http.StatusBadRequest,
		},
		{
			name:      "SaveURL Error",
			url:       "https://google.com",
			alias:     "test_alias",
			respError: "failed to add url",
			mockError: errors.New("unexpected error"),
			code:      http.StatusInternalServerError,
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			urlSaverMock := mocks.NewURLSaver(t)

			if tc.respError == "" || tc.mockError != nil {
				urlSaverMock.On("SaveURL", tc.url, mock.AnythingOfType("string")).
					Return(int64(1), tc.mockError).
					Once()
			}

			handler := save.New(slogdiscard.NewDiscardLogger(), urlSaverMock, 6)

			input := fmt.Sprintf(`{"url": "%s", "alias": "%s"}`, tc.url, tc.alias)

			req, err := http.NewRequest(http.MethodPost, "/save", bytes.NewReader([]byte(input)))
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tc.code, rr.Code)

			var resp Response

			err = json.NewDecoder(rr.Body).Decode(&resp)
			require.NoError(t, err)
			if tc.respError == "" {
				if tc.alias != "" {
					assert.Equal(t, tc.alias, resp.Alias)
				} else {
					assert.NotEmpty(t, resp.Alias)
				}
			}
		})
	}
}
