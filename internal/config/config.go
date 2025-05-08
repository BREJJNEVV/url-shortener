package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

// Config структура - каждое поле этого конфига соответствует полю config.yaml
// type Config struct {
// 	Env         string `yaml:"env" env-default:"local" env-required:"red"`
// 	StoragePath string `yaml:"storage_path" env-required:"true"`
// 	HTTPServer  `yaml:"http-server"`
// }
// type HTTPServer struct {
// 	Address     string        `yaml:"address" env-default:"localhost:8080"`
// 	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
// 	IdleTimeout time.Duration `yaml:"idleTimeout" env-default:"60s"`
// }

type Config struct {
	Env         string `yaml:"env" env-default:"local"`
	StoragePath string `yaml:"storage_path" env-required:"true"`
	HTTPServer  `yaml:"http_server"`
}

type HTTPServer struct {
	Address     string        `yaml:"address" env-default:"localhost:8080"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"` // было "idler_timeout"
	User        string        `yaml:"user" env-required:"true"`
	Password    string        `yaml:"password" env-required:"true" env:"HTTP_SERVER_PASSWORD"` // секретная штука для парлоя
}

// MustLoad функция, которая прочитает файл с конфигом, создаст и заполнит из config.yaml
// и упадёт, если конфиг не загрузился
func MustLoad() *Config {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH is not set")
	}
	// log.Printf("CONFIG_PATH = %s")
	//check if file exist
	if _, err := os.Stat(configPath); os.IsNotExist(err) { //если это именно ошибка isNotExist, то:
		log.Fatalf("Config file does not exist: %s", configPath)
	}
	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}
	return &cfg

}
