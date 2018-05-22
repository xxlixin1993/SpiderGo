// 抓取tripadvisor网站游记
package main

import (
	"fmt"
	"strconv"
	"time"
	"log"
	"sync"
	"github.com/PuerkitoBio/goquery"
	"github.com/olivere/elastic"
	"os"
	"Spider/client"
)

var (
	// 要抓取的游记最大id
	tripadvisorTotalId = 10000

	// 起多少个goroutine去抓取
	fetchGoroutineTotal = 3

	// 要抓取的url ex: https://www.tripadvisor.cn/TourismBlog-t6598
	tripadvisorDetail = "https://www.tripadvisor.cn/TourismBlog-t"

	// es client
	esClient *elastic.Client

	// Tripadvisor协程池
	tPool map[int]*Tripadvisor
)

const (
	// es 索引
	kTripadvisorTitleIndex = "tti"

	// 间隔时间 s
	kIntervalSecond = 58

	// 休息时间 s
	kSleepSecond = 60

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
	esClient, esErr = elastic.NewClient()
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
		sleep:    make(chan int),
		urlChan: make(chan string),
		done:    make(chan int),
	}
}


// 开始获取页面信息
func doTripadvisor() {
	tPool = make(map[int]*Tripadvisor)

	esChan := newEsChannel()

	esChan.esg.Add(1)
	go esChan.output(kTripadvisorTitleIndex)

	for gnum := 0; gnum < fetchGoroutineTotal; gnum ++ {
		tPool[gnum] = newTripadvisor()

		tPool[gnum].twg.Add(1)
		go tPool[gnum].fetchTripadvisor(esChan)
	}

	go timerJob()

	for i := 1; i <= tripadvisorTotalId; i++ {
		tPool[i%fetchGoroutineTotal].urlChan <- tripadvisorDetail + strconv.Itoa(i)
	}

	for key := range tPool {
		close(tPool[key].done)
		tPool[key].twg.Wait()
	}

	close(esChan.done)
	esChan.esg.Wait()
}

// 间隔一段时间在执行
func timerJob(){
	t := time.NewTimer(time.Second * kIntervalSecond)

	for _ = range t.C {
		for _, val := range tPool {
			val.urlChan <- kSleepFlag
		}
	}
}



// 抓取
func (t *Tripadvisor) fetchTripadvisor(esChan *EsChannel) {
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

			doc, err := client.ProxyRequest(url)

			if err != nil {
				log.Printf("http do request err (%s)", err)
				continue
			}

			title := doc.Find(".title-text").Text()

			s := doc.Find(".strategy-description").Each(func(i int, s *goquery.Selection) {

			})
			esContent := &EsContent{
				Title:   title,
				Content: s.Text(),
				Url:     url,
			}
			esChan.esChan <- esContent
		}
	}
}
