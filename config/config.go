package config

import (
	"log"
	"sync"

	"github.com/ilyakaznacheev/cleanenv"
)

// Config
type Config struct {
	Server   Server
	Postgres Postgres
}

// Server config
type Server struct {
	Port         string `env:"PORT"`
	JwtSecretKey string `env:"JWT_SECRET_KEY"`
	ReadTimeout  int    `env:"READ_TIMEOUT"`
	WriteTimeout int    `env:"WRITE_TIMEOUT"`
	IdleTimeout  int    `env:"IDLE_TIMEOUT"`
}

// Postgresql config
type Postgres struct {
	PostgresqlHost     string `env:"POSTGRES_HOST"`
	PostgresqlUser     string `env:"POSTGRES_USER"`
	PostgresqlPassword string `env:"POSTGRES_PASSWORD"`
	PostgresqlDbname   string `env:"POSTGRES_DB"`
}

var (
	config *Config
	once   sync.Once
)

// Get the config file
func GetConfig() *Config {
	once.Do(func() {
		log.Println("read application configuration")
		config = &Config{}
		if err := cleanenv.ReadConfig(".env", config); err != nil {
			help, _ := cleanenv.GetDescription(config, nil)
			log.Println(help)
			log.Fatal(err)
		}
	})
	return config
}
