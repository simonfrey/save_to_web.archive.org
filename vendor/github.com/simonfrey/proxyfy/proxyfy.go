//Package proxyfy provides an api compatible http.Client for making requests
//All request are routed trough a random proxy provided by gimmeproxy.com
//For getting more requests visit https://a.paddle.com/v2/click/14088/32188?link=975
package proxyfy

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/google/go-querystring/query"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
	"errors"
)

//***Proxyfy Structs

type gimmeProxyResponse struct {
	SupportsHTTPS  bool            `json:"supportsHttps"`
	Protocol       string          `json:"protocol"`
	IP             string          `json:"ip"`
	Port           string          `json:"port"`
	Get            bool            `json:"get"`
	Post           bool            `json:"post"`
	Cookies        bool            `json:"cookies"`
	Referer        bool            `json:"referer"`
	UserAgent      bool            `json:"user-agent"`
	AnonymityLevel int             `json:"anonymityLevel"`
	Websites       map[string]bool `json:"websites"`
	TsChecked      int             `json:"tsChecked"`
	Curl           string          `json:"curl"`
	IPPort         string          `json:"ipPort"`
	Type           string          `json:"type"`
	Speed          float64         `json:"speed"`
	Country        string          `json:"country"`
}

type GimmeProxyConfig struct {
	ApiKey         string  `url:"api_key,omitempty"`
	Get            bool    `url:"get,omitempty"`
	Post           bool    `url:"post,omitempty"`
	Cookies        bool    `url:"cookies,omitempty"`
	Referer        bool    `url:"referer,omitempty"`
	UserAgent      bool    `url:"user-agent,omitempty"`
	SupportsHTTPS  bool    `url:"supportsHttps,omitempty"`
	AnonymityLevel int     `url:"anonymityLevel,omitempty"`
	Protocol       string  `url:"protocol,omitempty"`
	Port           string  `url:"port,omitempty"`
	Country        string  `url:"country,omitempty"`
	MaxCheckPeriod int     `url:"maxCheckPeriod,omitempty"`
	Websites       string  `url:"websites,omitempty"`
	MinSpeed       float64 `url:"minSpeed,omitempty"`
	NotCountry     string  `url:"notCountry,omitempty"`
	IPPort         bool    `url:"ipPort,omitempty"`
	Curl           bool    `url:"curl,omitempty"`
}

type proxyStorage struct {
	Proxies []*url.URL
	mux     sync.Mutex
}

type Proxyfy struct {
	pStorage proxyStorage
}

//***Internal Methods for handling the proxy storage

func (c *proxyStorage) addProxy(url *url.URL) {
	if url == nil{
		return
	}
	if url.Scheme != "" && url.Host != "" {
		c.mux.Lock()
		defer c.mux.Unlock()
	
		c.Proxies = append(c.Proxies, url)
		proxyJson, _ := json.Marshal(c)
		ioutil.WriteFile("proxyfySAVE", proxyJson, 0644)
	}

}

func (c *proxyStorage) getRandomProxy() *url.URL {
	c.mux.Lock()
	defer c.mux.Unlock()
	if len(c.Proxies) > 0 {
		return c.Proxies[rand.Intn(len(c.Proxies))]

	} else {
		return nil
	}
}

func (c *proxyStorage) getAllProxys() []*url.URL {
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.Proxies
}

func (c *proxyStorage) removeProxy(toRemove *url.URL){
	c.mux.Lock()
	defer c.mux.Unlock()
	tmpProxies := make([]*url.URL,0)
	for k := range c.Proxies{
		if c.Proxies[k] != toRemove{
			tmpProxies = append(tmpProxies,c.Proxies[k])
		}
	}
	c.Proxies = tmpProxies
		proxyJson, _ := json.Marshal(c)
		ioutil.WriteFile("proxyfySAVE", proxyJson, 0644)
}

