package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"time"
	"sync"
	"strings"
	"flag"

	"git.fchannel.org/fchannel-index/activitypub"
)

var TorProxy func (*http.Request) (*url.URL, error) // Signature of http.ProxyURL
var ForceTor bool = true

type proxy uint8
const (
	tor proxy = iota
	clear
)

type state struct {
	sync.WaitGroup
	sync.Mutex
	seen map[string]struct{}
}

func main() {
	flag.BoolVar(&ForceTor, "forcetor", false, "force all connections to go through Tor")
	proxy := flag.String("tor", "127.0.0.1:9050", "tor proxy address")
	flag.Parse()

	if *proxy != "" {
		TorProxy = http.ProxyURL(&url.URL{Scheme: "socks5", Host: *proxy})
	} else {
		log.Printf("turning force tor off since proxy is empty")
		ForceTor = false
	}

	log.Printf("starting crawl (force tor: %t)", ForceTor)

	s := state{}
	s.seen = map[string]struct{}{}
	then := time.Now()
	s.Add(1)
	s.Walk("https://fchan.xyz", 0)
	s.Wait()

	log.Print("done crawl in ", time.Since(then))

	if err := CreateHTMLIndex(s.seen); err != nil {
		panic(err)
	}
}

func (s *state) Walk(cur string, depth int) {
	defer s.Done()
	s.Lock()
	s.seen[cur] = struct{}{}
	s.Unlock()

	log.Printf("walking %s (depth: %d)", cur, depth)
	check, err := GetInstances(cur + "/following")
	if err != nil {
		log.Printf("fatal error on %s: %s", cur, err)
		return
	}

	followers, err := GetInstances(cur + "/followers")
	if err != nil {
		log.Printf("non-fatal error on %s: %s", cur, err)
	}

	check = append(check, followers...)

	s.Lock()
	for _, e := range check {
		e = strings.TrimPrefix(e, " ")
		if _, wasSeen := s.seen[e]; !wasSeen {
			s.seen[e] = struct{}{}
			s.Add(1)
			go s.Walk(e, depth+1)
		}
	}
	s.Unlock()
}

func GetInstances(route string) ([]string, error) {
	req, err := http.NewRequest("GET", route, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "FChannel-Index-Scan")

	resp, err := RouteProxy(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var respCollection activitypub.Collection

	if err := json.Unmarshal(body, &respCollection); err != nil {
		return nil, err
	}

	var instances []string
	for _, e := range respCollection.Items {
		instances = append(instances, e.Id)
	}

	return instances, nil
}

func RouteProxy(req *http.Request) (*http.Response, error) {
	if ForceTor || GetPathProxyType(req.URL.Host) == tor {
		if TorProxy == nil {
			return nil, fmt.Errorf("no tor proxy configured")
		}

		log.Printf("tor request: %s", req.URL)
		proxyTransport := &http.Transport{Proxy: TorProxy}
		client := &http.Client{Transport: proxyTransport, Timeout: time.Second * 15}
		return client.Do(req)
	}

	log.Printf("request: %s", req.URL)
	return http.DefaultClient.Do(req)
}

func GetPathProxyType(path string) proxy {
	if TorProxy != nil {
		re := regexp.MustCompile(`(http://|http://)?(www.)?\w+\.onion`)
		onion := re.MatchString(path)
		if onion {
			return tor
		}
	}

	return clear
}

func CreateHTMLIndex(index map[string]struct{}) error {
	file, err := os.Create("instance-index.html")
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err = file.WriteString(`<div style="max-width: 800px; margin: 0 auto;">
<h1 style="text-align: center;">Current known instances</h1>
<ul style="list-style-type: none;">
`); err != nil {
		return err
	}

	instances := map[string]struct{}{}

	for e := range index {
		re := regexp.MustCompile(`https?://[^/]*`)
		domain := re.FindString(e)
		instances[domain] = struct{}{}
	}

	for e := range instances {
		if _, err = file.WriteString(fmt.Sprintf("<li><a href=\"%s\">%s</a></li>\n", e, e)); err != nil {
			panic(err)
		}
	}

	_, err = file.WriteString("</ul>\n</div>\n")
	return err
}
