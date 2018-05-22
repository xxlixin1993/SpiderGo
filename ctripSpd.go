// 抓取携程r网站游记  特别容易被封 建议使用ip代理
package main

import (
	"sync"
	"time"
	"Spider/client"
	"os"
	"fmt"
	"github.com/olivere/elastic"
	"log"
	"github.com/PuerkitoBio/goquery"
)

var (
	// 起多少个goroutine去抓取
	cFetchGoroutineTotal = 1

	// 列表页 p最大22067
	ctripListUrlFmt = "http://you.ctrip.com/TravelSite/Home/IndexTravelListHtml?Idea=0&Type=2&Plate=0&p=%d"

	// 携程域名
	ctripDomain = "http://you.ctrip.com"

	// Ctrip协程池
	cPool map[int]*Ctrip
)

const (
	// es 索引 ci
	kCtripIndex = "ci"

	// 间隔时间 s
	kCtripIntervalSecond = 5

	// 休息时间 s
	kCtripSleepSecond = 5

	// 休息标记
	kCtripSleepFlag = "ctripSleep"
)

type Ctrip struct {
	sleep   chan int
	urlChan chan string
	done    chan int
	twg     sync.WaitGroup
}


func main() {
	start := time.Now()

	var esErr error
	client.EsClient, esErr = elastic.NewClient()
	if esErr != nil {
		log.Printf("es client err : %s", esErr)
		os.Exit(10)
	}

	doCtrip()

	secs := time.Since(start).Seconds()

	fmt.Printf("time: %f", secs)
}

func newCtrip() *Ctrip {
	return &Ctrip{
		sleep:   make(chan int),
		urlChan: make(chan string),
		done:    make(chan int),
	}
}

func doCtrip() {

}

// 抓取列表页获取详情页id
func fetchCtripList() {

}

// 抓取详情页
func (c *Ctrip) fetchCtrip(esChan *client.EsChannel) {
	defer c.twg.Done()
	for {
		select {
		case <-c.done:
			close(c.urlChan)
			return
		case url := <-c.urlChan:
			if url == kSleepFlag {
				time.Sleep(time.Second * kSleepSecond)
				continue
			}

			// 先从列表页获取详情页url

			// 在从详情页中获取title content url


			doc, err := client.ProxyRequestHtml(url)

			if err != nil {
				log.Printf("http do request err (%s)", err)
				continue
			}

			title, _ := doc.Find(".strategy-title .title-text").Html()

			s := doc.Find(".strategy-description").Each(func(i int, s *goquery.Selection) {

			})
			esContent := &client.EsContent{
				Title:   title,
				Content: s.Text(),
				Url:     url,
			}
			if title != "" {
				esChan.EsChan <- esContent
			} else {
				log.Printf("None tile %s, url %s\n", title, url)
			}
		}
	}
}
