package client

import (
	"Spider/tool"
	"encoding/json"
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
)

type ReqClient struct {
	Client  *http.Client
	IpProxy string
}

func getClient() *ReqClient {
	var client *http.Client
	var ipProxy string

	ipPoolLen := len(ipPool)
	randomIp := tool.GenerateRangeNum(ipPoolLen)

	if randomIp != ipPoolLen {
		ipProxy = "//" + ipPool[randomIp]

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

	reqClient := &ReqClient{
		Client:  client,
		IpProxy: ipProxy,
	}

	return reqClient
}

// html
func ProxyRequestHtml(uri string) (*goquery.Document, error) {
	reqClient := getClient()
	resp, err := reqClient.Client.Get(uri)

	if err != nil {
		log.Printf("request error(%s)", reqClient.IpProxy)
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("ip maybe don not have permission, ip(%s)", reqClient.IpProxy)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Printf("doc error(%s)", reqClient.IpProxy)
		return nil, err
	}

	defer resp.Body.Close()
	return doc, err
}

type BdJson struct {
	Errno int        `json:"errno"`
	Msg   string     `json:"msg"`
	Data  BdDataJson `json:"data"`
}

type BdDataJson struct {
	NotesList  []BdInfo `json:"notes_list"`
	NotesCount int      `json:"notes_count"`
	Abtest     int      `json:"abtest"`
}

type BdInfo struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Nid     string `json:"nid"`
}

// json
func ProxyRequestJson(uri string) (*BdJson, error) {
	reqClient := getClient()
	resp, err := reqClient.Client.Get(uri)

	if err != nil {
		log.Printf("request error(%s)", reqClient.IpProxy)
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("ip maybe don not have permission")
	}

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Printf("response error(%s)", reqClient.IpProxy)
		return nil, err
	}

	res := &BdJson{}
	JsonErr := json.Unmarshal(body, &res)
	return res, JsonErr
}

type IpApiJson struct {
	Code string `json:"code"`
	Msg []Msg `json:"msg"`
}

type Msg struct {
	Port string `json:"port"`
	Ip	string `json:"ip"`
}

func RequestIpApi(url string) (*IpApiJson, error){
	client := &http.Client{}
	resp, err := client.Get(url)

	if err != nil {
		log.Printf("api request error(%s)", err)
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("ip maybe don not have permission")
	}

	body, err := ioutil.ReadAll(resp.Body)



	if err != nil {
		log.Printf("api response error(%s)", err)
		return nil, err
	}

	res := &IpApiJson{}
	JsonErr := json.Unmarshal(body, &res)
	return res, JsonErr
}
