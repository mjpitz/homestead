package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/blevesearch/bleve/v2"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/urfave/cli/v2"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/mjpitz/myago/flagset"
	"github.com/mjpitz/myago/zaputil"
)

func newCache(loader func(ctx context.Context, name string) (bleve.Index, error)) *Cache {
	cache := make(chan map[string]bleve.Index, 1)
	cache <- make(map[string]bleve.Index)

	return &Cache{
		loader: loader,
		cache:  cache,
	}
}

type Cache struct {
	loader func(ctx context.Context, name string) (bleve.Index, error)
	cache  chan map[string]bleve.Index
}

func (c *Cache) Open(ctx context.Context, name string) (bleve.Index, error) {
	index := <-c.cache
	defer func() {
		c.cache <- index
	}()

	existing, ok := index[name]
	if ok && existing != nil {
		return existing, nil
	}

	var err error
	index[name], err = c.loader(ctx, name)
	return index[name], err
}

type Search struct {
	Target string `json:"target"`
}

type Query struct {
	PanelID int `json:"panelId"`
	Range   struct {
		From time.Time `json:"from"`
		To   time.Time `json:"to"`
		Raw  struct {
			From string `json:"from"`
			To   string `json:"to"`
		} `json:"raw"`
	} `json:"range"`
	Interval       string `json:"interval"`
	IntervalMillis int    `json:"intervalMs"`
	MaxDataPoints  int    `json:"maxDataPoints"`
	Targets        []struct {
		Target string `json:"target"`
		RefID  string `json:"refId"`
		Type   string `json:"type"`
	} `json:"targets"`
	AdhocFilters []struct {
		Key      string `json:"key"`
		Operator string `json:"operator"`
		Value    string `json:"value"`
	} `json:"adhocFilters"`
}

type TimeSeries struct {
	Target     string      `json:"target"`
	Datapoints [][]float32 `json:"datapoints"`
}

type Table struct {
	Columns []Column        `json:"columns"`
	Rows    [][]interface{} `json:"rows"`
	Type    string          `json:"type"`
}

type Column struct {
	Text string `json:"text"`
	Type string `json:"type"`
}

type TagKey struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// Server provides a generic implementation of a JSON DataSource in Grafana. Data is backed by bleve search indexes. A
// description of the spec can be found here: https://grafana.com/grafana/plugins/simpod-json-datasource/
type Server struct {
	cache *Cache
}

// Probe is called to ensure that datasource is working properly. We should conditionally return this based on the
// presence of the requested dataset.
func (s *Server) Probe(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	_, err := s.cache.Open(r.Context(), vars["dataset"])
	if err != nil {
		http.NotFound(w, r)
	} else {
		http.ServeContent(w, r, "", time.Now(), bytes.NewReader(nil))
	}
}

func (s *Server) Search(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dataset, err := s.cache.Open(r.Context(), vars["dataset"])
	if err != nil {
		http.NotFound(w, r)
		return
	}

	q := &Search{}
	err = json.NewDecoder(r.Body).Decode(q)
	if err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	fields, err := dataset.Fields()
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	filtered := fields
	if len(q.Target) > 0 {
		filtered = make([]string, 0, len(fields))
		for _, field := range fields {
			if strings.Contains(field, q.Target) {
				filtered = append(filtered, field)
			}
		}
	}

	err = json.NewEncoder(w).Encode(filtered)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
	}
}

