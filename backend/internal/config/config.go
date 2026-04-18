package config

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Env        string `env:"APP_ENV"      env-default:"dev"`
	Database   struct {
		Server   string `env:"DB_SERVER"`
		Port     int    `env:"DB_PORT"     env-default:"1433"`
		User     string `env:"DB_USER"`
		Password string `env:"DB_PASSWORD"`
		Name     string `env:"DB_NAME"`
	}
	HttpServer struct {
		Address string `env:"HTTP_ADDRESS" env-default:"localhost:8000"`
	}
	JWT struct {
		Secret string `env:"JWT_SECRET"`
		Expiry string `env:"JWT_EXPIRY"  env-default:"30m"`
	}
	AzureBlob struct {
		ConnectionString string `env:"AZURE_BLOB_CONNECTION_STRING"`
		ContainerName    string `env:"AZURE_BLOB_CONTAINER"         env-default:"book-covers"`
	}
	CORS struct {
		AllowedOrigin string `env:"CORS_ORIGIN" env-default:"*"`
	}
}

func MustLoad() *Config {
	// Load .env file if present (local dev).
	// In Azure App Service env vars are set directly, so this is safely ignored there.
	_ = godotenv.Load()

	var cfg Config
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatal(err)
	}
	return &cfg
}
