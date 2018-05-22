package client

import (
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"log"
	"Spider/tool"
	"net/url"
	"time"
)

// 代理请求
func ProxyRequest(uri string) (*goquery.Document, error) {
	var client *http.Client
	var ipProxy string

	ipPoolLen := len(ipPool)
	randomIp := tool.GenerateRangeNum(ipPoolLen)

	if randomIp != ipPoolLen {
		ipProxy = ipPool[randomIp]
		//ipProxy = "//222.76.187.13:8118"
		proxy := func(_ *http.Request) (*url.URL, error) {
			// 设置代理ip
			return url.Parse(ipProxy)
		}

		transport := &http.Transport{Proxy: proxy}
		client = &http.Client{Transport: transport, Timeout: time.Second * 10}
	} else {
		// 不使用代理
		client = &http.Client{}
	}

	resp, err := client.Get(uri)

	if err != nil {
		log.Printf("request error(%s)", ipProxy)
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("ip maybe don not have permission")
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Printf("doc error(%s)", ipProxy)
		return nil, err
	}

	defer resp.Body.Close()
	return doc, err
}