func (s *Server) Query(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dataset, err := s.cache.Open(r.Context(), vars["dataset"])
	if err != nil {
		http.NotFound(w, r)
		return
	}

	q := &Query{}
	err = json.NewDecoder(r.Body).Decode(q)
	if err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	start := q.Range.From.Format(time.RFC3339)
	end := q.Range.To.Format(time.RFC3339)

	query := fmt.Sprintf(`timestamp:>="%s" timestamp:<"%s"`, start, end)

	search := bleve.NewSearchRequestOptions(bleve.NewQueryStringQuery(query), 1<<16, 0, false)
	search.SortBy([]string{"timestamp"})
	search.Fields = append(search.Fields, "timestamp")

	for _, target := range q.Targets {
		switch target.Type {
		case "timeserie":
			search.Fields = append(search.Fields, target.Target)
		}
	}

	res, err := dataset.SearchInContext(r.Context(), search)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	results := make([]interface{}, 0, len(q.Targets))
	for _, target := range q.Targets {
		switch target.Type {
		case "timeserie":
			ts := &TimeSeries{
				Target: target.Target,
			}

			for _, hit := range res.Hits {
				timestamp, _ := time.Parse(time.RFC3339, hit.Fields["timestamp"].(string))
				val, ok := hit.Fields[target.Target].(float64)
				if !ok {
					continue
				}

				ts.Datapoints = append(ts.Datapoints, []float32{
					float32(val), float32(timestamp.UnixMilli()),
				})
			}

			results = append(results, ts)
		}
	}

	err = json.NewEncoder(w).Encode(results)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
	}
}

func (s *Server) TagKeys(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dataset, err := s.cache.Open(r.Context(), vars["dataset"])
	if err != nil {
		http.NotFound(w, r)
		return
	}

	fields, err := dataset.Fields()
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	query := bleve.NewMatchAllQuery()
	searchReq := bleve.NewSearchRequestOptions(query, 1, 0, false)
	searchReq.SortBy([]string{"-timestamp"})
	searchReq.Fields = []string{"*"}

	res, err := dataset.SearchInContext(r.Context(), searchReq)
	hit := res.Hits[0]

	keys := make([]TagKey, len(fields))
	for i, field := range fields {
		kind := "string"

		switch v := hit.Fields[field].(type) {
		case string:
			_, err = time.Parse(time.RFC3339, v)
			if err == nil {
				kind = "datetime"
			}
		case float64:
			kind = "number"
		}

		keys[i] = TagKey{
			Type: kind,
			Text: field,
		}
	}

	err = json.NewEncoder(w).Encode(keys)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
	}
}

func (s *Server) TagValues(w http.ResponseWriter, r *http.Request) {
	http.NotFound(w, r)
}

type Config struct {
	DatasetBaseDir string         `json:"dataset_base_dir" usage:"where to look for datasets"`
	Log            zaputil.Config `json:"log"`
}

func main() {
	cfg := &Config{}

	app := &cli.App{
		Name:      "emquery",
		Usage:     "Translate Grafana requests to bleve indexes.",
		UsageText: "emquery [options]",
		Flags:     flagset.Extract(cfg),
		Before: func(ctx *cli.Context) error {
			ctx.Context = zaputil.Setup(ctx.Context, cfg.Log)
			return nil
		},
		Action: func(ctx *cli.Context) error {
			idx := &Server{
				cache: newCache(func(ctx context.Context, name string) (bleve.Index, error) {
					return bleve.Open(filepath.Join(cfg.DatasetBaseDir, name, "latest"))
				}),
			}

			routes := mux.NewRouter()

			idxRoutes := routes.PathPrefix("/{dataset}").Subrouter()

			idxRoutes.HandleFunc("/", idx.Probe).Methods(http.MethodGet)
			idxRoutes.HandleFunc("/search", idx.Search).Methods(http.MethodPost)
			idxRoutes.HandleFunc("/query", idx.Query).Methods(http.MethodPost)
			idxRoutes.HandleFunc("/tag-keys", idx.TagKeys).Methods(http.MethodPost)
			idxRoutes.HandleFunc("/tag-values", idx.TagValues).Methods(http.MethodPost)

			handler := cors.Default().Handler(routes)
			handler = h2c.NewHandler(handler, &http2.Server{})

			zaputil.Extract(ctx.Context).Info("listening on :6060")
			return http.ListenAndServe(":6060", handler)
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
