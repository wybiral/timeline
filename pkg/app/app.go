package app

import (
	"encoding/binary"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/boltdb/bolt"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/wybiral/timeline/pkg/fanout"
)

var updatesBucket = []byte("updates")

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type App struct {
	DB        *bolt.DB
	Router    *mux.Router
	Templates *template.Template
	Clients   *fanout.Fanout
}

func NewApp() (*App, error) {
	a := &App{}
	opts := &bolt.Options{Timeout: 1 * time.Second}
	db, err := bolt.Open("timeline.db", 0666, opts)
	if err != nil {
		return nil, err
	}
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(updatesBucket)
		return err
	})
	if err != nil {
		return nil, err
	}
	a.DB = db
	r := mux.NewRouter().StrictSlash(true)
	// static file handler
	fsHandler := http.StripPrefix(
		"/static/",
		http.FileServer(http.Dir("./static/")),
	)
	r.PathPrefix("/static/").Handler(fsHandler).Methods("GET")
	r.HandleFunc("/", a.indexGet).Methods("GET")
	r.HandleFunc("/live", a.liveGet).Methods("GET")
	r.HandleFunc("/input", a.inputGet).Methods("GET")
	a.Router = r
	t, err := template.ParseGlob("templates/*")
	if err != nil {
		return nil, err
	}
	a.Templates = t
	a.Clients = fanout.NewFanout()
	return a, nil
}

func (a *App) Run() error {
	log.Println("Serving at http://localhost:8888")
	return http.ListenAndServe("0.0.0.0:8888", a.Router)
}

func (a *App) indexGet(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	updates := make([]string, 0)
	err := a.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(updatesBucket)
		c := b.Cursor()
		i := 0
		k, v := c.Last()
		for k != nil && i < 10 {
			k, v = c.Prev()
			i++
		}
		k, v = c.Next()
		for k != nil {
			fmt.Println(string(v))
			updates = append(updates, string(v))
			k, v = c.Next()
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	a.Templates.ExecuteTemplate(w, "index.html", updates)
}

func (a *App) liveGet(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer ws.Close()
	ch := make(chan []byte, 10)
	a.Clients.Add(ch)
	defer a.Clients.Remove(ch)
	for chunk := range ch {
		err := ws.WriteMessage(websocket.TextMessage, chunk)
		if err != nil {
			return
		}
	}
}

func (a *App) inputGet(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer ws.Close()
	throttle := time.NewTicker(time.Second / 25)
	defer throttle.Stop()
	for {
		_, data, err := ws.ReadMessage()
		if err != nil {
			return
		}
		err = a.DB.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket(updatesBucket)
			id, err := b.NextSequence()
			if err != nil {
				return err
			}
			return b.Put(itob(id), data)
		})
		if err != nil {
			return
		}
		a.Clients.Send(data)
		<-throttle.C
	}
}

// convert uint64 to big engian bytes (for IDs)
func itob(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, v)
	return b
}