func (c *Proxyfy) _do(req *http.Request) (resp *http.Response, err error) {
tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}

	proxyURL := c.pStorage.getRandomProxy()

	for i := 0; i < 10 && proxyURL == nil; i++{
		time.Sleep(500 * time.Millisecond)
		proxyURL = c.pStorage.getRandomProxy()

		fmt.Println(proxyURL)
	}

	if proxyURL == nil{
		return nil, errors.New("Could not get any proxy. Maybe you hit the daily limit of 240 requests to gimmeproxy api. For getting more get yourself an API Key: https://a.paddle.com/v2/click/14088/32188?link=975")
	}

	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
		Proxy:           http.ProxyURL(proxyURL),
	}

	timeout := time.Duration(10 * time.Second)

	client := http.Client{
		Transport: transport,
		Timeout:   timeout,
	}

	resp, err = client.Do(req)

	if err != nil{
		if strings.Contains(err.Error(),"proxyconnect tcp: dial tcp") || strings.Contains(err.Error(),"Client.Timeout exceeded while awaiting headers") || strings.Contains(err.Error(),"read: connection reset by peer"){
			go c.pStorage.removeProxy(proxyURL)
			return c._do(req)
		}
	}
	return
}

func (c *Proxyfy) loadNewProxys(gimmeConfig GimmeProxyConfig) {
	qs, _ := query.Values(gimmeConfig)

	for {
		requestCountError := 5

		for i := 0; i < 240 || gimmeConfig.ApiKey != ""; i++ {

			tlsConfig := &tls.Config{
				InsecureSkipVerify: true,
			}

			transport := &http.Transport{
				TLSClientConfig: tlsConfig,
			}

			timeout := time.Duration(15 * time.Second)
			client := http.Client{Transport: transport, Timeout: timeout}

			//Get Proxy
			rResp, err := client.Get("https://gimmeproxy.com/api/getProxy?" + qs.Encode())
			if err != nil {
				continue
			}

			if rResp.StatusCode == 429 {
				requestCountError--
				if requestCountError <= 0 {
					break
				}
				continue
			}

			var pResponse gimmeProxyResponse
			defer rResp.Body.Close()
			body, err := ioutil.ReadAll(rResp.Body)

			err = json.Unmarshal(body, &pResponse)
			if err != nil {
				continue
			}

			aUrl := pResponse.Protocol + "://" + pResponse.IP + ":" + pResponse.Port

			proxyUrl, _ := url.Parse(aUrl)

			c.pStorage.addProxy(proxyUrl)
		}

		time.Sleep(1 * time.Hour)
	}
}


//***Public Methods for the use of proxyfy

//GetAllProxys returns a slice containing all proxies that are in use
func (c *Proxyfy) GetAllProxys() []*url.URL {
	return c.pStorage.getAllProxys()
}

//GetRandomProxy returns a random *url.URL for usage with own http.Client
func (c *Proxyfy) GetRandomProxy() *url.URL {
	return c.pStorage.getRandomProxy()
}


//Do executes the given *http.Request using a random proxy
func (c *Proxyfy) Do(req *http.Request) (resp *http.Response, err error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}

	proxyURL := c.pStorage.getRandomProxy()

	for i := 0; i < 10 && proxyURL == nil; i++{
		time.Sleep(500 * time.Millisecond)
		proxyURL = c.pStorage.getRandomProxy()

		fmt.Println(proxyURL)
	}

	if proxyURL == nil{
		return nil, errors.New("Could not get any proxy. Maybe you hit the daily limit of 240 requests to gimmeproxy api. For getting more get yourself an API Key: https://a.paddle.com/v2/click/14088/32188?link=975")
	}

	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
		Proxy:           http.ProxyURL(proxyURL),
	}

	timeout := time.Duration(10 * time.Second)

	client := http.Client{
		Transport: transport,
		Timeout:   timeout,
	}

	resp, err = client.Do(req)

	return

}

//Get is a wrapper around Do(). Executes a GET request using a random proxy
func (c *Proxyfy) Get(url string) (resp *http.Response, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return c._do(req)
}

