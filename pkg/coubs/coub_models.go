package coubs

import (
	"encoding/json"
	"time"
)

type PageResponse struct {
	Page       int               `json:"page"`
	PerPage    int               `json:"per_page"`
	TotalPages int               `json:"total_pages"`
	Coubs      []json.RawMessage `json:"coubs"`
}

type TimelineResponse struct {
	Page       int    `json:"page"`
	PerPage    int    `json:"per_page"`
	TotalPages int    `json:"total_pages"`
	Coubs      []Coub `json:"coubs"`
}

type Blob struct {
	URL  string `json:"url"`
	Size int64  `json:"size"`
}

type Blobs map[string]json.RawMessage

type HTML5 struct {
	Video Blobs `json:"video"`
	Audio Blobs `json:"audio"`
}

type Mobile struct {
	Video string   `json:"video"`
	Audio []string `json:"audio"`
}

type Share struct {
	Default string `json:"default"`
}

type FileVersions struct {
	HTML5  HTML5  `json:"html5"`
	Mobile Mobile `json:"mobile"`
	Share  Share  `json:"share"`
}

type AudioVersions struct {
	Template string   `json:"template"`
	Versions []string `json:"versions"`
}

type ImageVersions struct {
	Template string   `json:"template"`
	Versions []string `json:"versions"`
}

type FirstFrameVersions struct {
	Template string   `json:"template"`
	Versions []string `json:"versions"`
}

type Dimensions struct {
	Big []int `json:"big"`
	Med []int `json:"med"`
}

type AvatarVersions struct {
	Template string   `json:"template"`
	Versions []string `json:"versions"`
}

type Channel struct {
	ID             int            `json:"id"`
	Permalink      string         `json:"permalink"`
	Title          string         `json:"title"`
	IFollowHim     bool           `json:"i_follow_him"`
	FollowersCount int            `json:"followers_count"`
	FollowingCount int            `json:"following_count"`
	AvatarVersions AvatarVersions `json:"avatar_versions"`
}

type Coub struct {
	Favorite             bool          `json:"favorite"`
	Recoub               bool          `json:"recoub"`
	Like                 bool          `json:"like"`
	Dislike              bool          `json:"dislike"`
	Reaction             string        `json:"reaction"`
	ID                   int           `json:"id"`
	Type                 string        `json:"type"`
	Permalink            string        `json:"permalink"`
	Title                string        `json:"title"`
	ChannelID            int           `json:"channel_id"`
	CreatedAt            time.Time     `json:"created_at"`
	UpdatedAt            time.Time     `json:"updated_at"`
	ViewsCount           int           `json:"views_count"`
	PublishedAt          time.Time     `json:"published_at"`
	FileVersions         FileVersions  `json:"file_versions"`
	AudioVersions        AudioVersions `json:"audio_versions,omitempty"`
	ImageVersions        ImageVersions `json:"image_versions"`
	AudioFileURL         string        `json:"audio_file_url"`
	Channel              Channel       `json:"channel"`
	Picture              string        `json:"picture"`
	TimelinePicture      string        `json:"timeline_picture"`
	RecoubsCount         int           `json:"recoubs_count"`
	RemixesCount         int           `json:"remixes_count"`
	LikesCount           int           `json:"likes_count"`
	DislikesCount        int           `json:"dislikes_count"`
	RawVideoThumbnailURL string        `json:"raw_video_thumbnail_url"`
	RawVideoTitle        string        `json:"raw_video_title"`
	Duration             float64       `json:"duration"`
}
