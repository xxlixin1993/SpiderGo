package client

import (
	"github.com/labstack/gommon/log"
	"time"
)

var ipPool = []string{
	//"218.60.8.99:3129",
	//"218.60.8.98:3129",
	//暂时被猫途鹰封了
	//"118.114.77.47:8080",

	"27.154.144.211:21364",
	"183.15.122.201:40309",
	"113.124.222.238:40192",
	"115.215.50.125:29384",
}

const (
	kIpApi = "http://piping.mogumiao.com/proxy/api/get_ip_bs?appKey=9b96ddf28bdc4261a87f3213d718b533&count=10&expiryDate=0&format=1&newLine=2"

)

func DoIp(sleepSecond time.Duration){
	ticker := time.NewTicker(sleepSecond)
	for _ = range ticker.C {
		go GetIpPool()
	}
}

var IpStr string

func GetIpPool() {
	result, err := RequestIpApi(kIpApi)

	if err != nil {
		log.Fatalf("get ip error(%s)", err)
	}

	// 临时存ip:port 之后赋值个ipPool
	ipPoolTmp := make([]string, 0)

	if result.Code == "0" {
		for _, val := range result.Msg {
			//IpStr += val.Ip + ":" + val.Port + "\n"
			ipPoolTmp = append(ipPoolTmp, val.Ip + ":" + val.Port)
		}
	}

	if len(ipPoolTmp) > 0 {
		//tool.Write(IpStr)
		ipPool = ipPoolTmp
	}


}
