package main

import (
	"errors"
	"flag"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/Necroforger/dualpng"
	"github.com/nfnt/resize"

	"github.com/gorilla/mux"
)

// Flags
var (
	Port         = flag.String("p", ":80", "Server port")
	SessionLimit = flag.Int("-session-limit", 10, "Controls how many sessions can exist at time.")
)

// Session represents websocket connection information.
type Session struct {
	sync.RWMutex
	ID     string
	Img1   image.Image
	Img2   image.Image
	Result image.Image
	Gamma  int
}

// Sessions contains all the connected sessions.
var sessions = []*Session{}

func findSession(ID string) (*Session, error) {
	for _, v := range sessions {
		if v.ID == ID {
			return v, nil
		}
	}
	return nil, errors.New("Not found")
}

// ImageHandler ...
func ImageHandler(w http.ResponseWriter, r *http.Request) {
	var (
		vars    = mux.Vars(r)
		ID      = vars["id"]
		imgname = vars["imgname"]
	)
	s, err := findSession(ID)
	if err != nil {
		writeStatus(w, 404)
		return
	}

	s.RLock()
	defer s.RUnlock()

	var img image.Image

	switch imgname {
	case "img1":
		img = s.Img1
	case "img2":
		img = s.Img2
	default:
		writeStatus(w, 404)
		return
	}

	if img != nil {
		writePNG(w, img)
	} else {
		http.ServeFile(w, r, "static/images/placeholder.png")
	}
}

// ResultHandler ...
// MODES: gamma | nogamma
func ResultHandler(w http.ResponseWriter, r *http.Request) {
	var (
		vars = mux.Vars(r)
		ID   = vars["id"]
		mode = vars["mode"]
	)
	s, err := findSession(ID)
	if err != nil {
		writeStatus(w, 404)
		return
	}

	s.Lock()
	defer s.Unlock()

	if s.Result == nil {
		http.ServeFile(w, r, "static/images/placeholder.png")
		return
	}

	switch mode {
	case "nogamma":
		writePNG(w, s.Result)
	default:
		writeGAMApng(w, s.Result, s.Gamma)
	}
}

// MergeHandler handles merge requests.
func MergeHandler(w http.ResponseWriter, r *http.Request) {
	var (
		vars = mux.Vars(r)
		ID   = vars["id"]
	)

	s, err := findSession(ID)
	if err != nil {
		writeStatus(w, 404)
		return
	}

	s.Lock()
	defer s.Unlock()

	if s.Img1 == nil || s.Img2 == nil {
		writeStatus(w, http.StatusPreconditionRequired)
		log.Println("Either img1 or img2 is nil")
		return
	}

	if err := r.ParseForm(); err != nil {
		writeStatus(w, http.StatusInternalServerError)
		log.Println("Error parsing form: ", err)
		return
	}

	parseInt := func(str string) int {
		if str == "" {
			return 0
		}
		n, e := strconv.Atoi(str)
		if e != nil {
			log.Println("Error parsing integer: ", e)
			err = e
		}
		return n
	}

	r1start := parseInt(r.Form.Get("r1start"))
	r1end := parseInt(r.Form.Get("r1end"))
	r2start := parseInt(r.Form.Get("r2start"))
	r2end := parseInt(r.Form.Get("r2end"))
	gamma := parseInt(r.Form.Get("gamma"))
	width := parseInt(r.Form.Get("width"))
	height := parseInt(r.Form.Get("height"))
	if err != nil {
		writeStatus(w, http.StatusInternalServerError)
		return
	}

	s.Gamma = gamma

	img1, img2 := s.Img1, s.Img2

	if width > 0 || height > 0 {
		img1 = resize.Resize(uint(width), uint(height), img1, resize.Lanczos3)
		img2 = resize.Resize(uint(width), uint(height), img2, resize.Lanczos3)
	}

	s.Result = dualpng.MergeImages(
		dualpng.LevelImage(
			img1, uint8(r1start), uint8(r1end),
		),
		dualpng.LevelImage(
			img2, uint8(r2start), uint8(r2end),
		),
		nil,
	)

	writeStatus(w, 200)
}

// UploadHandler ...
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	var (
		vars    = mux.Vars(r)
		ID      = vars["id"]
		imgname = vars["imgname"]
	)

	if imgname != "img1" && imgname != "img2" {
		writeStatus(w, 404)
		return
	}

	s, err := findSession(ID)
	if err != nil {
		writeStatus(w, 404)
		return
	}

	if err := r.ParseMultipartForm((1 << 10) * 24); err != nil {
		log.Println("Error parsing form: ", err)
		writeStatus(w, http.StatusInternalServerError)
		return
	}

	formfile, _, err := r.FormFile("img")
	if err != nil {
		log.Println("Error retrieving form file: ", err)
		writeStatus(w, http.StatusInternalServerError)
		return
	}
	defer formfile.Close()

	img, _, err := image.Decode(formfile)
	if err != nil {
		log.Println(err)
		writeStatus(w, http.StatusInternalServerError)
		return
	}

	s.Lock()
	defer s.Unlock()

	switch imgname {
	case "img1":
		s.Img1 = img
	case "img2":
		s.Img2 = img
	}

	writePNG(w, img)
}

func main() {
	r := mux.NewRouter()

	sessions = append(sessions, &Session{
		ID: "TEST",
	})

	r.HandleFunc("/image/{id}/{imgname}", ImageHandler)
	r.HandleFunc("/result/{id}/{mode}", ResultHandler)
	r.HandleFunc("/upload/{id}/{imgname}", UploadHandler).Methods("POST")
	r.HandleFunc("/merge/{id}", MergeHandler).Methods("POST")
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("static")))

	srv := &http.Server{
		Handler:      r,
		Addr:         "127.0.0.1" + *Port,
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 10,
	}

	log.Println("Attempting to start server on port [" + *Port + "]")
	if err := srv.ListenAndServe(); err != nil {
		log.Println("error starting server: ", err)
	}
}

func writeStatus(w http.ResponseWriter, status int) {
	w.WriteHeader(status)
	fmt.Fprint(w, http.StatusText(status))
}

func writePNG(w http.ResponseWriter, img image.Image) {
	w.Header().Set("content-type", "image/png")
	png.Encode(w, img)
}

func writeGAMApng(w http.ResponseWriter, img image.Image, gAMA int) {
	w.Header().Set("content-type", "image/png")
	dualpng.Encode(w, img, uint32(gAMA))
}