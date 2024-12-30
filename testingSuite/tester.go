package main

import (
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	url := "http://localhost:33327/api/ping"
	iterations := 10000

	start := time.Now()

	var errors atomic.Uint64
	var wg sync.WaitGroup
	for i := 1; i < iterations; i++ {
		wg.Add(1)

		go func() {
			err := createRequest(url)
			if err != nil {
				errors.Add(1)
			}
			wg.Done()
		}()

	}

	wg.Wait()

	fmt.Println("errors:", errors.Load())
	fmt.Println(time.Since(start))
}

func createRequest(url string) error {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	_, err = client.Do(req)
	if err != nil {
		return err
	}

	return nil
}
