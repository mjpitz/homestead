package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/urfave/cli/v2"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/mjpitz/homestead/internal/index"
	_ "github.com/mjpitz/homestead/internal/index"
	"github.com/mjpitz/myago/flagset"
	"github.com/mjpitz/myago/zaputil"
)

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
	loader func(ctx context.Context, name string) (index.Index, error)
}

// Probe is called to ensure that datasource is working properly. We should conditionally return this based on the
// presence of the requested dataset.
func (s *Server) Probe(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dataset, err := s.loader(r.Context(), vars["dataset"])
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer dataset.Close()

	http.ServeContent(w, r, "", time.Now(), bytes.NewReader(nil))
}

func (s *Server) Search(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dataset, err := s.loader(r.Context(), vars["dataset"])
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer dataset.Close()

	q := &Search{}
	err = json.NewDecoder(r.Body).Decode(q)
	if err != nil && err != io.EOF {
		log.Println(err.Error())
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	fields := dataset.Schema()

	filtered := make([]string, 0)
	for _, field := range fields {
		if len(q.Target) == 0 || strings.Contains(string(field.Text), q.Target) {
			filtered = append(filtered, string(field.Text))
		}
	}

	err = json.NewEncoder(w).Encode(filtered)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
	}
}

func (s *Server) Query(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dataset, err := s.loader(r.Context(), vars["dataset"])
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer dataset.Close()

	q := &Query{}
	err = json.NewDecoder(r.Body).Decode(q)
	if err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	query := []index.Query{
		{"timestamp", ">=", q.Range.From.UnixMilli()},
		{"timestamp", "<", q.Range.To.UnixMilli()},
	}

	for _, filter := range q.AdhocFilters {
		query = append(query, index.Query{
			Field:    filter.Key,
			Operator: filter.Operator,
			Value:    filter.Value,
		})
	}

	docs := dataset.Query(query...)
	hits := dataset.Get(map[string]interface{}{}, docs...)

	results := make([]interface{}, 0, len(q.Targets))
	for _, target := range q.Targets {
		switch target.Type {
		case "timeserie":
			ts := &TimeSeries{
				Target: target.Target,
			}

			for _, hit := range hits {
				doc := *(hit.(*map[string]interface{}))

				val, ok := doc[target.Target].(float64)
				if !ok {
					continue
				}

				ts.Datapoints = append(ts.Datapoints, []float32{
					float32(val), float32(doc["timestamp"].(int64)),
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
	dataset, err := s.loader(r.Context(), vars["dataset"])
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer dataset.Close()

	fields := dataset.Schema()

	err = json.NewEncoder(w).Encode(fields)
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
				loader: func(ctx context.Context, name string) (index.Index, error) {
					path := filepath.Join(cfg.DatasetBaseDir, name, "latest")

					return index.Open(path, true)
				},
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

			zaputil.Extract(ctx.Context).Info("listening on :8080")
			return http.ListenAndServe(":8080", handler)
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
