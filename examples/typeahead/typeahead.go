// A simple spelling corrector implemented as a HTTP server.
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/aaw/levtrie"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var usage = `
typeahead implements a simple spelling corrector served over HTTP.

Example: /search?q=helo returns spelling corrections for "helo".

Accepted query params are;
 q: The string query. Default is the empty string.
 n: The max number of results. Default is 10.
 p: The length of the prefix of the query string to ignore for edit distance.
    Default is 1/5 the length of the query string.
 d: The edit distance to search within. Default is 1/3 the length of the
    non-ignored suffix of the query.
 e: If non-zero and fewer than the desired number of results are found with the
    specified criteria, the results will be augmented with strings that have a
    prefix that matches the query criteria. Default: 1

Parameters:
`

var dictFile = flag.String("dictionary", "/usr/share/dict/words",
	"A file containing correctly spelled words, one per line.")

var port = flag.Int("port", 3000, "The port the server will listen on.")

var logger *log.Logger

// newSearchHandler loads the dictionary file at filename into a Trie and
// returns the Trie wrapped in a searchHandler. The dictionary file should
// contain a list of words, one per line.
func newSearchHandler(filename string) searchHandler {
	t := levtrie.New()
	logger.Printf("Loading %v, this may take a few seconds...\n", filename)
	start := time.Now()
	file, err := os.Open(filename)
	if err != nil {
		panic(fmt.Sprintf("%v: %v", filename, err))
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	count := 0
	for scanner.Scan() {
		word := strings.ToLower(scanner.Text())
		t.Set(word, "")
		count += 1
	}
	elapsed := time.Since(start)
	logger.Printf("Loaded %v words from %v in time %v.\n",
		count, filename, elapsed)
	return searchHandler{t: t}
}

type searchHandler struct {
	t *levtrie.Trie
}

// uniq returns up to n strings in the input slice, omitting duplicates.
func uniq(xs []string, n int) []string {
	seen := make(map[string]bool)
	j := 0
	for i, x := range xs {
		if !seen[x] {
			seen[x] = true
			xs[j] = xs[i]
			j++
			if j >= n {
				return xs[:j]
			}
		}
	}
	return xs[:j]
}

// config specifies parameters for a Trie search
type config struct {
	query          string
	limit          int
	dist           int8
	ignorePrefix   int
	expandSuffixes bool
}

// parseQuery parses query params into a config for searching a Trie. See usage
// message defined at the top of this file for a list of accepted query params.
func parseQuery(params map[string][]string) *config {
	cfg := &config{}
	if qp, ok := params["q"]; ok && len(qp) > 0 {
		cfg.query = qp[0]
	}
	cfg.limit = 10
	if qp, ok := params["n"]; ok && len(qp) > 0 {
		if i, err := strconv.Atoi(qp[0]); err == nil {
			cfg.limit = i
		}
	}
	pset := false
	if qp, ok := params["p"]; ok && len(qp) > 0 {
		if i, err := strconv.Atoi(qp[0]); err == nil {
			cfg.ignorePrefix = i
			pset = true
		}
	}
	if !pset {
		cfg.ignorePrefix = len(cfg.query) / 5
	}
	cfg.dist = 1
	dset := false
	if qp, ok := params["d"]; ok && len(qp) > 0 {
		if i, err := strconv.ParseInt(qp[0], 10, 8); err == nil {
			cfg.dist = int8(i)
			dset = true
		}
	}
	if !dset {
		raw_dist := (len(cfg.query) - cfg.ignorePrefix) / 3
		if raw_dist > 255 {
			raw_dist = 255
		}
		cfg.dist = int8(raw_dist)
	}
	cfg.expandSuffixes = true
	if qp, ok := params["e"]; ok && len(qp) > 0 {
		if i, err := strconv.Atoi(qp[0]); err == nil && i != 0 {
			cfg.expandSuffixes = false
		}
	}
	return cfg
}

func (s searchHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cfg := parseQuery(r.URL.Query())
	results := []string{}
	if cfg.query != "" {
		start := time.Now()
		kvResults := s.t.SuggestAfterExactPrefix(
			cfg.query, cfg.ignorePrefix, cfg.dist, cfg.limit)
		if cfg.expandSuffixes && len(kvResults) < cfg.limit {
			res := s.t.SuggestSuffixesAfterExactPrefix(
				cfg.query, cfg.ignorePrefix, cfg.dist, cfg.limit)
			kvResults = append(kvResults, res...)
		}
		elapsed := time.Since(start)
		for _, kv := range kvResults {
			results = append(results, kv.Key)
		}
		results = uniq(results, cfg.limit)
		logger.Printf("Query %+v returned %v results in time %v\n",
			cfg, len(results), elapsed)
	}
	j, _ := json.Marshal(results)
	fmt.Fprintf(w, string(j))
}

var indexText = `
<html>
  <head>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/jquery/1.11.2/jquery.min.js"
            integrity="sha256-1OxYPHYEAB+HIz0f4AdsvZCfFaX4xrTD9d2BtGLXnTI="
            crossorigin="anonymous"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/easy-autocomplete/1.3.5/jquery.easy-autocomplete.min.js"
            integrity="sha256-aS5HnZXPFUnMTBhNEiZ+fKMsekyUqwm30faj/Qh/gIA="
            crossorigin="anonymous"></script>
    <link rel="stylesheet"
          href="https://cdnjs.cloudflare.com/ajax/libs/easy-autocomplete/1.3.5/easy-autocomplete.min.css"
          integrity="sha256-fARYVJfhP7LIqNnfUtpnbujW34NsfC4OJbtc37rK2rs="
          crossorigin="anonymous" />
    <link rel="stylesheet"
          href="https://cdnjs.cloudflare.com/ajax/libs/easy-autocomplete/1.3.5/easy-autocomplete.themes.min.css"
          integrity="sha256-kK9BInVvQN0PQuuyW9VX2I2/K4jfEtWFf/dnyi2C0tQ="
          crossorigin="anonymous" />
  </head>
  <body>
    <form>
      <div id="remote">
        <input id="remote-suggest" />
      </div>
    </form>
    <script type="text/javascript">
      var options = {
        url: function(query) { return "../search?q=" + query; }
      };
      $("#remote-suggest").easyAutocomplete(options);
    </script>
  </body>
</html>
`

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, usage)
		flag.PrintDefaults()
	}
	flag.Parse()
	logger = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, indexText)
	})
	http.Handle("/search", newSearchHandler(*dictFile))
	logger.Printf("Serving on http://0.0.0.0:%d\n", *port)
	http.ListenAndServe(fmt.Sprintf(":%d", *port), nil)
}
