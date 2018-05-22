// 抓取tripadvisor网站游记
package main

import (
	"Spider/client"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/olivere/elastic"
	"log"
	"os"
	"strconv"
	"sync"
	"time"
)

var (
	// 要抓取的游记最大id
	tripadvisorTotalId = 10000

	// 起多少个goroutine去抓取
	fetchGoroutineTotal = 3

	// 要抓取的url ex: https://www.tripadvisor.cn/TourismBlog-t6598
	tripadvisorDetail = "https://www.tripadvisor.cn/TourismBlog-t"

	// Tripadvisor协程池
	tPool map[int]*Tripadvisor

	// 当前id 初始为0 当前跑到8562被封了 所以从这开始继续
	tripadvisorId = 8562
)

const (
	// es 索引 tti
	kTripadvisorTitleIndex = "tti"

	// 间隔时间 s
	kIntervalSecond = 5

	// 休息时间 s
	kSleepSecond = 31

	// 休息标记
	kSleepFlag = "sleep"
)

type Tripadvisor struct {
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

	doTripadvisor()

	secs := time.Since(start).Seconds()

	fmt.Printf("time: %f", secs)
}

func newTripadvisor() *Tripadvisor {
	return &Tripadvisor{
		sleep:   make(chan int),
		urlChan: make(chan string),
		done:    make(chan int),
	}
}

// 开始获取页面信息
func doTripadvisor() {
	tPool = make(map[int]*Tripadvisor)

	esChan := client.NewEsChannel()

	esChan.Esg.Add(1)
	go esChan.Output(kTripadvisorTitleIndex)

	for gnum := 0; gnum < fetchGoroutineTotal; gnum++ {
		tPool[gnum] = newTripadvisor()

		tPool[gnum].twg.Add(1)
		go tPool[gnum].fetchTripadvisor(esChan)
	}

	go timerJob()

	for i := tripadvisorId; i <= tripadvisorTotalId; i++ {
		tPool[i%fetchGoroutineTotal].urlChan <- tripadvisorDetail + strconv.Itoa(i)
	}

	for key := range tPool {
		close(tPool[key].done)
		tPool[key].twg.Wait()
	}

	close(esChan.Done)
	esChan.Esg.Wait()
}

// 间隔一段时间在执行
func timerJob() {
	t := time.NewTimer(time.Second * kIntervalSecond)

	for _ = range t.C {
		for _, val := range tPool {
			val.urlChan <- kSleepFlag
		}
	}
}

// 抓取
func (t *Tripadvisor) fetchTripadvisor(esChan *client.EsChannel) {
	defer t.twg.Done()
	for {
		select {
		case <-t.done:
			close(t.urlChan)
			return
		case url := <-t.urlChan:
			if url == kSleepFlag {
				time.Sleep(time.Second * kSleepSecond)
				continue
			}

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
