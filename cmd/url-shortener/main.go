package main

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"url-shortener/internal/config"
	"url-shortener/internal/http-server/handlers/delete"
	"url-shortener/internal/http-server/handlers/redirect"
	"url-shortener/internal/http-server/handlers/url/save"
	mwlogger "url-shortener/internal/http-server/middleware/logger"
	"url-shortener/internal/lib/logger/handlers/slogpretty"
	"url-shortener/internal/lib/logger/sl"
	"url-shortener/internal/storage/sqlite"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/joho/godotenv"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	fmt.Println("start")
	err := godotenv.Load() // проверка на инициализацию окружения
	if err != nil {
		log.Fatalf("Error loading .env filewww")
	}
	cfg := config.MustLoad()    // Создаём конфиг
	log := setupLogger(cfg.Env) // Создаём логгер

	fmt.Println(cfg)

	log.Info("starting url-shortener", slog.String("env", cfg.Env))
	log.Debug("debug messages are enabled")

	storage, err := sqlite.New(cfg.StoragePath) // кладём конфиг сюда, чтобы бд могла обратиться
	if err != nil {
		log.Error("failed to init storage", sl.Err(err))
		os.Exit(1)
	}

	storage.SaveURL("https://ya.ru/?npr=1&utm_referrer=https%3A%2F%2Fyandex.ru%2F", "ya2ndex")

	// storage.DeleteURL("test1")

	router := chi.NewRouter()
	log.Info("testing log")
	log.Info("testing log", slog.String("env", cfg.Env))

	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)    //логгирует все входящие запросы
	router.Use(mwlogger.New(log))    // Наш логгер
	router.Use(middleware.Recoverer) // Не даёт падать всему приложению при панике
	router.Use(middleware.URLFormat) // Красивые url (можно переименовывать url)
	_ = storage

	router.Route("/url", func(r chi.Router) {
		r.Use(middleware.BasicAuth("url-shortener", map[string]string{
			cfg.HTTPServer.User: cfg.HTTPServer.Password,
		}))
		//r - роутер
		r.Post("/", save.New(log, storage)) //отправляем запрос на сохраниние
		r.Delete("/{alias}", delete.New(log, storage))
	})
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/index.html")
	})
	router.Get("/{alias}", redirect.New(log, storage)) // первый параметр (alias) говорит нам по какому адрессу мы перейдём

	//storage.DeleteURL("ya2ndex")
	log.Info("starting server", slog.String("address", cfg.Address))

	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout, //Время на то, чтобы мы успели прочитать запрос
		WriteTimeout: cfg.HTTPServer.Timeout, //Время на то, чтобы мы успели написать ответ
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Error("failed to start server", sl.Err(err))
	}
	log.Error("server stopped")
}
func setupLogger(env string) *slog.Logger {
	var log *slog.Logger
	switch env {
	case envLocal:
		log = setupPrettySlog()
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	}
	return log
}
func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}
