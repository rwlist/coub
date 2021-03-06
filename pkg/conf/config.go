package conf

import (
	"github.com/caarlos0/env/v6"
)

type App struct {
	PrometheusBind string   `env:"PROMETHEUS_BIND" envDefault:":2112"`
	PostgresDSN    string   `env:"PG_DSN"`
	S3Endpoint     string   `env:"S3_ENDPOINT"`
	S3Region       string   `env:"S3_REGION"`
	S3AccessKey    string   `env:"S3_ACCESS_KEY_ID"`
	S3SecretKey    string   `env:"S3_SECRET_ACCESS_KEY"`
	S3Bucket       string   `env:"S3_BUCKET"`
	CoubUsername   string   `env:"COUB_USERNAME"`
	BindHTTP       string   `env:"BIND_HTTP" envDefault:":8080"`
	EnableBackup   bool     `env:"ENABLE_BACKUP" envDefault:"false"`
	BackupProfiles []string `env:"BACKUP_PROFILES" envSeparator:","`
}

func ParseEnv() (*App, error) {
	cfg := App{}
	err := env.Parse(&cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
