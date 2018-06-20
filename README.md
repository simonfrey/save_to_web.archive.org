![Save to web.archive.org logo](https://github.com/simonfrey/save_to_web.archive.org/raw/master/logo.png "Save to web.archive.org logo")

---

Help me to grow this project:

[![Donate Button](https://liberapay.com/assets/widgets/donate.svg)](https://liberapay.com/l1am0)

---

# Description
Scrapes the given website for internal links and saves the found ones into [web.archive.org](https://web.archive.org/)

# Installation
**I assume you have already installed go. ([Go installation manual](https://golang.org/doc/install))**

## Dependencies

Download the dependecies via `go get`

Execute the following two commands:

```
go get -u github.com/simonfrey/proxyfy
```

```
go get -u github.com/PuerkitoBio/goquery
```

## Download tool

Just clone the git repo

```
git clone https://github.com/simonfrey/save_to_web.archive.org.git
```

# Execution

Navigate into the directory of the git repo.

Execute with: 

*Please Replace `http[s]://[yourwebsite.com]` with the url of the website you want to scrape and save.*
```
go run main.go http[s]://[yourwebsite.com]
```

******Additional commandline arguments:**

`-p` for proxyfing the requests

`-i` for also crawling internal urls (e.g. /test/foo)

So if you want to use the tool with also crawling interal links and use a proxy for that it would be the following command

```
go run main.go -p -i http[s]://[yourwebsite.com] 
```
