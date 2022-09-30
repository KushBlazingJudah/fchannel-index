package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"git.fchannel.org/fchannel-index/activitypub"
)

var (
	forceProxy  = flag.Bool("force-proxy", false, "force all connections to go through the proxy proxy")
	proxy       = flag.String("proxy", "", "proxy for requests, such as \"socks5h://127.0.0.1:9050\" for Tor")
	supportsTor = flag.Bool("tor", false, "the configured proxy points at tor; automatically turned on if proxy port is 9050")
	first       = flag.String("first", "https://usagi.reisen", "first instance to be contacted")

	hc                    = http.DefaultTransport
	hct http.RoundTripper = nil

	hcp = &sync.Pool{
		New: func() any {
			return &http.Client{Transport: hc}
		},
	}
	hctp = &sync.Pool{
		New: func() any {
			return &http.Client{Transport: hct}
		},
	}

	errNoTor = fmt.Errorf("proxy does not support tor")
)

type state struct {
	sync.WaitGroup
	sync.Mutex

	// seen is a list of domains and if the request was successful or not.
	// It also doubles as a way to check if we already visted something or not.
	seen map[string]error
}

func setupProxy() {
	u, err := url.Parse(*proxy)
	if err != nil {
		log.Fatal("failed parsing proxy url: %v", err)
	}

	if u.Port() == "9050" && !*supportsTor {
		log.Print("marking tor as supported since port is 9050")
		*supportsTor = true
	}

	hct = &http.Transport{
		Proxy:               http.ProxyURL(u),
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 15 * time.Second,
	}

	if *forceProxy {
		hc = hct
	}
}

func isTor(host string) bool {
	return strings.HasSuffix(host, ".onion")
}

func route(r string) (*http.Response, error) {
	req, err := http.NewRequest("GET", r, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "FChannel-Index-Scan")

	if isTor(req.URL.Host) {
		if *supportsTor {
			c := hctp.Get().(*http.Client)
			defer hcp.Put(c)
			return c.Do(req)
		}
		return nil, errNoTor
	}

	c := hcp.Get().(*http.Client)
	defer hcp.Put(c)
	return c.Do(req)
}

func instances(list *[]string, r string) error {
	resp, err := route(r)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("status code is not 200: %d", resp.StatusCode)
	}

	var coll activitypub.Collection
	if err := json.NewDecoder(resp.Body).Decode(&coll); err != nil {
		return err
	}

	for _, e := range coll.Items {
		*list = append(*list, e.Id)
	}

	return nil
}

func (s *state) walk(cur string) {
	defer s.Done()

	log.Printf("walking %s", cur)

	// In order to build the list, we check two things:
	// - the /followers
	// - and the /following
	//
	// If one of the requests fail, this instance is marked as dead.
	//
	// From the information that both /followers and /following provide, we
	// build a new list of instances that weren't already seen and walk those.

	var found []string

	if err := instances(&found, cur+"/following"); err != nil {
		log.Printf("fatal error on %s/following: %v", cur, err)

		s.Lock()
		s.seen[cur] = err
		s.Unlock()
		return
	} else if err = instances(&found, cur+"/followers"); err != nil {
		log.Printf("fatal error on %s/followers: %v", cur, err)

		s.Lock()
		s.seen[cur] = err
		s.Unlock()
		return
	}

	// Everything here was successful, walk the list.
	s.Lock()
	for _, e := range found {
		e = strings.TrimSpace(e)
		if len(e) == 0 {
			continue
		}
		if _, wasSeen := s.seen[e]; !wasSeen {
			s.seen[e] = nil
			s.Add(1)
			go s.walk(e)
		}
	}
	s.Unlock()
}

func main() {
	flag.Parse()

	if *proxy != "" {
		setupProxy()
	}

	log.Printf("starting crawl (force proxy: %t)", *forceProxy)

	s := state{}
	s.seen = map[string]error{}
	then := time.Now()
	s.Add(1)
	s.walk(*first)
	s.Wait()

	log.Print("done crawl in ", time.Since(then))

	alive, dead := sort(s.seen)

	htf, err := os.Create("instances.html")
	if err != nil {
		log.Fatalf("failed creating instances.html: %v", err)
	}
	defer htf.Close()

	jsf, err := os.Create("instances.json")
	if err != nil {
		log.Fatalf("failed creating instances.json: %v", err)
	}
	defer htf.Close()

	if err := createHTML(alive, dead, htf); err != nil {
		log.Fatalf("error writing html: %v", err)
	}

	if err := createJSON(alive, dead, jsf); err != nil {
		log.Fatalf("error writing json: %v", err)
	}
}

func sort(index map[string]error) ([]string, map[string]error) {
	instances := map[string]struct{}{}
	deadInstances := map[string]error{}

	for e, err := range index {
		re := regexp.MustCompile(`https?://[^/]*`)
		domain := re.FindString(e)
		if err == nil {
			instances[domain] = struct{}{}
		} else {
			deadInstances[domain] = err
		}
	}

	i := make([]string, 0, len(instances))
	for k := range instances {
		i = append(i, k)
	}

	return i, deadInstances
}

func createHTML(alive []string, dead map[string]error, out io.Writer) error {
	if _, err := io.WriteString(out, `<div style="max-width: 800px; margin: 0 auto;">
<h1 style="text-align: center;">Current known instances</h1>
<ul style="list-style-type: none;">
`); err != nil {
		return err
	}

	for _, e := range alive {
		if _, err := fmt.Fprintf(out, "<li><a href=\"%s\">%s</a></li>\n", html.EscapeString(e), html.EscapeString(e)); err != nil {
			panic(err)
		}
	}
	if _, err := io.WriteString(out, `</ul>
<h2 style="text-align: center;">Dead</h2>
<ul style="list-style-type: none;">
`); err != nil {
		return err
	}

	for e, err := range dead {
		if _, err = fmt.Fprintf(out, "<li><b>%s</b>: <code>%v</code></li>\n", html.EscapeString(e), html.EscapeString(err.Error())); err != nil {
			return err
		}
	}

	_, err := io.WriteString(out, "</ul>\n</div>\n")
	return err
}

func createJSON(alive []string, dead map[string]error, out io.Writer) error {
	o := struct {
		Alive []string          `json:"alive"`
		Dead  map[string]string `json:"dead"`
	}{alive, map[string]string{}}
	for k, v := range dead {
		o.Dead[k] = v.Error()
	}
	return json.NewEncoder(out).Encode(o)
}
