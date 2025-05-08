package delete

import (
	"errors"
	"log/slog"
	"net/http"
	"url-shortener/internal/lib/logger/sl"
	"url-shortener/internal/storage"

	resp "url-shortener/internal/lib/api/response"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

type URLDeletter interface {
	DeleteURL(alias string) error
}

// func New(log *slog.Logger, urlDeletter URLDeletter) http.HandlerFunc { //может ничего не возвращать?
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		const op = "handlers.url.delete.New"

// 		log := log.With(
// 			slog.String("op", op),
// 			slog.String("method", r.Method),
// 			slog.String("path", r.URL.Path),
// 			slog.String("request_id", middleware.GetReqID(r.Context())),
// 		)
// 		log.Debug("Route parameters",
// 			slog.Any("params", chi.RouteContext(r.Context()).URLParams))

// 		alias := chi.URLParam(r, "alias")
// 		log = log.With(slog.String("alias", alias))
// 		if alias == "" {
// 			log.Error("Empty alias parameter")
// 			w.WriteHeader(http.StatusBadRequest)
// 			render.JSON(w, r, resp.Error("alias parameter is required"))
// 			return
// 		}

// 		err := urlDeletter.DeleteURL(alias, log)
// 		if errors.Is(err, storage.ErrURLNotFound) {
// 			log.Info("url not found")
// 			w.WriteHeader(http.StatusNotFound)
// 			render.JSON(w, r, resp.Error("not found"))
// 			return
// 		}
// 		if err != nil {
// 			log.Error("failed to get url", sl.Err(err))
// 			w.WriteHeader(http.StatusInternalServerError)
// 			render.JSON(w, r, resp.Error("internal error"))

// 			return
// 		}
// 		log.Info("url is deleted")
// 		w.WriteHeader(http.StatusNoContent)

//		}
//	}
////////////////////////////////////////////////////////////////////////////////////////
// func New(log *slog.Logger, urlDeletter URLDeletter) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		const op = "handlers.url.delete.New"

// 		logger := log.With(
// 			slog.String("op", op),
// 			slog.String("request_id", middleware.GetReqID(r.Context())),
// 		)

// 		logger.Debug("Request URL", slog.String("url", r.URL.String()))

// 		alias := chi.URLParam(r, "alias")
// 		if alias == "" {
// 			logger.Error("empty alias")
// 			w.WriteHeader(http.StatusBadRequest)
// 			render.JSON(w, r, resp.Error("alias required"))
// 			return
// 		}

func New(log *slog.Logger, urlDeletter URLDeletter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.redirect.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		alias := chi.URLParam(r, "alias")
		if alias == "" {
			log.Error("alias is empty",
				slog.Any("route_context", chi.RouteContext(r.Context())))

			render.JSON(w, r, resp.Error("not validate"))

			return
		}

		log.Debug("Calling DeleteURL in storage")
		err := urlDeletter.DeleteURL(alias)
		log.Debug("DeleteURL result", slog.Any("error", err))
		if errors.Is(err, storage.ErrURLNotFound) {
			log.Info("url not found")
			w.WriteHeader(http.StatusNotFound)
			render.JSON(w, r, resp.Error("not found"))
			return
		}
		if err != nil {
			log.Error("failed to get url", sl.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("internal error"))

			return
		}
		log.Info("url is deleted")
		w.WriteHeader(http.StatusNoContent)

	}
}
