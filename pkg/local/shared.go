package local

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/rwlist/coub/pkg/coubs"
	log "github.com/sirupsen/logrus"
)

type State struct {
	Profile    string
	Page       int
	PerPage    int
	TotalPages int

	IndexInPage   int
	Favorite      bool                `json:"favorite"`
	Recoub        bool                `json:"recoub"`
	Like          bool                `json:"like"`
	Dislike       bool                `json:"dislike"`
	ID            int                 `json:"id"`
	Type          string              `json:"type"`
	Permalink     string              `json:"permalink"`
	Title         string              `json:"title"`
	ViewsCount    int                 `json:"views_count"`
	PublishedAt   time.Time           `json:"published_at"`
	FileVersions  coubs.FileVersions  `json:"file_versions"`
	AudioVersions coubs.AudioVersions `json:"audio_versions,omitempty"`
	Channel       string              `json:"channel"`
	LikesCount    int                 `json:"likes_count"`
	DislikesCount int                 `json:"dislikes_count"`
}

type SharedState struct {
	State
	mux sync.Mutex
}

func NewSharedState() *SharedState {
	return &SharedState{}
}

func (s *SharedState) Get() State {
	s.mux.Lock()
	defer s.mux.Unlock()
	return s.State
}

func (s *SharedState) DownloadingProfilePage(profile string, page int) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.Profile = profile
	s.Page = page
}

func (s *SharedState) GotProfilePage(profile string, page int, response *coubs.PageResponse) {
	s.mux.Lock()
	defer s.mux.Unlock()

	s.Profile = profile
	s.Page = page
	s.PerPage = response.PerPage
	s.TotalPages = response.TotalPages
	s.IndexInPage = -1
}

func (s *SharedState) DownloadingCoub(profile string, page, index int, rawCoub json.RawMessage) {
	var coub coubs.Coub
	err := json.Unmarshal(rawCoub, &coub)
	if err != nil {
		log.WithError(err).Error("Failed to unmarshal coub")
		return
	}

	s.mux.Lock()
	defer s.mux.Unlock()
	s.IndexInPage = index
	s.Favorite = coub.Favorite
	s.Recoub = coub.Recoub
	s.Like = coub.Like
	s.Dislike = coub.Dislike
	s.ID = coub.ID
	s.Type = coub.Type
	s.Permalink = coub.Permalink
	s.Title = coub.Title
	s.ViewsCount = coub.ViewsCount
	s.PublishedAt = coub.PublishedAt
	s.FileVersions = coub.FileVersions
	s.AudioVersions = coub.AudioVersions
	s.Channel = coub.Channel.Title
	s.LikesCount = coub.LikesCount
	s.DislikesCount = coub.DislikesCount
}
