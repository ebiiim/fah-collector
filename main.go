package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
)

type collector struct {
	M          sync.Map
	CachedJSON string
	LastCached time.Time
	CacheTTL   time.Duration
	KeyTTL     time.Duration
	PK         string
}

func newCollector(cacheTTL time.Duration, keyTTL time.Duration, pk string) *collector {
	c := &collector{
		CacheTTL: cacheTTL,
		KeyTTL:   keyTTL,
		PK:       pk,
	}
	return c
}

type data struct {
	jsonData string
	ts       time.Time
}

func (c *collector) Post(w http.ResponseWriter, r *http.Request) {
	p, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// parse ident
	l := []map[string]interface{}{}
	if err := json.NewDecoder(bytes.NewBuffer(p)).Decode(&l); err != nil {
		log.Printf("Post: err=%v s=%s", err, string(p))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	for _, v := range l {
		ident, ok := v[c.PK]
		if !ok {
			log.Printf("no value found c.PK=%s v=%v", c.PK, v)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		p, err := json.Marshal(v)
		if !ok {
			log.Printf("could not marshal err=%v v=%v", err, v)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		c.M.Store(ident, data{jsonData: string(p), ts: time.Now()})
	}
	w.WriteHeader(http.StatusOK)
}

func (c *collector) GetAll(w http.ResponseWriter, r *http.Request) {
	if c.LastCached.Add(c.CacheTTL).Before(time.Now()) {
		log.Println("INFO: creating JSON")
		c.CachedJSON = "{"
		c.M.Range(func(key, value interface{}) bool {
			d, ok := value.(data)
			if !ok {
				log.Println("GetAll: invalid data type")
				return true
			}
			if d.ts.Add(c.KeyTTL).Before(time.Now()) {
				c.M.Delete(key) // old data
				return true
			}
			c.CachedJSON += fmt.Sprintf("\"%s\": %s,", key, d.jsonData)
			return true
		})
		if len(c.CachedJSON) == 1 {
			c.CachedJSON += "}"
		} else {
			c.CachedJSON = c.CachedJSON[:len(c.CachedJSON)-1] + "}"
		}
		c.LastCached = time.Now()
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(c.CachedJSON))
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [-port] [-refresh] [-clientDropTimer] [-pk]\n", os.Args[0])
		flag.PrintDefaults()
	}
	var port string
	flag.StringVar(&port, "port", "80", "Listen Port")
	var ttl time.Duration
	flag.DurationVar(&ttl, "refresh", 1*time.Second, "Cache data JSON for T; Viewers get same result during the time; e.g., \"1s\"")
	var cdt time.Duration
	flag.DurationVar(&cdt, "clientDropTimer", 15*time.Second, "Drop client from data if no updates for T; e.g., \"15s\"")
	var pk string
	flag.StringVar(&pk, "pk", "sc_hostname", "Key for identify clients; Default is Pod name (sc_hostname)")

	flag.Parse()

	r := chi.NewRouter()
	r.Use(httprate.LimitByIP(1000, 1*time.Minute))
	r.Use(middleware.Heartbeat("/healthz"))
	r.Use(middleware.Logger)

	coll := newCollector(ttl, cdt, pk)
	r.With(middleware.AllowContentType("application/json")).Post("/", coll.Post)
	r.Get("/all", coll.GetAll)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatal(err)
	}
}
