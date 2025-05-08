package save

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	resp "url-shortener/internal/lib/api/response"
	"url-shortener/internal/lib/logger/sl"
	"url-shortener/internal/lib/random"
	"url-shortener/internal/storage"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

// const aliasLenght = 5

// //go:generate go run github.com/vektra/mockery/v2@v2.53.3 --name=URLSaver --output=mocks

// type Request struct {
// 	URL   string `json:"url" validate:"required,url"`
// 	Alias string `json:"alias,omitempty"`
// }
// type Response struct {
// 	resp.Response
// 	Alias string `json:"alias,omitempty"`
// }
// type URLSaver interface {
// 	SaveURL(urlToSave string, alias string) (int64, error)
// 	//AliasNotExists(alias string) bool
// }

// func New(log *slog.Logger, urlSaver URLSaver) http.HandlerFunc { //КОНЦЕПЦИЯ: Чтобы чтот использовать в дочерних функуиях,
// 	// просто передвай ему экземпляр например слоггера, пусть с ней работает
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		const op = "handlers.url.save.New"

// 		log = log.With(
// 			slog.String("op", op), // Просто вывод в логе того, что происходит
// 			slog.String("request_id", middleware.GetReqID(r.Context())),
// 		)

// 		//Создаём объект запросов, который будем анмаршиллить(переводится как отмена сделки)
// 		var req Request

// 		// помогает распарсить наш запрос
// 		err := render.DecodeJSON(r.Body, &req)
// 		//«Распарсить» — это провести парсинг, то есть автоматизированный сбор и анализ данных с интернет-страниц помощью скриптов, то есть парсеров
// 		if err != nil {
// 			// Выводим сообщение об ошибке и её тип
// 			log.Error("failed to decode request body", sl.Err(err))
// 			w.WriteHeader(http.StatusBadRequest) // 400
// 			//Рендерим JSON с ответом нашему клиенту
// 			render.JSON(w, r, resp.Error("failed to decode request body"))

// 			return
// 		}
// 		// сразу с сообщением об успешном сохранении, и запрос, который мы декодировали в объект REQUEST
// 		log.Info("request body decoded", slog.Any("request", req))
// 		//Валидация данных — это процесс проверки данных

// 		if err := validator.New().Struct(req); err != nil {
// 			validateErr := err.(validator.ValidationErrors)

// 			log.Error("invalid request", sl.Err(err)) //или "validation failed"

// 			render.JSON(w, r, resp.ValidationError(validateErr))

// 			return
// 		}

// 		alias := req.Alias
// 		if alias == "" {
// 			alias = random.NewRandomString(aliasLenght)

// 			// for !urlSaver.AliasNotExists(alias) {
// 			// 	alias = random.NewRandomString(aliasLenght)
// 			// }
// 		}
// 		id, err := urlSaver.SaveURL(req.URL, alias)
// 		if errors.Is(err, storage.ErrURLExists) {
// 			log.Info("url already exists", slog.String("url", req.URL))
// 			w.WriteHeader(http.StatusConflict) // 409
// 			render.JSON(w, r, Response{
// 				Response: resp.Error("url already exists"),
// 				Alias:    alias,
// 			})
// 			return
// 		}

// 		if err != nil {
// 			log.Error("failed to save url", sl.Err(err))
// 			w.WriteHeader(http.StatusInternalServerError)
// 			render.JSON(w, r, resp.Error("failed to save url"))
// 			return

// 		}
// 		log.Info("url added", slog.Int64("id", id))
// 		w.WriteHeader(http.StatusCreated) // 201
// 		ResponseOK(w, r, alias)
// 	}
// }

// func ResponseOK(w http.ResponseWriter, r *http.Request, alias string) {
// 	render.JSON(w, r, Response{
// 		Response: resp.OK(),
// 		Alias:    alias,
// 	})
// }

type Request struct {
	URL   string `json:"url" validate:"required,url"`
	Alias string `json:"alias,omitempty"`
}

type Response struct {
	resp.Response
	Alias string `json:"alias,omitempty"`
}

// TODO: move to config if needed
const aliasLength = 6

//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=URLSaver
type URLSaver interface {
	SaveURL(urlToSave string, alias string) (int64, error)
}

func New(log *slog.Logger, urlSaver URLSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.save.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		err := render.DecodeJSON(r.Body, &req)
		if errors.Is(err, io.EOF) {
			// Такую ошибку встретим, если получили запрос с пустым телом.
			// Обработаем её отдельно
			log.Error("request body is empty")

			render.JSON(w, r, resp.Error("empty request"))

			return
		}
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to decode request"))

			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)

			log.Error("invalid request", sl.Err(err))

			render.JSON(w, r, resp.ValidationError(validateErr))

			return
		}

		alias := req.Alias
		if alias == "" {
			alias = random.NewRandomString(aliasLength)
		}

		id, err := urlSaver.SaveURL(req.URL, alias)
		if errors.Is(err, storage.ErrURLExists) {
			log.Info("url already exists", slog.String("url", req.URL))

			render.JSON(w, r, resp.Error("url already exists"))

			return
		}
		if err != nil {
			log.Error("failed to add url", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to add url"))

			return
		}

		log.Info("url added", slog.Int64("id", id))

		responseOK(w, r, alias)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, alias string) {
	render.JSON(w, r, Response{
		Response: resp.OK(),
		Alias:    alias,
	})
}
