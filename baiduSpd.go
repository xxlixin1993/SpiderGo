// 抓取百度旅游游记
package main

import (
	"Spider/client"
	"fmt"
	"github.com/olivere/elastic"
	"log"
	"os"
	"sync"
	"time"
)

var (
	// 百度游记 pn表示每页开始 rn表示pageSize
	// ex: https://lvyou.baidu.com/search/ajax/searchnotes?format=ajax&type=0&pn=0&rn=20
	baiduUrlFmt = "https://lvyou.baidu.com/search/ajax/searchnotes?format=ajax&type=0&pn=%d&rn=%d"

	// rn 最大20
	rn = 20

	// 从0开始
	pn = 0

	// 起多少个goroutine去抓取
	bdFetchGoroutineTotal = 1

	// baidu协程池
	bdPool map[int]*Baidu
)

const (
	// es 索引
	kBaiduIndex = "bdi"

	// 间隔时间 s
	kBdIntervalSecond = 5

	// 休息时间 s
	kBdSleepSecond = 5

	// 休息标记
	kBdSleepFlag = "bdsleep"
)

type Baidu struct {
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

	doBaidu()
	secs := time.Since(start).Seconds()

	fmt.Printf("time: %f", secs)
}

func newBaidu() *Baidu {
	return &Baidu{
		sleep:   make(chan int),
		urlChan: make(chan string),
		done:    make(chan int),
	}
}

// 开始获取页面信息
func doBaidu() {
	bdPool = make(map[int]*Baidu)

	esChan := client.NewEsChannel()

	esChan.Esg.Add(1)
	go esChan.Output(kBaiduIndex)

	for gnum := 0; gnum < bdFetchGoroutineTotal; gnum++ {
		bdPool[gnum] = newBaidu()

		bdPool[gnum].twg.Add(1)
		go bdPool[gnum].fetchBaidu(esChan)
	}

	go bdTimerJob()

	// 百度就1000 * 20
	for i := pn; i <= 1000; i++ {
		baiduUrl := fmt.Sprintf(baiduUrlFmt, pn*rn, rn)
		bdPool[pn%bdFetchGoroutineTotal].urlChan <- baiduUrl
	}

	for key := range bdPool {
		close(bdPool[key].done)
		bdPool[key].twg.Wait()
	}

	close(esChan.Done)
	esChan.Esg.Wait()
}

// 间隔一段时间在执行
func bdTimerJob() {
	t := time.NewTimer(time.Second * kBdIntervalSecond)

	for _ = range t.C {
		for _, val := range bdPool {
			val.urlChan <- kBdSleepFlag
		}
	}
}

// 抓取
func (bd *Baidu) fetchBaidu(esChan *client.EsChannel) {
	defer bd.twg.Done()
	for {
		select {
		case <-bd.done:
			close(bd.urlChan)
			return
		case url := <-bd.urlChan:
			if url == kBdSleepFlag {
				time.Sleep(time.Second * kBdSleepSecond)
				continue
			}

			res, err := client.ProxyRequestJson(url)

			if err != nil {
				log.Printf("http do request err (%s)", err)
				continue
			}

			if len(res.Data.NotesList) > 0 {
				for _, val := range res.Data.NotesList {
					esContent := &client.EsContent{
						Title:   val.Title,
						Content: val.Content,
						Url:     url,
					}
					if val.Title != "" {
						esChan.EsChan <- esContent
					} else {
						log.Printf("None tile %s, url %s\n", val.Title, url)
					}
				}
			}
		}
	}
}
