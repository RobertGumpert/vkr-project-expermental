package github_request

import (
	"fmt"
	"log"
	"sync"
	"testing"
)

const token = "aef3219befb0b5a71ebfcf5876dd8c8d9eeb0077"

func TestFlowNewClient(t *testing.T) {
	_, err := NewGithubClient(token, 5)
	if err != nil {
		t.Fatal(err)
	}
	log.Println("OK.")
}

func waitOneRequestsGroup(waitSignal chan bool, resultChannel chan *Response, gwg *sync.WaitGroup) {
	defer gwg.Done()
	wg := new(sync.WaitGroup)
	select {
	case <-waitSignal:
		wg.Add(1)
		go func(resultChannel chan *Response, wg *sync.WaitGroup) {
			defer wg.Done()
			response := <-resultChannel
			fmt.Println(response.Response.StatusCode)
			if response.Err != nil {
				log.Println(response.Err)
			}
			return
		}(resultChannel, wg)
	case response := <-resultChannel:
		if response == nil {
			log.Println("NIL")
		}
		if response.Err == nil && response.Response != nil{
			log.Println(response.Response.StatusCode)
		}
	}
	wg.Wait()
	return
}

func waitGroupRequestsGroup(responsesChannel, deferResponsesChannel chan map[string]*Response, gwg *sync.WaitGroup) {
	defer gwg.Done()
	wg := new(sync.WaitGroup)
	select {
	case mp := <-deferResponsesChannel:
		wg.Add(1)
		fmt.Println("Wait defer channel...")
		fmt.Println(mp)
		go func(deferResponsesChannel chan map[string]*Response, wg *sync.WaitGroup) {
			defer wg.Done()
			response := <-deferResponsesChannel
			fmt.Println("Defer channel...")
			fmt.Println(response)
			return
		}(deferResponsesChannel, wg)
	case mp := <-responsesChannel:
		fmt.Println("Get all!")
		fmt.Println(mp)
	}
	wg.Wait()
	return
}

func TestFlowTaskManagement(t *testing.T) {
	c, err := NewGithubClient("", 5)
	if err != nil {
		t.Fatal(err)
	}
	wg := new(sync.WaitGroup)
	for i := 0; i < 3; i++ {
		var (
			waitSignal    = make(chan bool)
			resultChannel = make(chan *Response)
		)
		task, err := c.AddOneRequest()
		if err != nil {
			log.Println("Skip...")
			continue
		}
		event, nowait := task(Request{
			URL: "https://api.github.com/repos/facebook/react",
		}, CORE, waitSignal, resultChannel)
		if nowait {
			wg.Add(1)
			go func(waitSignal chan bool, resultChannel chan *Response, wg *sync.WaitGroup) {
				defer wg.Done()
				wg.Add(1)
				go waitOneRequestsGroup(waitSignal, resultChannel, wg)
				event()
				return
			}(waitSignal, resultChannel, wg)
		} else {
			wg.Add(1)
			go func(waitSignal chan bool, resultChannel chan *Response, wg *sync.WaitGroup) {
				defer wg.Done()
				wg.Add(1)
				go waitOneRequestsGroup(waitSignal, resultChannel, wg)
				return
			}(waitSignal, resultChannel, wg)
		}
		log.Println("Next...")
	}
	for i := 0; i < 5; i++ {
		var (
			responsesChannel, deferResponsesChannel = make(chan map[string]*Response), make(chan map[string]*Response)
		)
		task, err := c.AddGroupRequests()
		if err != nil {
			log.Println("Skip...")
			continue
		}
		event, nowait := task([]Request{
			{
				URL: "https://api.github.com/repos/facebook/react",
			},
			{
				URL: "https://api.github.com/repos/gin-gonic/gin",
			},
			{
				URL: "https://api.github.com/repos/vuejs/vue",
			},
			{
				URL: "https://api.github.com/repos/angular/angular",
			},
			{
				URL: "https://api.github.com/repos/pallets/flask",
			},
			{
				URL: "https://api.github.com/repos/square/okhttp",
			},
			{
				URL: "https://api.github.com/repos/microsoft/terminal",
			},
			{
				URL: "https://api.github.com/repos/vercel/hyper",
			},
			{
				URL: "https://api.github.com/repos/alacritty/alacritty",
			},
		}, CORE, responsesChannel, deferResponsesChannel)
		if nowait {
			wg.Add(1)
			go func(responsesChannel, deferResponsesChannel chan map[string]*Response, wg *sync.WaitGroup) {
				defer wg.Done()
				wg.Add(1)
				go waitGroupRequestsGroup(responsesChannel, deferResponsesChannel, wg)
				event()
				return
			}(responsesChannel, deferResponsesChannel, wg)
		} else {
			wg.Add(1)
			go func(responsesChannel, deferResponsesChannel chan map[string]*Response, wg *sync.WaitGroup) {
				defer wg.Done()
				wg.Add(1)
				go waitGroupRequestsGroup(responsesChannel, deferResponsesChannel, wg)
				return
			}(responsesChannel, deferResponsesChannel, wg)
		}
		log.Println("Next...")
	}
	log.Println("Wait...")
	wg.Wait()
}
