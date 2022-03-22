//nolint:dupl
package local

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/davecgh/go-spew/spew"
	"github.com/go-chi/chi/v5"
	"github.com/rwlist/coub/pkg/conf"
	"gorm.io/gorm"
)

type Server struct {
	s3    *s3.S3
	db    *gorm.DB
	cfg   *conf.App
	state *SharedState
}

func NewServer(sss *s3.S3, db *gorm.DB, cfg *conf.App, state *SharedState) *Server {
	return &Server{
		s3:    sss,
		db:    db,
		cfg:   cfg,
		state: state,
	}
}

func (s *Server) Router() *chi.Mux {
	r := chi.NewRouter()
	r.Get("/file/{filename}", s.handleFile)

	r.Get("/profile", s.handleProfile)
	r.Get("/profile/{index:[0-9]+}", s.handleProfile)
	r.Get("/profile_{filter}/{index:[0-9]+}", s.handleProfile)

	r.Get("/liked", s.handleLiked)
	r.Get("/liked/{index:[0-9]+}", s.handleLiked)
	r.Get("/liked{filter}/{index:[0-9]+}", s.handleLiked)

	r.Get("/state", s.handleState)

	return r
}

func (s *Server) handleState(w http.ResponseWriter, r *http.Request) {
	state := s.state.Get()
	spew.Fdump(w, state)
}

func (s *Server) handleFile(w http.ResponseWriter, r *http.Request) {
	filename := filepath.Base(r.URL.Path)
	res, err := s.s3.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(s.cfg.S3Bucket),
		Key:    &filename,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if res.ContentType != nil {
		w.Header().Set("Content-Type", *res.ContentType)
	}
	if res.ContentLength != nil {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", *res.ContentLength))
	}

	_, _ = io.Copy(w, res.Body)
}

func (s *Server) handleProfile(w http.ResponseWriter, r *http.Request) {
	filter := s.db.Model(&ProfileCoub{})

	filterParam := chi.URLParam(r, "filter")
	if filterParam != "" {
		filter = filter.Where("profile = ?", filterParam)
	}

	var allCount int64
	if err := filter.Count(&allCount).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	numberRaw := chi.URLParam(r, "index")
	var number int
	if _, err := fmt.Sscanf(numberRaw, "%d", &number); err != nil {
		// take a random coub
		number = rand.Intn(int(allCount)) //nolint:gosec
	}

	var profileCoub ProfileCoub
	err := filter.Order("published_at ASC").Limit(1).Offset(number).Find(&profileCoub).Error
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	z0rViewer{
		CoubID:     profileCoub.CoubID,
		Number:     number,
		AllCount:   int(allCount),
		DefaultURL: "profile",
		Header:     "Profile",
	}.Render(w, r)
}

func (s *Server) handleLiked(w http.ResponseWriter, r *http.Request) {
	filter := s.db.Model(&LikedCoub{})

	filterParam := chi.URLParam(r, "filter")
	if filterParam != "" {
		filter = filter.Where("profile = ?", filterParam)
	}

	var allCount int64
	if err := filter.Count(&allCount).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	numberRaw := chi.URLParam(r, "index")
	var number int
	if _, err := fmt.Sscanf(numberRaw, "%d", &number); err != nil {
		// take a random coub
		number = rand.Intn(int(allCount)) //nolint:gosec
	}

	var likedCoub LikedCoub
	err := filter.Order("id ASC").Limit(1).Offset(number).Find(&likedCoub).Error
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	z0rViewer{
		CoubID:     likedCoub.CoubID,
		Number:     number,
		AllCount:   int(allCount),
		DefaultURL: "liked",
		Header:     "Liked",
	}.Render(w, r)
}

type z0rViewer struct {
	CoubID     int
	Number     int
	AllCount   int
	DefaultURL string
	Header     string
}

func (z z0rViewer) Render(w http.ResponseWriter, r *http.Request) {
	prev := z.Number - 1
	if prev < 0 {
		prev = z.AllCount - 1
	}
	next := z.Number + 1
	if next >= z.AllCount {
		next = 0
	}
	random := rand.Intn(z.AllCount) //nolint:gosec

	cleanURL := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	urlFirstElem := z.DefaultURL
	if len(cleanURL) > 0 {
		urlFirstElem = cleanURL[0]
	}

	w.Header().Set("Content-Type", "text/html")
	_, _ = fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<title>%s #%v</title>
</head>
<body>
<center>
<style>
body {
  font-family: sans-serif;
  font-size: 16px;
}
.viewer__video {
  max-width: 70%%;
  max-height: 70%%;
}
</style>
<script>
let hasPlayed = false;
function handleFirstPlay(event) {
  if(hasPlayed === false) {
    hasPlayed = true;
    let vid = event.target;
    vid.onplay = null;

    document.querySelector("audio").play();
  }
}

</script>
<video class="viewer__video" loop="loop" controls autoplay preload="auto" src="/file/%v_video.mp4" onplay="handleFirstPlay(event)"></video>
<br/>
<audio preload="auto" controls loop="loop" src="/file/%v_audio.mp3"></audio>
<p>
<a href="/%v/%v">Prev</a>
<a href="/%v/%v">Random</a>
<a href="/%v/%v">Next</a>
</p>
</center>
</body>
</html>
`,
		z.Header,
		z.Number,
		z.CoubID,
		z.CoubID,
		urlFirstElem,
		prev,
		urlFirstElem,
		random,
		urlFirstElem,
		next,
	)
}
