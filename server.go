package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os/exec"
	"sync"
	"time"
)

type server struct {
	sock   string
	cmd    *exec.Cmd
	prefix string

	req int
	con int

	rps int

	mu    sync.RWMutex
	reqMu sync.RWMutex
	rpsMu sync.RWMutex
	conMu sync.RWMutex
}

func (s *server) request(url string, w http.ResponseWriter, r *http.Request, c *Cache) {

	s.mu.RLock()
	defer s.mu.RUnlock()

	s.req = s.req + 1

	defer func() {
		s.reqMu.Lock()
		s.req = s.req + 1
		s.reqMu.Unlock()
	}()

	transport := &http.Transport{
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			return net.Dial("unix", s.sock)
		},
	}

	client := &http.Client{
		Transport: transport,
	}

	req, err := http.NewRequest("GET", "http://unix"+url, nil)
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

	resp, err := client.Do(req)
	if err != nil {
		s.request(url, w, r, c)
		return
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

	// w.Write(body)
	WriteThrough(r, w, c, body)

}

func (s *server) terminate() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cmd.Process.Kill()
	fmt.Println(s.sock, " killed")

}

func wait_for_startup(s *server, ch chan<- string) {

	for {
		// fmt.Println("checking server", url)

		req, _ := http.NewRequest("GET", "http://unix/", nil)
		transport := &http.Transport{
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", s.sock)
			},
		}

		client := &http.Client{
			Transport: transport,
		}
		_, err := client.Do(req)
		if err == nil {
			ch <- "started"
			break
		}

		time.Sleep(1 * time.Second)
	}
}
