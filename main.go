package main

import (
	"container/list"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"time"

	"git.fchannel.org/fchannel-index/activitypub"
	"git.fchannel.org/fchannel-index/util"
)

var TorProxy = "127.0.0.1:9050"

func main() {
	var queue = util.Queue{Queue: list.New()}

	queue.Enqueue("https://fchan.xyz")

	var err error
	var index []string

	if index, err = IndexInstances(queue); err != nil {
		fmt.Println(err)
		return
	}

	CreateHTMLIndex(index)
}

func IndexInstances(queue util.Queue) ([]string, error) {
	var index []string
	var alreadyChecked []string

	for queue.Len() > 0 {
		cur, err := queue.Dequeue()

		if err != nil {
			return index, err
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

		re := regexp.MustCompile(`https?://[^/]*`)
		domain := re.FindString(cur)

		if indexed := CheckIfIndex(index, domain); (len(followers) > 0 || len(following) > 0) && !indexed {
			index = append(index, domain)
		}
	}

	return index, nil
}

func GetInstances(route string) ([]string, error) {
	var instances []string

	req, err := http.NewRequest("GET", route, nil)

	if err != nil {
		return instances, err
	}

	req.Header.Set("User-Agent", "FChannel-Index-Scan")

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

func CreateHTMLIndex(index []string) error {
	file, err := os.Create("instance-index.html")

	if err != nil {
		return err
	}

	var text string

	text += fmt.Sprintln("<div style=\"max-width: 800px; margin: 0 auto;\">")
	text += fmt.Sprintln("<h1 style=\"text-align: center;\"> Current known instances</h1>")
	text += fmt.Sprintln("<ul style=\"list-style-type: none;\">")

	for _, e := range index {
		text += fmt.Sprintf("<li><a href=\"%s\">%s</a></li>\n", e, e)
	}

	text += fmt.Sprintln("</ul>")
	text += fmt.Sprintln("</div>")

	_, err = file.WriteString(text)

	return err
}
