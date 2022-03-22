package local

import (
	"encoding/json"

	"github.com/rwlist/coub/pkg/coubs"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Backup struct {
	downloader *Downloader
	client     *coubs.Client
	db         *gorm.DB
	state      *SharedState
}

func NewBackup(downloader *Downloader, client *coubs.Client, db *gorm.DB, state *SharedState) *Backup {
	return &Backup{
		downloader: downloader,
		client:     client,
		db:         db,
		state:      state,
	}
}

func (b *Backup) Profile(profile string) error {
	page := 1
	for {
		b.state.DownloadingProfilePage(profile, page)

		log.WithField("page", page).Info("Fetching profile page")
		pageResponse, err := b.client.ChannelTimeline(profile, page)
		if err != nil {
			return err
		}
		if len(pageResponse.Coubs) == 0 {
			break
		}

		b.state.GotProfilePage(profile, page, pageResponse)

		for index, rawCoub := range pageResponse.Coubs {
			b.state.DownloadingCoub(profile, page, index, rawCoub)

			err := b.downloader.DownloadCoub(rawCoub)
			if err != nil {
				return err
			}

			var coub coubs.Coub
			err = json.Unmarshal(rawCoub, &coub)
			if err != nil {
				return err
			}

			// check if exists in db
			var count int64
			err = b.db.Model(&ProfileCoub{}).Where("profile = ? AND coub_id = ?", profile, coub.ID).Count(&count).Error
			if err != nil {
				return err
			}
			if count > 0 {
				continue
			}

			if err := b.db.Create(&ProfileCoub{
				Profile:     profile,
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

func (b *Backup) Likes(profile string) error {
	page := 1
	for {
		log.WithField("page", page).Info("Fetching likes page")
		pageResponse, err := b.client.Likes(page)
		if err != nil {
			return err
		}
		if len(pageResponse.Coubs) == 0 {
			break
		}

		for _, rawCoub := range pageResponse.Coubs {
			err := b.downloader.DownloadCoub(rawCoub)
			if err != nil {
				return err
			}

			var coub coubs.Coub
			err = json.Unmarshal(rawCoub, &coub)
			if err != nil {
				return err
			}

			// check if exists in db
			var count int64
			err = b.db.Model(&LikedCoub{}).Where("profile = ? AND coub_id = ?", profile, coub.ID).Count(&count).Error
			if err != nil {
				return err
			}
			if count > 0 {
				continue
			}

			if err := b.db.Create(&LikedCoub{
				Profile: profile,
				CoubID:  coub.ID,
				Info:    rawCoub,
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
