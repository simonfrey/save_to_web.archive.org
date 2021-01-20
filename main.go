package main

import (
	"flag"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/simonfrey/proxyfy"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type SafeMap struct {
	v         map[string]int
	baseUrl   string
	domainUrl string
	queue     chan string
	mux       sync.Mutex
}

var wgQuery sync.WaitGroup
var internalUrls, useProxy bool

func (c *SafeMap) Add(url string, urlType int) {
	c.mux.Lock()
	defer c.mux.Unlock()

	if internalUrls && strings.HasPrefix(url, "/") {
		url = c.domainUrl + url
	}

	if strings.Index(url, c.baseUrl) == 0 && c.v[url] == 0 {
		c.v[url] = urlType
		log.Println(len(c.v), ":", url)

		if urlType == 1 {
			wgQuery.Add(1)

			//Enque newly found url
			go func() { c.queue <- url }()
		}
	}
}

func (c *SafeMap) Get() map[string]int {
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.v
}

func main() {
	//Setup expected flags
	useProxyPtr := flag.Bool("p", false, "Proxyfy the requests")
	internalUrlsPtr := flag.Bool("i", false, "Also use interal urls e.g. /hello/world")
	sleepBetweenRequests := flag.Bool("s", true, "Sleep between add request to not be flagged internet archive")

	//Parse commandline args
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		log.Fatal("Please specify the page you want to save. Form: http[s]://[yourwebsite.com]")
	}

	//Check if the url is valid and parse it into right format
	urlStruct, err := url.Parse(args[0])
	if err != nil {
		log.Fatal(err)
	}
	cUrl := urlStruct.String()

	pU, err := url.Parse(cUrl)
	if err != nil {
		log.Fatal(err)

	}
	dUrl := pU.Scheme + "://" + pU.Host

	useProxy = *useProxyPtr
	internalUrls = *internalUrlsPtr

	log.Printf("\n Save URL: %s\n Use Proxy: %t\n Crawl internal urls: %t\n", cUrl, useProxy, internalUrls)
	gimmeConfig := proxyfy.GimmeProxyConfig{
		Protocol:      "http",
		Get:           true,
		Post:          true,
		SupportsHTTPS: true,
		MinSpeed:      500,
	}
	proxyfy := proxyfy.NewProxyfyAdvancedConfig(gimmeConfig)

	//Setup Chanel
	queue := make(chan string)

	uMap := &SafeMap{
		v:         make(map[string]int),
		baseUrl:   cUrl,
		domainUrl: dUrl,
		queue:     queue,
	}

	//Add first baseurl to queue
	uMap.Add(cUrl, 1)

	//Close queue after all urls have been processed
	go func() {
		wgQuery.Wait()
		close(queue)
	}()

	//Endless loop to range over channel
	for sUrl := range queue {
		analyzeUrl(sUrl, uMap, proxyfy)
	}

	log.Printf("Found %d subelements on %s", len(uMap.Get()), cUrl)

	//Internet archive only allows single connection. So we have to do the request slowly
	for sUrl, urlType := range uMap.Get() {
		if urlType == 2 || urlType == 1 {
			addUrl(sUrl, proxyfy, *sleepBetweenRequests)

			if *sleepBetweenRequests {
				sleepTime := time.Duration(rand.Intn(10) + 5)
				fmt.Printf("Sleep for %d seconds\n", sleepTime)
				time.Sleep(sleepTime)
			}
		}
	}

	log.Println("Done")
}

func analyzeUrl(sUrl string, uMap *SafeMap, proxyfy *proxyfy.Proxyfy) {

	var err error
	var res *http.Response

	if useProxy {
		res, err = proxyfy.Get(sUrl)
	} else {
		res, err = http.Get(sUrl)
	}
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Printf("status code error: %d %s\n", res.StatusCode, res.Status)
		return
	}
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Println(err)
		return
	}

	// use CSS selector found with the browser inspector
	// for each, use index and item
	doc.Find("body a").Each(func(index int, item *goquery.Selection) {
		linkTag := item
		link, _ := linkTag.Attr("href")
		uMap.Add(strings.Split(link, "#")[0], 1)
	})

	doc.Find("body img").Each(func(index int, item *goquery.Selection) {
		linkTag := item
		link, _ := linkTag.Attr("src")
		uMap.Add(strings.Split(link, "#")[0], 2)
	})

	wgQuery.Done()
}

func addUrl(sUrl string, proxyfy *proxyfy.Proxyfy, sleepBetweenRequests bool) {
	baseUrl := "https://web.archive.org/save/"
	for i := 0; i < 50; i++ {
		log.Println("[", i, "] Try for ", sUrl)
		var err error
		var res *http.Response

		if useProxy {
			res, err = proxyfy.Get(baseUrl + sUrl)
		} else {
			res, err = http.Get(baseUrl + sUrl)
		}

		if err != nil {
			log.Println(err)
			continue
		}

		if res.StatusCode == 200 {
			break
		}

		if sleepBetweenRequests {
			sleepTime := time.Duration(rand.Intn(5) + 5)
			fmt.Printf("Sleep for %d seconds\n", sleepTime)
			time.Sleep(sleepTime)
		}
	}
	log.Println("Added: ", sUrl)

}
