package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"time"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Proxy_port     int32  `yaml:"port"`
	Static_path    string `yaml:"static_path"`
	Proxy_settings struct {
		Free_ports      []int32 `yaml:"free_ports"`
		Max_load        int32   `yaml:"max_load"`
		Startup_servers int8    `yaml:"startup_servers"`
	}
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
	start          int
	free_ports     []int32
	active_servers int8
	total_req      int16
	free_servers   int8
	in_progress    int
	total_servers  int
	servers        []*server
	host           string
	logFile        *log.Logger
}

func NewManager(con Config) *manager {
	total := len(con.Proxy_settings.Free_ports)

	logFile, err := os.OpenFile("proxy.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("Error opening log file")
	}
	defer logFile.Close()

	logger := log.New(logFile, "", log.LstdFlags)

	m := &manager{
		logFile:        logger,
		start:          int(con.Proxy_settings.Startup_servers),
		free_ports:     con.Proxy_settings.Free_ports,
		active_servers: 0,
		total_req:      0,
		total_servers:  total,
		free_servers:   int8(total),
		in_progress:    0,
		servers:        []*server{},
		host:           "0.0.0.0",
	}

	m.create_mass(con.Proxy_settings.Startup_servers)
	go m.monitor(con.Proxy_settings.Max_load)

	return m
}

func (m *manager) create(port int32) {

	m.in_progress++
	m.free_servers = m.free_servers - 1

	var cmd *exec.Cmd = exec.Command("uvicorn", "server:app", "--port", strconv.Itoa(int(port)), "--host", m.host)
	cmd.Env = append(os.Environ(), "VIRTUAL_ENV=/venv", "PATH=/venv/bin:"+os.Getenv("PATH"))

	logFile, err := os.OpenFile("logs/"+strconv.Itoa(int(port))+".log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
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
	url := fmt.Sprintf("http://%s:%d", m.host, port)
	go wait_for_startup(url, ch)

	<-ch

	fmt.Println("starting server on:", port)

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

func (m *manager) create_mass(n int8) {
	for range n {
		m.gen_one()
	}
}

func (m *manager) gen_one() {
	if m.active_servers < int8(m.total_servers) {
		fmt.Println("upscaling")
		port, err := m.popPort()
		if err != nil {
			return
		}

		go m.create(port)
	}
}

func (m *manager) popPort() (int32, error) {
	if len(m.free_ports) == 0 {
		return 0, fmt.Errorf("no free ports")
	}
	port := m.free_ports[0]
	m.free_ports = m.free_ports[1:]
	return port, nil
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

	go m.logFile.Println("Proxying request to:", lcs.url)

	lcs.request(url, w, r)

}

func (m *manager) descale() {

	if m.active_servers <= int8(m.start) {
		return
	}
	fmt.Println("descaling")
	se := m.servers[0]
	m.servers = m.servers[1:]

	se.cmd.Process.Kill()
	m.active_servers = m.active_servers - 1
	m.free_ports = append(m.free_ports, se.port)
	m.free_servers++

}

func (m *manager) monitor(max int32) {
	var pings int

	for {

		time.Sleep(1 * time.Second)

		if m.active_servers == 0 {
			continue
		}

		rps := int32(m.total_req / int16(m.active_servers))
		if rps > max {

			m.gen_one()
		}

		if rps < max {
			pings++

			if pings > 10 {

				m.descale()
				pings = 0
			}
		}

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
		fmt.Println("huh")
	}

	m := NewManager(config)

	fs := http.FileServer(http.Dir(config.Static_path))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	fmt.Println(m.servers)
	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {

		http.ServeFile(w, r, "static/favicon.ico")
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		url := r.URL.Path
		m.proxy(url, w, r)
	})

	http.ListenAndServe(":"+strconv.Itoa(int(config.Proxy_port)), nil)
}
