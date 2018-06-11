![Save to web.archive.org logo](https://github.com/simonfrey/save_to_web.archive.org/raw/master/logo.png "Save to web.archive.org logo")

---

# Description
Scrapes the given website for internal links and saves the found ones into [web.archive.org](https://web.archive.org/)

# Installation
**I asume you have go already installed. ([Go installation manual](https://golang.org/doc/install))**

## Dependencies
Download the dependecies via `go get`

Execute the both following commands:
`go get -u github.com/L1am0/proxyfy`

`go get -u github.com/PuerkitoBio/goquery`

## Download tool
Just clone this git repo

`git clone https://github.com/simonfrey/save_to_web.archive.org.git`

# Exectution

Navigate into the directory of the git repo.

Execute with `go run main.go http[s]://[yourwebsite.com]`. With `http[s]://[yourwebsite.com]` being the website you want to scrape and save.
