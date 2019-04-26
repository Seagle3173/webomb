package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	h      bool   //帮助
	u      string //进行爆破的url
	t      int    //线程数
	m      string //指定的字典文件
	d      int    //每次请求之间的延迟
	agent  string //代理的header文件，使用json
	proxy  string //使用的代理的ip和端口
	random bool   //是否使用随机的agent
)

var (
	//默认使用userAgent配置发送请求
	userAgent = map[string]string{
		"User-Agent": "Mozilla/5.0 (Windows NT 6.1; WOW64; rv:34.0) Gecko/20100101 Firefox/34.0",
		"Connection": "keep-alive",
		"Accept":     "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3",
	}
	//agentList为user-agent的随机列表
	agentList = []string{
		//Opera浏览器
		"Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/39.0.2171.95 Safari/537.36 OPR/26.0.1656.60",
		//Firefox浏览器
		"Mozilla/5.0 (Windows NT 6.1; WOW64; rv:34.0) Gecko/20100101 Firefox/34.0",
		//Safari浏览器
		"Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/534.57.2 (KHTML, like Gecko) Version/5.1.7 Safari/534.57.2",
		//Chrome浏览器
		"Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/39.0.2171.71 Safari/537.36",
		//360浏览器
		"Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/30.0.1599.101 Safari/537.36",
		//淘宝浏览器
		"Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/536.11 (KHTML, like Gecko) Chrome/20.0.1132.11 TaoBrowser/2.0 Safari/536.11",
		//猎豹浏览器
		"Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.1 (KHTML, like Gecko) Chrome/21.0.1180.71 Safari/537.1 LBBROWSER",
		//QQ浏览器
		"Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 6.1; WOW64; Trident/5.0; SLCC2; .NET CLR 2.0.50727; .NET CLR 3.5.30729; .NET CLR 3.0.30729; Media Center PC 6.0; .NET4.0C; .NET4.0E; QQBrowser/7.0.3698.400)",
	}

	waitGroup sync.WaitGroup

	requests = []*http.Request{}
)

func init() {
	flag.BoolVar(&h, "h", false, "the help")
	flag.StringVar(&u, "u", "", "set the url or dns (the url must be the directory)")
	flag.IntVar(&t, "t", 3, "set the number of threads")
	flag.StringVar(&m, "m", "", "set the wordlist you use")
	flag.IntVar(&d, "d", 0, "set the delay of two requests")
	flag.StringVar(&agent, "agent", "", "set the agent-header you use (default firefox)")
	flag.StringVar(&proxy, "proxy", "", "set the proxy you use (the famat is like this: https://127.0.0.1:4444) (do not support socks)")
	flag.BoolVar(&random, "random", false, "use the random agent or not")
	flag.Usage = usage
}
func main() {
	//start := time.Now().Unix()
	flag.Parse()
	if h || len(os.Args) == 1 || u==""{
		flag.Usage()
		return
	}
	if !(strings.HasPrefix(u, "https://") || strings.HasPrefix(u, "http://")) {
		switch resp0, _ := http.Get("http://" + u + "/"); resp0.StatusCode == http.StatusOK {
		case true:
			u = "http://" + u + "/"
		case false:
			switch resp1, _ := http.Get("https://" + u + "/"); resp1.StatusCode == http.StatusOK {
			case true:
				u = "https://" + u + "/"
			case false:
				log.Println("Can not connect the host")
				return
			}
		}
	}

	//设置proxy http、https或者socks5
	if proxy != "" {
		switch strings.HasPrefix(proxy, "https") {
		case true:
			os.Setenv("HTTPS_PROXY", proxy)
		case false:
			switch strings.HasPrefix(proxy, "http") {
			case true:
				os.Setenv("HTTP_PROXY", proxy)
			case false:
				log.Printf("Do not support the proxy of socks")
			}
		}
	}

	//建立客户端请求header
	transport := http.Transport{
		DisableKeepAlives: true,
	}
	client := &http.Client{
		Transport: &transport,
	}

	//agent参数的优先级高于
	if agent != "" {

		//清除header内容
		for key := range userAgent {
			delete(userAgent, key)
		}

		//读取json文件内容
		file, err := ioutil.ReadFile(agent)
		if err != nil {
			fmt.Printf("Agent File Open Failed: %v\n", err)
			return
		}

		//进行反序列化，并储存到userAgent中
		err = json.Unmarshal(file, &userAgent)
		if err != nil {
			fmt.Printf("httphandle: Json File Decode Failed: %v\n", err)
			return
		}
	}

	//创建线程数的通信信道
	for i := 0; i < t; i++ {
		req, err := http.NewRequest("GET", u, nil)
		if err != nil {
			fmt.Printf("Error Cearte Request: %v", err)
		}
		requests = append(requests, req)
		for key, value := range userAgent {
			req.Header.Add(key, value)
		}
	}

	//获取字典文件
	wordlist, err := os.Open(m)
	if err != nil {
		fmt.Printf("Error Open Wordlist File: %v", err)
		return
	}
	defer wordlist.Close()
	buf := bufio.NewReader(wordlist)

	//循环读取字典内容并发送请求
	switch random && agent == "" {
	case true:
	loop0:
		for {
			for _, req := range requests {
				menu, _, c := buf.ReadLine()
				if c == io.EOF {
					break loop0
				}
				waitGroup.Add(1)
				go randomRespHandle(client, req, u+string(menu))
			}
			waitGroup.Wait()
		}
		waitGroup.Wait()
	case false:
	loop1:
		for {
			for _, req := range requests {
				menu, _, c := buf.ReadLine()
				if c == io.EOF {
					break loop1
				}
				waitGroup.Add(1)
				go responseHandle(client, req, u+string(menu))
			}
			waitGroup.Wait()
		}
		waitGroup.Wait()
	}
	//end := time.Now().Unix()
	//fmt.Println(end - start)
}

func usage() {
	fmt.Fprintf(os.Stderr, `webomb version: 0.2
Brute Force the website direcories! a proper wordlist is very important!
Usage: webomb [-random] [-t number] [-m wordlist] [-d delay] [-u url]

Options:
`)
	flag.PrintDefaults()
}
