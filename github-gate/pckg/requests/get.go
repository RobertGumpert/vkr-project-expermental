package requests

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
)

func Deserialize(t interface{}, response *http.Response) error {
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, t)
	if err != nil {
		return err
	}
	return nil
}

func GET(client *http.Client, url string, headers map[string]string) (*http.Response, error) {
	return req("GET", client, url, headers, nil)
}

func NewGET(url string, headers map[string]string) (*http.Response, error){
	client := new(http.Client)
	return req("GET", client, url, headers, nil)
}

func POST(client *http.Client, url string, headers map[string]string, body interface{}) (*http.Response, error) {
	return req("POST", client, url, headers, body)
}

func NewPOST(url string, headers map[string]string, body interface{}) (*http.Response, error){
	client := new(http.Client)
	return req("POST", client, url, headers, body)
}

func req(met string, client *http.Client, url string, headers map[string]string, body interface{}) (*http.Response, error) {
	var (
		req              *http.Request
		err              error
		jsonBodyIOReader io.Reader
	)
	if met == "GET" {
		req, err = http.NewRequest(met, url, nil)
	} else {
		if body != nil {
			b, err := json.Marshal(body)
			if err != nil {
				return nil, err
			}
			jsonBodyIOReader = bytes.NewReader(b)
		}
	}
	if met == "POST" {
		req, err = http.NewRequest(met, url, jsonBodyIOReader)
	}
	if err != nil {
		return nil, err
	}
	if headers != nil {
		for key, val := range headers {
			req.Header.Set(key, val)
		}
	}
	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return response, nil
}
