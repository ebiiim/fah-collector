package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
)

const paramIdent = "paramIdent"

type collector struct {
	M          sync.Map
	CachedJSON string
	LastCached time.Time
	CacheTTL   time.Duration
}

func newCollector(cacheTTL time.Duration) *collector {
	c := &collector{
		CacheTTL: cacheTTL,
	}
	return c
}

func (c *collector) Post(w http.ResponseWriter, r *http.Request) {
	p, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ident := chi.URLParam(r, paramIdent)
	c.M.Store(ident, p)
	w.WriteHeader(http.StatusOK)
}

func (c *collector) GetAll(w http.ResponseWriter, r *http.Request) {
	if c.LastCached.Add(c.CacheTTL).Before(time.Now()) {
		log.Println("INFO: creating JSON")
		c.CachedJSON = "{"
		c.M.Range(func(key, value interface{}) bool {
			c.CachedJSON += fmt.Sprintf("\"%s\": %s,", key, value)
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
	var port string
	flag.StringVar(&port, "port", "80", "Listen Port")
	var ttl time.Duration
	flag.DurationVar(&ttl, "refresh", 5*time.Second, "Refresh Interval")
	flag.Parse()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Heartbeat("/healthz"))
	r.Use(httprate.LimitByIP(100, 1*time.Minute))

	coll := newCollector(ttl)
	r.With(middleware.AllowContentType("application/json")).Post(fmt.Sprintf("/{%s}", paramIdent), coll.Post)
	r.Get("/all", coll.GetAll)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatal(err)
	}
}
