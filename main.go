package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"time"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Proxy_port     int32 `yaml:"port"`
	proxy_settings struct {
		free_ports      []int32 `yaml:"free_ports"`
		max_load        int32   `yaml:"max_load"`
		startup_servers int8    `yaml:"startup_servers"`
	}
	static_path string `yaml:"static_path"`
}

type server struct {
	url  string
	port int32
	req  int
	cmd  *exec.Cmd
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

type manager struct {
	free_ports     []int32
	active_servers int8
	total_req      int16
	free_servers   int8
	in_progress    int
	servers        []*server
	host           string
}

func NewManager(ports []int32, start int8) *manager {
	total := len(ports)
	m := &manager{
		free_ports:     ports,
		active_servers: 0,
		total_req:      0,
		free_servers:   int8(total),
		in_progress:    0,
		servers:        []*server{},
		host:           "0.0.0.0",
	}

	pts := ports[:start]
	fmt.Println(pts)
	m.create_mass(pts)
	go m.monitor()

	return m
}

func (m *manager) create(port int32, host string) {

	m.in_progress++
	m.free_servers = m.free_servers - 1

	var cmd *exec.Cmd = exec.Command("uvicorn", "server:app", "--port", strconv.Itoa(int(port)), "--host", host)
	cmd.Env = append(os.Environ(), "VIRTUAL_ENV=/venv", "PATH=/venv/bin:"+os.Getenv("PATH"))

	logFile, err := os.OpenFile("logs/process.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("Error opening log file")
	}
	defer logFile.Close()

	cmd.Stdout = logFile
	cmd.Stderr = logFile

	err = cmd.Start()
	if err != nil {
		fmt.Println("Error starting cmd")
		return
	}

	ch := make(chan string)
	url := fmt.Sprintf("http://%s:%d", host, port)
	go wait_for_startup(url, ch)

	<-ch
	m.in_progress = m.in_progress - 1
	m.active_servers++

	s := &server{url, port, 0, cmd}
	m.servers = append(m.servers, s)

	err = cmd.Wait()
	if err != nil {
		fmt.Println("Server closed with ", err)
		return
	}

}

func (m *manager) create_mass(ports []int32) {
	for _, port := range ports {
		go m.create(port, m.host)
	}
}

func (m *manager) proxy(url string, w http.ResponseWriter, r *http.Request) {
	var lcs *server
	minReq := int(^uint(0) >> 1)

	m.total_req++

	for _, srv := range m.servers {
		if srv.req < minReq {

			minReq = srv.req
			lcs = srv

		}
	}

	if lcs == nil {
		w.Write([]byte("internal server error"))
		return
	}

	lcs.request(url, w, r)

}

func (m *manager) monitor() {

	for {
		time.Sleep(1 * time.Second)
		fmt.Print("\033[H\033[2J")

		request_last_second := m.total_req
		fmt.Println(request_last_second)
		m.total_req = 0

	}
}

func main() {

	var config Config

	file, err := os.Open("config.yaml")
	if err != nil {
		fmt.Println("Error opening config file:", err)
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		fmt.Println("Error reading config file:", err)
		return
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		fmt.Errorf(err.Error())
	}

	m := NewManager(config.proxy_settings.free_ports, 2)

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	fmt.Println(m.servers)
	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {

		http.ServeFile(w, r, "static/favicon.ico")
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		url := r.URL.Path
		m.proxy(url, w, r)
	})

	http.ListenAndServe(":8080", nil)
}
