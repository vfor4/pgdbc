package main

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"
)

// https://go.dev/blog/context

type Results []Result

type Result struct {
	Title, URL string
}

// /search?q=golang&timeout=1s
func handleSearch(w http.ResponseWriter, req *http.Request) {
	var (
		ctx    context.Context
		cancel context.CancelFunc
	)

	timeout, err := time.ParseDuration(req.FormValue("timeout"))
	if err == nil {
		ctx, cancel = context.WithTimeout(context.Background(), timeout)
	} else {
		ctx, cancel = context.WithCancel(context.Background())
	}
	defer cancel()

	query := req.FormValue("q")
	if query == "" {
		http.Error(w, "no query", http.StatusBadRequest)
		return
	}

	cx := req.FormValue("cx")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ctx = newContext(ctx, cx)
	start := time.Now()
	results, err := search(ctx, query)
	elapsed := time.Since(start)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := resultsTemplate.Execute(w, struct {
		Results          Results
		Timeout, Elapsed time.Duration
	}{
		Results: results,
		Timeout: timeout,
		Elapsed: elapsed,
	}); err != nil {
		log.Println(err)
		return
	}
}

var resultsTemplate = template.Must(template.New("results").Parse(`
<html>
<head/>
<body>
  <ol>
  {{range .Results}}
    <li>{{.Title}} - <a href="{{.URL}}">{{.URL}}</a></li>
  {{end}}
  </ol>
  <p>{{len .Results}} results in {{.Elapsed}}; timeout {{.Timeout}}</p>
</body>
</html>
`))

func fromRequest(req *http.Request) (string, error) {
	return "", nil
}

func newContext(ctx context.Context, value string) context.Context {
	return nil
}

func search(ctx context.Context, query string) (Results, error) {
	req, err := http.NewRequest(http.MethodGet, "https://www.googleapis.com/customsearch/v1", nil)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()
	q.Set("q", query)
	apiKey, found := os.LookupEnv("GG_API_KEY")
	if !found {
		panic("missing api key")
	}
	q.Set("key", apiKey)
	cx, ok := ctx.Value("cx").(string)
	if !ok || cx == "" {
		cx, found = os.LookupEnv("GG_SEARCH_ID")
		if !found {
			panic("missing search engine id")
		}
	}
	q.Set("cx", cx)
	req.URL.RawQuery = q.Encode()

	var results Results
	err = httpDo(ctx, req, func(resp *http.Response, err error) error {
		if err 
	})
	return nil, nil
}

func httpDo(ctx context.Context, req *http.Request, f func(*http.Response, error) error) error {
	c := make(chan error, 1) // why buffered channel here?
	req = req.WithContext(ctx)
	go func() {
		c <- f(http.DefaultClient.Do(req))
	}()
	select {
	case <-ctx.Done():
		<-c
		return ctx.Err()
	case err := <-c:
		return err
	}
}
