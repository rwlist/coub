package coub

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/rwlist/coub/pkg/conf"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type ProfileCoub struct {
	gorm.Model
	CoubID      int       `gorm:"not null"`
	PublishedAt time.Time `gorm:"not null"`
	Info        []byte    `gorm:"type:jsonb;not null"`
}

type Downloader struct {
	client *Client
	s3     *s3.S3
	db     *gorm.DB
	cfg    *conf.App
}

func NewDownloader(client *Client, s3cli *s3.S3, db *gorm.DB, cfg *conf.App) *Downloader {
	return &Downloader{
		client: client,
		s3:     s3cli,
		db:     db,
		cfg:    cfg,
	}
}

func (d *Downloader) AutoMigrate() error {
	return d.db.AutoMigrate(&ProfileCoub{})
}

func (d *Downloader) DownloadProfile(profile string) error {
	page := 1
	for {
		log.WithField("page", page).Info("Fetching profile page")
		pageResponse, err := d.client.ChannelTimeline(profile, page)
		if err != nil {
			return err
		}
		if len(pageResponse.Coubs) == 0 {
			break
		}

		for _, rawCoub := range pageResponse.Coubs {
			var coub Coub
			err := json.Unmarshal(rawCoub, &coub)
			if err != nil {
				return err
			}

			// check if exists in db
			var count int64
			err = d.db.Model(&ProfileCoub{}).Where("coub_id = ?", coub.ID).Count(&count).Error
			if err != nil {
				return err
			}
			if count > 0 {
				continue
			}

			err = d.downloadCoub(&coub)
			if err != nil {
				return err
			}

			if err := d.db.Create(&ProfileCoub{
				CoubID:      coub.ID,
				PublishedAt: coub.PublishedAt,
				Info:        rawCoub,
			}).Error; err != nil {
				return err
			}
		}

		page++
		if page > pageResponse.TotalPages {
			break
		}
	}

	return nil
}

func (d *Downloader) downloadCoub(coub *Coub) error {
	log.WithField("coub_id", coub.ID).Info("Downloading coub")

	videoURL, err := bestURL(coub.FileVersions.HTML5.Video)
	if err != nil {
		return err
	}

	audioURL, err := bestURL(coub.FileVersions.HTML5.Audio)
	if err != nil {
		return err
	}

	videoKey := fmt.Sprintf("%d_video.mp4", coub.ID)
	audioKey := fmt.Sprintf("%d_audio.mp3", coub.ID)

	if err := d.upload(videoURL, videoKey); err != nil {
		return err
	}

	if err := d.upload(audioURL, audioKey); err != nil {
		return err
	}

	return nil
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

func bestURL(video Blobs) (string, error) {
	res := ""
	size := int64(0)

	for _, rawBlob := range video {
		var blob Blob
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
