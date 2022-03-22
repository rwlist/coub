package local

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/davecgh/go-spew/spew"
	"github.com/rwlist/coub/pkg/conf"
	"github.com/rwlist/coub/pkg/coubs"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Downloader struct {
	client *coubs.Client
	s3     *s3.S3
	db     *gorm.DB
	cfg    *conf.App
}

func NewDownloader(client *coubs.Client, s3cli *s3.S3, db *gorm.DB, cfg *conf.App) *Downloader {
	return &Downloader{
		client: client,
		s3:     s3cli,
		db:     db,
		cfg:    cfg,
	}
}

func (d *Downloader) DownloadCoub(rawCoub []byte) error {
	var coub coubs.Coub
	err := json.Unmarshal(rawCoub, &coub)
	if err != nil {
		spew.Dump(rawCoub)
		return err
	}

	var count int64
	err = d.db.Model(&SavedCoub{}).Where("coub_id = ?", coub.ID).Count(&count).Error
	if err != nil {
		return err
	}
	if count > 0 {
		log.WithField("coub_id", coub.ID).Info("coub was downloaded before")
		return nil
	}

	log.WithField("coub_id", coub.ID).Info("Downloading coub")

	videoURL, err := bestURL(coub.FileVersions.HTML5.Video)
	if err != nil {
		return err
	}
	videoKey := fmt.Sprintf("%d_video.mp4", coub.ID)
	err = d.upload(videoURL, videoKey)
	if err != nil {
		return err
	}

	var noAudio bool

	audioURL, err := bestURL(coub.FileVersions.HTML5.Audio)
	if err != nil {
		noAudio = true
		log.WithField("coub_id", coub.ID).Info("coub has no audio")
	} else {
		audioKey := fmt.Sprintf("%d_audio.mp3", coub.ID)
		if err := d.upload(audioURL, audioKey); err != nil {
			return err
		}
	}

	return d.db.Create(&SavedCoub{
		CoubID:  coub.ID,
		Info:    rawCoub,
		NoAudio: noAudio,
	}).Error
}

func (d *Downloader) upload(url, key string) error {
	resp, err := http.Get(url) //nolint
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var body []byte
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	_, err = d.s3.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(d.cfg.S3Bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(body),
	})
	return err
}

func bestURL(video coubs.Blobs) (string, error) {
	res := ""
	size := int64(0)

	for _, rawBlob := range video {
		var blob coubs.Blob
		err := json.Unmarshal(rawBlob, &blob)
		if err != nil {
			continue
		}

		if blob.Size > size {
			size = blob.Size
			res = blob.URL
		}
	}

	if res == "" {
		return "", errors.New("no suitable media found")
	}
	return res, nil
}
