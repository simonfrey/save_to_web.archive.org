# Proxyfy

![Proxyfy Logo](https://image.flaticon.com/icons/svg/148/148800.svg)


---

Help me to grow this project:

[![Donate Button](https://liberapay.com/assets/widgets/donate.svg)](https://liberapay.com/l1am0)

---

---

Wrapper around gimmeproxy.com - API compatible to http.Client

With proxyfy you can simply add proxied requests to your go client.

It is as simple as changing `http.Get("https://github.com")` to `proxyfy.Get("https://github.com")`. From now on all the request get forwarded trough a random proxy.

*For getting more than the 240 free request, please visit [gimmeproxy.com](https://a.paddle.com/v2/click/14088/32188?link=975) and get yourself an API key. It's just a few bucks per month*

---

# Installation

Simply execute `go get -u github.com/L1am0/proxyfy` in your shell.

# Usage
You have to setup proxyfy via an initalizer.
There are two different ones available:

## Simple

```go
proxyfy := proxyfy.NewProxyfy(apiKey,schema string)
```

Here is already a part of the config set:
```go
GimmeProxyConfig{
	ApiKey:         apiKey,
	Protocol:       scheme,
	MaxCheckPeriod: 30,
	Get:            true,
	Post:           true,
	SupportsHTTPS:  true,
	Referer:true,
	MinSpeed: 2000,
}
```

## Advanced

```go
proxyfy := NewProxyfyAdvancedConfig(gimmeConfig GimmeProxyConfig)
```

The gimmeConfig is defined via the following struct:
```go
type GimmeProxyConfig struct {
	ApiKey         string
	Get            bool
	Post           bool
	Cookies        bool
	Referer        bool
	UserAgent      bool
	SupportsHTTPS  bool
	AnonymityLevel int
	Protocol       string
	Port           string
	Country        string
	MaxCheckPeriod int
	Websites       string
	MinSpeed       float64
	NotCountry     string
	IPPort         bool
	Curl           bool
}
```

For documentation on the different values visit: [https://gimmeproxy.com/](https://a.paddle.com/v2/click/14088/32188?link=975)

# Examples
## [Basic] Use Proxyfys build in http.Client
Fire 30 GET requests and print the http response code

```go
package main

import(
	"github.com/L1am0/proxyfy"
	"fmt"
)
func main() {
	proxyfy := proxyfy.NewProxyfy("", "http")

	for i := 0; i < 30; i++ {
		resp, err := proxyfy.Get("https://t3n.de")
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Println(resp.StatusCode)
	}

}
```

## [Advanced] GetRandomProxy
Use your own setup of a http.Client with Proxyfy providing you with a random proxy url


```go
package main

import(
	"github.com/L1am0/proxyfy"
	"fmt"
	"net/http"
)
func main() {
	proxyfy := proxyfy.NewProxyfy("", "http")
	proxyURL := proxyfy.GetRandomProxy()

	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
	}
	client := http.Client{
		Transport: transport,
	}

	req, err := http.NewRequest("GET", "https://t3n.de", nil) 
	if err != nil {
		fmt.Println(err)
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(resp.StatusCode)
}
```

# Available Functions

## GetAllProxys
GetAllProxys returns a slice containing all proxies that are available

```go 
func (c *Proxyfy) GetAllProxys() []*url.URL 
```

## GetRandomProxy
GetRandomProxy returns a random *url.URL for usage with own http.Client

```go
func (c *Proxyfy) GetRandomProxy() *url.URL
```

## Do
Do executes the given *http.Request using a random proxy

```go
func (c *Proxyfy) Do(req *http.Request) (resp *http.Response, err error)
```

Similar to `http.Do()`

## Get
Get is a wrapper around Do(). Executes a GET request using a random proxy

```go
func (c *Proxyfy) Get(url string) (resp *http.Response, err error)
```
Similar to `http.Get()`

## Head

Head is a wrapper around Do(). Executes a HEAD request using a random proxy

```go
func (c *Proxyfy) Head(url string) (resp *http.Response, err error) 
```

Similar to `http.Head`

## Post

Post is a wrapper around Do(). Executes a POST request using a random proxy

```go
func (c *Proxyfy) Post(url string, contentType string, body io.Reader) (resp *http.Response, err error)
```

Similar to `http.Post`

## PostForm

PostForm is a wrapper around Post(). Executes a Post request using a random proxy and sending data as x-www-form-urlencoded

```go
func (c *Proxyfy) PostForm(url string, data url.Values) (resp *http.Response, err error) 
```

Similar to `http.PostForm`

## NewProxyfyAdvancedConfig
NewProxyfyAdvancedConfig sets up proxyfy with an advanced configuration.

```go
func NewProxyfyAdvancedConfig(gimmeConfig GimmeProxyConfig) *Proxyfy
```

Also have a look in the part Usage of this README

## NewProxyfy

NewProxyfy sets up proxyfy with a minimal amount of input data

```go
func NewProxyfy(apiKey, scheme string) *Proxyfy
```

Also have a look in the part Usage of this README

---

*For getting more than the 240 free request, please visit [gimmeproxy.com](https://a.paddle.com/v2/click/14088/32188?link=975) and get yourself an API key. It's just a few bucks per month*


**License**

MIT License

**Icons**

 Icons made by Smashicons from www.flaticon.com is licensed by CC 3.0 BY
