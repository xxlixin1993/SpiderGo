package client

import (
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"log"
)

// 代理请求
func ProxyRequest(uri string) (*goquery.Document, error) {
	client := &http.Client{}

	resp, err := client.Get(uri)

	if err != nil {
		log.Printf("request error(%s)", err)
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)

	defer resp.Body.Close()
	return doc, err
}
