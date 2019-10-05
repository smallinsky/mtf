package main

import (
	"context"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

func main() {
	time.Sleep(time.Millisecond * 990)
	req, err := http.NewRequest(http.MethodGet, "https://example.com/urlpath", nil)
	if err != nil {
		log.Fatalf("failed to create http request: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		log.Fatalf("failed to call api: %v", err)
	}
	buff, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("failed to read response body: %v", err)

	}
	log.Printf("Response body: '%s'", string(buff))
}
