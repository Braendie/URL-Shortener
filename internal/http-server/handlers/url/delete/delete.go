package delete

import (
	"errors"
	"log/slog"
	"net/http"

	resp "github.com/Braendie/url-shortener/internal/lib/api/response"
	"github.com/Braendie/url-shortener/internal/lib/logger/sl"
	"github.com/Braendie/url-shortener/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

//go:generate go run github.com/vektra/mockery/v2@v2.53.4 --name=URLDeleter
type URLDeleter interface {
	DeleteURL(alias string) error
}

func New(log *slog.Logger, urlDeleter URLDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.delete.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		alias := chi.URLParam(r, "alias")
		if alias == "" {
			log.Info("alias is empty")

			render.JSON(w, r, resp.Error("invalid request"))

			return
		}

		err := urlDeleter.DeleteURL(alias)
		if err != nil {
			if errors.Is(err, storage.ErrURLNotFound) {
				log.Info("not found")

				render.JSON(w, r, resp.Error("not found"))
			} else {
				log.Error("failed to delete url", sl.Err(err))

				render.JSON(w, r, resp.Error("internal error"))
			}
			return
		}

		log.Info("alias deleted", slog.String("alias", alias))

		render.NoContent(w, r)
	}
}
