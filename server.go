package main

import (
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"sync"
	"time"
)

type server struct {
	url    string
	port   int32
	req    int
	cmd    *exec.Cmd
	prefix string
	mu     sync.RWMutex
}

func (s *server) request(url string, w http.ResponseWriter, r *http.Request) {

	s.mu.RLock()
	defer s.mu.RUnlock()
	s.req = s.req + 1

	req_url := s.url + url

	req, err := http.NewRequest("GET", req_url, nil)
	if err != nil {
		http.Error(w, "Failed to request", http.StatusInternalServerError)
		return
	}

	for name, values := range r.Header {
		for _, value := range values {
			req.Header.Add(name, value)
		}
	}

	for _, cookie := range r.Cookies() {
		req.AddCookie(&http.Cookie{Name: cookie.Name, Value: cookie.Value})
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Failed to request", http.StatusInternalServerError)
	}

	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	for name, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(name, value)

		}
	}

	for _, cookie := range resp.Cookies() {
		http.SetCookie(w, cookie)
	}

	w.Write(body)

}

func (s *server) terminate() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cmd.Process.Kill()
	fmt.Println(s.port, " killed")

}

func wait_for_startup(url string, ch chan<- string) {

	for {
		// fmt.Println("checking server", url)
		var req_url string = url
		req, _ := http.NewRequest("GET", req_url, nil)

		client := &http.Client{}
		_, err := client.Do(req)
		if err == nil {
			ch <- "started"
			break
		}

		time.Sleep(1 * time.Second)
	}
}
