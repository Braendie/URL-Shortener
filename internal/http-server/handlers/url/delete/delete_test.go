package delete_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Braendie/url-shortener/internal/http-server/handlers/url/delete"
	"github.com/Braendie/url-shortener/internal/http-server/handlers/url/delete/mocks"
	"github.com/Braendie/url-shortener/internal/lib/logger/handlers/slogdiscard"
	"github.com/Braendie/url-shortener/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestDeleteHandler(t *testing.T) {
	testCases := []struct {
		name      string
		alias     string
		respError string
		mockError error
		code      int
	}{
		{
			name:  "valid",
			alias: "test_alias",
			code:  http.StatusNoContent,
		},
		{
			name:      "empty alias",
			alias:     "",
			respError: "invalid request",
			code:      http.StatusNotFound,
		},
		{
			name:      "not found",
			alias:     "test_alias",
			respError: "not found",
			mockError: storage.ErrURLNotFound,
			code:      http.StatusNotFound,
		},
		{
			name:      "storage error",
			alias:     "test_alias",
			respError: "internal error",
			mockError: errors.New("some error"),
			code:      http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			urlDeleterMock := mocks.NewURLDeleter(t)

			if tc.alias != "" {
				urlDeleterMock.On("DeleteURL", mock.AnythingOfType("string")).
					Return(tc.mockError).
					Once()
			}
			handler := delete.New(slogdiscard.NewDiscardLogger(), urlDeleterMock)
			r := chi.NewRouter()
			r.Delete("/{alias}", handler)

			req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("/%s", tc.alias), nil)
			require.NoError(t, err)
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			assert.Equal(t, tc.code, rr.Code)
		})
	}
}
