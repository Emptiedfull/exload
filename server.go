package main

import (
	"io"
	"net/http"
	"os/exec"
	"time"
)

type server struct {
	url    string
	port   int32
	req    int
	cmd    *exec.Cmd
	prefix string
}

func (s *server) request(url string, w http.ResponseWriter, r *http.Request) {

	s.req = s.req + 1

	req_url := s.url + url

	req, err := http.NewRequest("GET", req_url, nil)
	if err != nil {
		http.Error(w, "Failed to request", http.StatusInternalServerError)
		return
	}

	headers := make(map[string]string)
	for name, values := range r.Header {
		for _, value := range values {
			headers[name] = value
			req.Header.Add(name, value)
		}
	}

	cookies := make(map[string]string)
	for _, cookie := range r.Cookies() {
		cookies[cookie.Name] = cookie.Value
		req.AddCookie(&http.Cookie{Name: cookie.Name, Value: cookie.Value})
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Failed to request", http.StatusInternalServerError)
	}

	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	w.Write(body)

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
