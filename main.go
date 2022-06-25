package main

import (
	"container/list"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"git.fchannel.org/fchannel-index/activitypub"
	"git.fchannel.org/fchannel-index/util"
)

var TorProxy = "127.0.0.1:9050"

func main() {
	var queue = util.Queue{Queue: list.New()}

	queue.Enqueue("https://fchan.xyz")

	if err := IndexInstances(queue); err != nil {
		fmt.Println(err)
	}
}

func IndexInstances(queue util.Queue) error {
	var index []string
	var alreadyChecked []string

	for queue.Len() > 0 {
		cur, err := queue.Dequeue()

		if err != nil {
			return err
		}

		re := regexp.MustCompile(`https?://[^/]*`)
		domain := re.FindString(cur)

		if indexed := CheckIfIndex(index, domain); !indexed {
			index = append(index, domain)
		}

		following, err := GetInstances(cur + "/following")

		for _, e := range following {
			if indexed := CheckIfIndex(alreadyChecked, e); !indexed {
				alreadyChecked = append(alreadyChecked, e)
				queue.Enqueue(e)
			}
		}

		followers, err := GetInstances(cur + "/followers")

		for _, e := range followers {
			if indexed := CheckIfIndex(alreadyChecked, e); !indexed {
				alreadyChecked = append(alreadyChecked, e)
				queue.Enqueue(e)
			}
		}
	}

	for _, e := range index {
		fmt.Println(e)
	}

	return nil
}

func GetInstances(route string) ([]string, error) {
	var instances []string

	req, err := http.NewRequest("GET", route, nil)

	if err != nil {
		return instances, err
	}

	resp, err := RouteProxy(req)

	if err != nil {
		return instances, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return instances, err
	}

	var respCollection activitypub.Collection

	if err := json.Unmarshal(body, &respCollection); err != nil {
		return instances, err
	}

	for _, e := range respCollection.Items {
		instances = append(instances, e.Id)
	}

	return instances, nil
}

func CheckIfIndex(index []string, value string) bool {
	for _, e := range index {
		if e == value {
			return true
		}
	}
	return false
}

func RouteProxy(req *http.Request) (*http.Response, error) {

	var proxyType = GetPathProxyType(req.URL.Host)

	if proxyType == "tor" {
		proxyUrl, err := url.Parse("socks5://" + TorProxy)

		if err != nil {
			return nil, err
		}

		proxyTransport := &http.Transport{Proxy: http.ProxyURL(proxyUrl)}
		client := &http.Client{Transport: proxyTransport, Timeout: time.Second * 15}
		return client.Do(req)
	}

	return http.DefaultClient.Do(req)
}

func GetPathProxyType(path string) string {
	if TorProxy != "" {
		re := regexp.MustCompile(`(http://|http://)?(www.)?\w+\.onion`)
		onion := re.MatchString(path)
		if onion {
			return "tor"
		}
	}

	return "clearnet"
}
