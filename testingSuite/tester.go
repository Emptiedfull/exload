package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	url := "http://localhost:8000/api/ping"
	// sock := "/tmp/b69f04c1d7.sock"
	iterations := 10000

	start := time.Now()

	err, _ := ramp(url)
	// err := createRequestUnix("/", sock)

	fmt.Println("no of requests", iterations)
	fmt.Println("errors:", err)
	fmt.Println(time.Since(start) / time.Duration(iterations))

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
	// defer res.Body.Close()

	// // Read the response body
	// body, err := io.ReadAll(res.Body)
	// if err != nil {
	// 	return err
	// }

	// // Convert the body to a string and print it
	// bodyStr := string(body)
	// fmt.Println(bodyStr)

	// return nil

	return nil
}

func createRequestUnix(url string, sock string) error {
	transport := &http.Transport{
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			return net.Dial("unix", sock)
		},
	}

	client := &http.Client{
		Transport: transport,
	}

	req, err := http.NewRequest("GET", "http://unix"+url, nil)
	if err != nil {
		fmt.Println(err)
		return err
	}

	_, err = client.Do(req)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil

}

func doXrequests(num int, url string) (errz uint64) {
	var errors atomic.Uint64
	var wg sync.WaitGroup
	for i := 1; i <= num; i++ {
		wg.Add(1)

		go func() {
			// err := createRequestUnix("/", sock)
			err := createRequest(url)
			if err != nil {
				errors.Add(1)
			}
			wg.Done()
		}()

	}

	wg.Wait()
	return errors.Load()
}

func ramp(url string) (errz int, timeT int) {

	timeTaken := 0
	errors := 0

	for i := 1; i <= 20; i++ {
		err := doXrequests(i*100, url)
		errors += int(err)
		timeTaken += 1
		fmt.Println(fmt.Sprintf("%d requests completed", i*100))
		time.Sleep(1 * time.Second)
	}

	return errors, timeTaken
}
