package main

import (
	"flag"
	"github.com/L1am0/proxyfy"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/url"
	"strings"
	"sync"
)

type SafeMap struct {
	v       map[string]int
	baseUrl string
	queue   chan string
	mux     sync.Mutex
}

var wgQuery sync.WaitGroup

func (c *SafeMap) Add(url string, urlType int) {
	c.mux.Lock()
	defer c.mux.Unlock()

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

	log.Println("Save URL: ", cUrl)

	gimmeConfig := proxyfy.GimmeProxyConfig{
		Protocol:       "http",
		Get:            true,
		Post:           true,
		SupportsHTTPS:  true,
		MinSpeed:       500,
	}
	proxyfy := proxyfy.NewProxyfyAdvancedConfig(gimmeConfig)

	//Setup Chanel
	queue := make(chan string)

	uMap := SafeMap{
		v:       make(map[string]int),
		baseUrl: cUrl,
		queue:   queue,
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

	log.Printf("Found %d subelements on %s",len(uMap.Get()),cUrl)

	//Internet archive only allows single connection. So we have to do the request slowly
	for sUrl, urlType := range uMap.Get() {
		if urlType == 2 || urlType == 1 {
			addUrl(sUrl,proxyfy)
		}
	}


	log.Println("Done")
}

func analyzeUrl(sUrl string, uMap SafeMap, proxyfy *proxyfy.Proxyfy) {

	res, err := proxyfy.Get(sUrl)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Println("status code error: %d %s", res.StatusCode, res.Status)
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

func addUrl(sUrl string, proxyfy *proxyfy.Proxyfy) {
	baseUrl := "https://web.archive.org/save/"
	for i := 0; i < 50; i++ {
		log.Println("[", i, "] Try for ", sUrl)
		resp, err := proxyfy.Get(baseUrl + sUrl)
		if err != nil {
			log.Println(err)
			continue
		}

		if resp.StatusCode == 200 {
			break
		}
	}
	log.Println("Added: ", sUrl)


}
