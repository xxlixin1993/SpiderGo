package client

import (
	"sync"
	"fmt"
	"log"
	"context"
	"github.com/olivere/elastic"
)

// es client
var EsClient *elastic.Client

type EsContent struct {
	Title   string
	Url     string
	Content string
}

type EsChannel struct {
	EsChan chan *EsContent
	Done   chan int
	Esg    sync.WaitGroup
}


func NewEsChannel() *EsChannel {
	return &EsChannel{
		EsChan: make(chan *EsContent),
		Done:   make(chan int),
	}
}

// 写入es
func (esc *EsChannel) Output(index string) {
	defer esc.Esg.Done()
	for {
		select {
		case <-esc.Done:
			close(esc.EsChan)
			return
		case data := <-esc.EsChan:
			// 判断必须有title才能输出到es
			// 需要先建es index和中文分词option
			// 1. curl -XPUT http://localhost:9200/tti
			// 2. curl -XPOST http://localhost:9200/tti/fulltext/_mapping -H 'Content-Type:application/json' -d'
			//{
			// "properties": {
			//     "content": {
			//         "type": "text",
			//         "analyzer": "ik_max_word",
			//         "search_analyzer": "ik_max_word"
			//     }
			// }
			//}'
			if data.Title != "" {
				put1, err := EsClient.Index().Index(index).Type("fulltext").
					BodyJson(data).Do(context.Background())

				if err != nil {
					log.Printf("insert es err (%s)", err)
				}
				fmt.Printf("Id %s to index %s, type %s, url %s\n", put1.Id, put1.Index, put1.Type, data.Url)
			}
		}
	}
}