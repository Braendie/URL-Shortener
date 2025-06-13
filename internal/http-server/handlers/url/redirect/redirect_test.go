package redirect_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Braendie/url-shortener/internal/http-server/handlers/url/redirect"
	"github.com/Braendie/url-shortener/internal/http-server/handlers/url/redirect/mocks"
	"github.com/Braendie/url-shortener/internal/lib/api"
	"github.com/Braendie/url-shortener/internal/lib/logger/handlers/slogdiscard"
	"github.com/Braendie/url-shortener/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/assert/v2"
	"github.com/stretchr/testify/require"
)

func TestRedirectHandler(t *testing.T) {
	testCases := []struct {
		name      string
		alias     string
		url       string
		respError string
		mockError error
		code      int
	}{
		{
			name:  "valid",
			alias: "test_alias",
			url:   "https://www.google.com",
			code:  http.StatusFound,
		},
		{
			name:      "empty alias",
			alias:     "",
			respError: "invalid request",
			code:      http.StatusNotFound,
		},
		{
			name:      "not found",
			alias:     "some_alias",
			respError: "not found",
			mockError: storage.ErrURLNotFound,
			code:      http.StatusNotFound,
		},
		{
			name:      "storage error",
			alias:     "some_alias",
			respError: "internal error",
			mockError: errors.New("some error"),
			code:      http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			urlGetterMock := mocks.NewURLGetter(t)

			if tc.respError == "" || tc.mockError != nil {
				urlGetterMock.On("GetURL", tc.alias).
					Return(tc.url, tc.mockError).Once()
			}

			r := chi.NewRouter()
			r.Get("/{alias}", redirect.New(slogdiscard.NewDiscardLogger(), urlGetterMock))

			ts := httptest.NewServer(r)
			defer ts.Close()
			url := ts.URL + "/" + tc.alias
			if tc.respError == "" {
				redirectedToUrl, err := api.GetRedirect(url)
				require.NoError(t, err)
				assert.Equal(t, tc.url, redirectedToUrl)
			} else {
				req, err := http.NewRequest(http.MethodGet, url, nil)
				require.NoError(t, err)

				resp, err := (&http.Client{}).Do(req)
				require.NoError(t, err)

				defer resp.Body.Close()
				
				assert.Equal(t, tc.code, resp.StatusCode)
			}
		})
	}
}
