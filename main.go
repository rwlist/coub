package main

import (
	"github.com/rwlist/coub/pkg/coubs"
	"github.com/rwlist/coub/pkg/local"
	"math/rand"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/davecgh/go-spew/spew"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/rwlist/coub/pkg/conf"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	log.SetFormatter(&log.JSONFormatter{})
	log.SetReportCaller(true)
	log.SetLevel(log.DebugLevel)

	cfg, err := conf.ParseEnv()
	if err != nil {
		log.WithError(err).Fatal("failed to parse config from env")
	}

	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		err := http.ListenAndServe(cfg.PrometheusBind, mux) //nolint:govet
		if err != nil && err != http.ErrServerClosed {
			log.WithError(err).Fatal("prometheus server error")
		}
	}()

	db, err := gorm.Open(postgres.Open(cfg.PostgresDSN), &gorm.Config{})
	if err != nil {
		log.WithError(err).Fatal("failed to connect to postgres")
	}
	db = db.Debug()

	err = local.AutoMigrate(db)
	if err != nil {
		log.WithError(err).Fatal("failed to migrate tables")
	}

	cookies := coubs.NewCookies(db)
	err = cookies.AutoMigrate()
	if err != nil {
		log.WithError(err).Fatal("failed to migrate kv table")
	}

	c, err := cookies.Get()
	spew.Dump(c, err)

	// Configure to use MinIO Server
	s3Config := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(cfg.S3AccessKey, cfg.S3SecretKey, ""),
		Endpoint:         aws.String(cfg.S3Endpoint),
		Region:           aws.String(cfg.S3Region),
		DisableSSL:       aws.Bool(false),
		S3ForcePathStyle: aws.Bool(true),
	}
	newSession, err := session.NewSession(s3Config)
	if err != nil {
		log.WithError(err).Fatal("failed to create new session")
	}

	s3Client := s3.New(newSession)

	const enableBackup = false

	cli := coubs.NewClient(cookies)
	downloader := local.NewDownloader(cli, s3Client, db, cfg)
	backup := local.NewBackup(downloader, cli, db)

	if enableBackup {
		err = backup.Profile(cfg.CoubUsername)
		if err != nil {
			log.WithError(err).Fatal("failed to download profile")
		}
	}

	log.Info("done")

	server := local.NewServer(s3Client, db, cfg)
	r := server.Router()
	err = http.ListenAndServe(cfg.BindHTTP, r)
	if err != nil {
		log.WithError(err).Fatal("http server error")
	}
}
