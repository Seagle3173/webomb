package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"time"
)

//发送请求以及处理response
func responseHandle(client *http.Client, r *http.Request, s string) {
	r.URL, _ = url.Parse(s)
	time.Sleep(time.Millisecond * time.Duration(d))
	resp, err := client.Do(r)
	if err != nil {
		fmt.Printf("Error response: %v\n", err)
		waitGroup.Done()
		return
	}
	if resp.StatusCode < 300 {
		fmt.Printf("%s %v\n", resp.Request.URL, resp.StatusCode)
	}
	resp.Body.Close()
	waitGroup.Done()
}

func randomRespHandle(client *http.Client, r *http.Request, s string) {
	r.URL, _ = url.Parse(s)
	r.Header.Set("User-Agent", agentList[rand.Intn(8)])
	time.Sleep(time.Millisecond * time.Duration(d))
	resp, err := client.Do(r)
	if err != nil {
		fmt.Printf("Error response: %v\n", err)
		waitGroup.Done()
		return
	}
	if resp.StatusCode < 300 {
		fmt.Printf("%s %v\n", resp.Request.URL, resp.StatusCode)
	}
	resp.Body.Close()
	waitGroup.Done()
}