//Head is a wrapper around Do(). Executes a HEAD request using a random proxy
func (c *Proxyfy) Head(url string) (resp *http.Response, err error) {
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return nil, err
	}
	return c._do(req)
}

//Post is a wrapper around Do(). Executes a POST request using a random proxy
func (c *Proxyfy) Post(url string, contentType string, body io.Reader) (resp *http.Response, err error) {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)

	return c._do(req)
}

//PostForm is a wrapper around Post(). Executes a Post request using a random proxy and sending data as x-www-form-urlencoded
func (c *Proxyfy) PostForm(url string, data url.Values) (resp *http.Response, err error) {
	return c.Post(url, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
}

//NewProxyfyAdvancedConfig sets up proxyfy with an advanced configuration.
//GimmeProxyConfig has following form (for documentation on the different values visit: https://gimmeproxy.com/#api)
//type GimmeProxyConfig struct {
//	ApiKey         string
//	Get            bool
//	Post           bool
//	Cookies        bool
//	Referer        bool
//	UserAgent      bool
//	SupportsHTTPS  bool
//	AnonymityLevel int
//	Protocol       string
//	Port           string
//	Country        string
//	MaxCheckPeriod int
//	Websites       string
//	MinSpeed       float64
//	NotCountry     string
//	IPPort         bool
//	Curl           bool
//}
func NewProxyfyAdvancedConfig(gimmeConfig GimmeProxyConfig) *Proxyfy {

	//Init Random with time as seed
	rand.Seed(time.Now().Unix())

	//Return Struct
	pf := Proxyfy{
		pStorage: proxyStorage{
			Proxies: make([]*url.URL, 0),
		},
	}

	//Load saved proxys if the file exists and contains valid json
	if _, err := os.Stat("proxyfySAVE"); err == nil {
		//TMP proxyStorage
		tmpStorage := proxyStorage{
			Proxies: make([]*url.URL, 0),
		}

		//Only overwrite if file was parsed correctly
		jsonBlob, err := ioutil.ReadFile("proxyfySAVE")
		if err == nil {

			err := json.Unmarshal(jsonBlob, &tmpStorage)
			if err == nil {
				if gimmeConfig.Protocol != "" {
					tmpSlice := make([]*url.URL, 0)

					//Filter loaded proxys for scheme
					for i := 0; i < len(tmpStorage.Proxies); i++ {
						if tmpStorage.Proxies[i].Scheme == gimmeConfig.Protocol {
							tmpSlice = append(tmpSlice, tmpStorage.Proxies[i])
						}
					}
					tmpStorage.Proxies = tmpSlice
				}
				pf.pStorage = tmpStorage
			}
		}
	}

	//Load new proxys in an own thread
	go pf.loadNewProxys(gimmeConfig)

	return &pf
}

//NewProxyfy sets up proxyfy with a minimal amount of input data
//It aready sets sane (in my eyes) defaults:
//GimmeProxyConfig{
//	ApiKey:         apiKey,
//	Protocol:       scheme,
//	MaxCheckPeriod: 30,
//	Get:            true,
//	Post:           true,
//	SupportsHTTPS:  true,
//	Referer:true,
//	MinSpeed: 2000,
//}
func NewProxyfy(apiKey, scheme string) *Proxyfy {
	//Check for sane scheme
	if scheme != "http" && scheme != "socks5" && scheme != "socks4" {
		fmt.Println("[Proxyfy] Given scheme not valid. A valid scheme would be (http|socks5|socks4). Falling back to http")
		scheme = "http"
	}

	//Setup Config Struct
	gimmeConfig := GimmeProxyConfig{
		ApiKey:         apiKey,
		Protocol:       scheme,
		MaxCheckPeriod: 240,
		Get:            true,
		Post:           true,
		SupportsHTTPS:  true,
	}

	return NewProxyfyAdvancedConfig(gimmeConfig)
}
