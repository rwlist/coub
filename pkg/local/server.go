package local

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/go-chi/chi/v5"
	"github.com/rwlist/coub/pkg/conf"
	"gorm.io/gorm"
	"io"
	"math/rand"
	"net/http"
	"path/filepath"
	"strings"
)

type Server struct {
	s3  *s3.S3
	db  *gorm.DB
	cfg *conf.App
}

func NewServer(s3 *s3.S3, db *gorm.DB, cfg *conf.App) *Server {
	return &Server{
		s3:  s3,
		db:  db,
		cfg: cfg,
	}
}

func (s *Server) Router() *chi.Mux {
	r := chi.NewRouter()
	r.Get("/file/{filename}", s.handleFile)
	r.Get("/profile", s.handleProfile)
	r.Get("/profile/{index:[0-9]+}", s.handleProfile)
	r.Get("/profile_{filter}/{index:[0-9]+}", s.handleProfile)
	return r
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
		number = rand.Intn(int(allCount))
	}

	var profileCoub ProfileCoub
	err := filter.Order("published_at ASC").Limit(1).Offset(number).Find(&profileCoub).Error
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	prev := number - 1
	if prev < 0 {
		prev = int(allCount) - 1
	}
	next := number + 1
	if next >= int(allCount) {
		next = 0
	}
	random := rand.Intn(int(allCount))

	cleanURL := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	urlFirstElem := "profile"
	if len(cleanURL) > 0 {
		urlFirstElem = cleanURL[0]
	}

	w.Header().Set("Content-Type", "text/html")
	_, _ = fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<title>Profile #%v</title>
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
		number,
		profileCoub.CoubID,
		profileCoub.CoubID,
		urlFirstElem,
		prev,
		urlFirstElem,
		random,
		urlFirstElem,
		next,
	)
}
