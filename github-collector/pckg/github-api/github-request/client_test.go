package github_request

import (
	"fmt"
	"log"
	"sync"
	"testing"
)

const token = "aef3219befb0b5a71ebfcf5876dd8c8d9eeb0077"

func TestFlowNewClient(t *testing.T) {
	_, err := NewGithubClient(token)
	if err != nil {
		t.Fatal(err)
	}
	log.Println("OK.")
}

func TestFlowRateLimit(t *testing.T) {
	c, err := NewGithubClient("")
	if err != nil {
		t.Fatal(err)
	}
	wg := new(sync.WaitGroup)
	for i := 0; i < 121; i++ {
		var (
			waitSignal    = make(chan bool)
			resultChannel = make(chan *Response)
		)
		go c.Request("https://api.github.com/repos/facebook/react/topics", nil, CORE, waitSignal, resultChannel)
		select {
		case <-waitSignal:
			wg.Add(1)
			go func(resultChannel chan *Response, wg *sync.WaitGroup) {
				defer wg.Done()
				response := <-resultChannel
				fmt.Println(response.Response.StatusCode)
				if response.Err != nil {
					log.Println(err)
				}
				return
			}(resultChannel, wg)
		case response := <-resultChannel:
			if response.Err == nil {
				log.Println(response.Response.StatusCode)
			}
		}
		log.Println("Next...")
	}
	log.Println("Wait...")
	wg.Wait()
}
