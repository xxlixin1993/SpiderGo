// 抓取携程网站游记  特别容易被封 建议使用ip代理
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
	"os/signal"
	"syscall"
	"strings"
)

var (
	// 起多少个goroutine去抓取
	cFetchGoroutineTotal = 5

	// 列表页 p最大22067
	ctripListPId = 22067

	// 当前列表页id
	ctripListNowPId = 1

	// 列表页 p最大22067
	ctripListUrlFmt = "http://you.ctrip.com/TravelSite/Home/IndexTravelListHtml?Idea=0&Type=2&Plate=0&p=%d"

	// Ctrip协程池
	cPool map[int]*Ctrip
)

const (
	// es 索引 ci
	kCtripIndex = "ci"

	// 间隔时间 s
	kCtripIntervalSecond = 5

	// 休息时间 s
	kCtripSleepSecond = 1

	// 休息标记
	kCtripSleepFlag = "ctripSleep"

	// 携程域名
	kCtripDomain = "http://you.ctrip.com"
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
	waitSignal()

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
	cPool = make(map[int]*Ctrip)

	esChan := client.NewEsChannel()

	esChan.Esg.Add(1)
	go esChan.Output(kCtripIndex)

	for gnum := 0; gnum < cFetchGoroutineTotal; gnum++ {
		cPool[gnum] = newCtrip()

		cPool[gnum].twg.Add(1)
		go cPool[gnum].fetchCtrip(esChan)
	}

	go ctripTimerJob()

	for i := ctripListNowPId; i <= ctripListPId; i++ {
		listUri := fmt.Sprintf(ctripListUrlFmt, i)
		cPool[i%cFetchGoroutineTotal].urlChan <- listUri
	}

	for key := range cPool {
		close(cPool[key].done)
		cPool[key].twg.Wait()
	}

	close(esChan.Done)
	esChan.Esg.Wait()
}

// 间隔一段时间在执行
func ctripTimerJob() {
	t := time.NewTimer(time.Second * kCtripIntervalSecond)

	for _ = range t.C {
		for _, val := range cPool {
			val.urlChan <- kCtripSleepFlag
		}
	}
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
			if url == kCtripSleepFlag {
				time.Sleep(time.Second * kCtripSleepSecond)
				continue
			}

			// 先从列表页获取详情页url
			fmt.Println(url)
			detailUrlSlice, err := fetchList(url)
			if err != nil {
				log.Printf("http do list request err (%s)", err)
				continue
			}

			if len(detailUrlSlice) > 0 {
				// 在从详情页中获取title content url
				for _, detailUri := range detailUrlSlice {
					doc, err := client.ProxyRequestHtml(detailUri)
					if err != nil {
						log.Printf("http do detail request err (%s)", err)
						continue
					}

					var title string

					// title有两种 第一种
					titleFirst, err := doc.Find(".ctd_head_left h2").Html()
					if err != nil {
						log.Printf("find detail title1 err (%s)", err)
						continue
					}
					title = strings.TrimSpace(titleFirst)

					if title == "" {
						// title第二种
						titleSecond, err := doc.Find(".title1").Html()
						if err != nil {
							log.Printf("find detail title2 err (%s)", err)
							continue
						}
						title = strings.TrimSpace(titleSecond)
					}

					content := doc.Find(".ctd_content p").Text()

					esContent := &client.EsContent{
						Title:   title,
						Content: content,
						Url:     detailUri,
					}

					if esContent.Title != "" {
						esChan.EsChan <- esContent
					} else {
						log.Printf("None title %s, url %s\n", esContent.Title, esContent.Url)
					}
				}
			} else {
				log.Printf("Do not fetched list url %s\n", url)
			}
		}
	}
}

// 返回该列表页详情页的uri
func fetchList(uri string) ([]string, error) {
	doc, err := client.ProxyRequestHtml(uri)
	if err != nil {
		return nil, err
	}

	findRes := make([]string, 0)
	doc.Find(".cpt").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if exists {
			findRes = append(findRes, kCtripDomain+href)
		}
	})

	return findRes, nil
}

// Wait signal
func waitSignal() {
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan)

	sig := <-sigChan

	fmt.Printf("signal: %d", sig)

	switch sig {
	case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
		fmt.Println("out")
	case syscall.SIGUSR1:
		fmt.Println("catch the signal SIGUSR1")
	default:
		fmt.Println("signal do not know")
	}
}
