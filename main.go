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

	"git.fchannel.org/fchannel-index/activitypub"
)

// Set to nil if you don't want to use the Tor proxy
var TorProxy = http.ProxyURL(&url.URL{Scheme: "socks5", Host: "127.0.0.1:9050"})

// Set to true if you always want to use Tor
var ForceTor bool = true

type proxy uint8
const (
	tor proxy = iota
	clear
)

func main() {
	index := Walk("https://fchan.xyz", map[string]struct{}{}, 0)

	if err := CreateHTMLIndex(index); err != nil {
		panic(err)
	}
}

func Walk(cur string, seen map[string]struct{}, depth int) []string {
	log.Printf("walking %s (depth: %d)", cur, depth)
	index := []string{cur}
	check, err := GetInstances(cur + "/following")
	if err != nil {
		log.Printf("fatal error on %s: %s", cur, err)
		return nil
	}

	followers, err := GetInstances(cur + "/followers")
	if err != nil {
		log.Printf("non-fatal error on %s: %s", cur, err)
	}

	check = append(check, followers...)

	for _, e := range check {
		if _, wasSeen := seen[e]; !wasSeen {
			seen[cur] = struct{}{}
			index = append(index, Walk(e, seen, depth+1)...)
		}
	}

	return index
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

func CreateHTMLIndex(index []string) error {
	file, err := os.Create("instance-index.html")
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err = file.WriteString(`<div style="max-width: 800px; margin: 0 auto;">
<h1 style="text-align: center;"> Current known instances</h1>
<ul style="list-style-type: none;">
`); err != nil {
		return err
	}

	for _, e := range index {
		if _, err = file.WriteString(fmt.Sprintf("<li><a href=\"%s\">%s</a></li>\n", e, e)); err != nil {
			panic(err)
		}
	}

	_, err = file.WriteString("</ul>\n</div>\n")
	return err
}
