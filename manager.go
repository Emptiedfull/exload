package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
)

type manager struct {
	free_ports     []int32
	active_servers int8
	free_servers   int8
	in_progress    int
	total_servers  int
	servers        []*server
	host           string
	logFile        *log.Logger
	file           *os.File
	UrlMap         map[string]*pen

	mu sync.RWMutex
}

type pen struct {
	servers []*server
	command command
	con     int
}

type command struct {
	com  string
	args []string
}

func NewManager(con Config) *manager {
	total := len(con.Proxy_settings.Free_ports)

	logFile, err := os.OpenFile("logs/proxy.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("Error opening log file")
	}

	logger := log.New(logFile, "", log.LstdFlags)

	m := &manager{
		logFile:        logger,
		free_ports:     con.Proxy_settings.Free_ports,
		active_servers: 0,
		total_servers:  total,
		free_servers:   int8(total),
		in_progress:    0,
		servers:        []*server{},
		host:           "0.0.0.0",
		file:           logFile,
		UrlMap:         make(map[string]*pen),
		mu:             sync.RWMutex{},
	}

	go m.dyno(*con.Proxy_settings.Scale_pings)
	m.UrlMap = m.setupUrlMap(con.ServerOptions)

	return m
}

func (m *manager) setupUrlMap(s map[string]ServerOption) map[string]*pen {

	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, ser := range s {
		fmt.Println(ser.Prefix, ser.Command, ser.Args)

		var srvArr = make([]*server, 0)
		for i := 0; i < int(*ser.Startup_servers); i++ {
			wg.Add(1)
			go func(comm command) {
				defer wg.Done()
				srv, err := m.gen_one(comm)
				if srv == nil {
					m.logErr("error starting", err)
				} else {
					mu.Lock()
					srvArr = append(srvArr, srv)
					mu.Unlock()
				}
			}(command{ser.Command, ser.Args})
		}

		wg.Wait()
		fmt.Println(srvArr)
		m.UrlMap[ser.Prefix] = &pen{servers: srvArr, command: command{com: ser.Command, args: ser.Args}, con: 0}

	}

	return m.UrlMap
}

func (m *manager) logStr(s ...interface{}) {
	go m.logFile.Println(s...)
}

func (m *manager) logErr(s string, e error) {
	red := color.New(color.FgRed).SprintFunc()
	go m.logFile.Println(red(fmt.Sprintf("%s: %v", s, e)))
}

func (m *manager) create(port int32, cmd *exec.Cmd, done chan<- *server) {

	fmt.Println("starting server on", port, cmd)

	m.in_progress++
	m.free_servers = m.free_servers - 1

	// var cmd *exec.Cmd = exec.Command("uvicorn", "server:app", "--port", strconv.Itoa(int(port)), "--host", m.host)
	cmd.Env = append(os.Environ(), "VIRTUAL_ENV=/venv", "PATH=/venv/bin:"+os.Getenv("PATH"))

	logFile, err := os.OpenFile("logs/server_logs/"+strconv.Itoa(int(port))+".log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		m.logErr("error opening logs", err)
	}
	defer logFile.Close()

	cmd.Stdout = logFile
	cmd.Stderr = logFile

	err = cmd.Start()
	if err != nil {
		m.logErr("error starting server", err)
		return
	}

	ch := make(chan string)
	url := fmt.Sprintf("http://%s:%d", m.host, port)
	go wait_for_startup(url, ch)

	<-ch
	close(ch)

	m.logStr("Server started on", port)

	m.in_progress = m.in_progress - 1
	m.active_servers++

	s := &server{url, port, 0, cmd, "/", sync.RWMutex{}}

	done <- s
	close(done)

	m.servers = append(m.servers, s)

	err = cmd.Wait()
	if err != nil {
		m.logErr("server closed with", err)
		return
	}

}

func (m *manager) gen_one(comm command) (*server, error) {
	if m.active_servers < int8(m.total_servers) {

		port, err := m.popPort()
		if err != nil {
			return nil, err
		}

		nargs := make([]string, len(comm.args))

		for i, arg := range comm.args {
			if arg == "<port>" {
				nargs[i] = strconv.Itoa(int(port))
			} else {
				nargs[i] = arg
			}
		}

		var cmd *exec.Cmd = exec.Command(comm.com, nargs...)

		done := make(chan *server)

		go m.create(port, cmd, done)

		return <-done, nil
	} else {
		return nil, nil
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

	parts := strings.SplitN(url, "/", -1)

	if len(parts) < 3 {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		fmt.Println(parts)
		return
	}
	path := "/" + strings.Join(parts[2:], "/")
	dest := "/" + parts[1]

	m.mu.RLock()         // Use RLock for concurrent reads
	defer m.mu.RUnlock() // Ensure RUnlock after the operation

	pen, ok := m.UrlMap[dest]
	if !ok || pen == nil {
		http.Error(w, "Destination not found", http.StatusNotFound)
		return
	}
	pen.con++

	srv_options := pen.servers
	if len(srv_options) == 0 {
		http.Error(w, "No servers available", http.StatusServiceUnavailable)
		return
	}

	var lcs *server

	var min int = int(^uint(0) >> 1)

	fmt.Println(srv_options)

	for _, srv := range srv_options {
		if srv.req < min {
			lcs = srv
			min = srv.req

		}
	}

	if lcs == nil {
		lcs = m.UrlMap["/api"].servers[0]
	}

	lcs.request(path, w, r)
	pen.con = pen.con - 1

}

func (m *manager) upscale(prefix string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	p := m.UrlMap[prefix]
	srv, err := m.gen_one(p.command)
	if err != nil {
		m.logErr("error upscaling", err)
	}

	p.servers = append(p.servers, srv)
}

func (m *manager) descale(prefix string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	p := m.UrlMap[prefix]
	if len(p.servers) > 3 {
		srv := p.servers[0]
		srv.terminate()
		p.servers = p.servers[1:]
	} else {
		m.logErr("minimum server limit", fmt.Errorf("2"))
	}
	fmt.Println(m.UrlMap[prefix].servers)
}

func (m *manager) dyno(m_pings int) {
	ticker := time.NewTicker(1 * time.Second)

	var descale int = 0
	var upscale int = 0

	var upscale_rate int = 2
	var descale_rate int = 6

	go func() {

		for range ticker.C {
			for pre, pen := range m.UrlMap {

				if len(pen.servers) > 0 {
					load := pen.con / len(pen.servers)
					fmt.Println(pre, pen.con, load)

					if load > 10 {
						descale = 0
						upscale++

						if upscale >= upscale_rate {
							fmt.Println("trying to upscale")
							upscale = 0
							go m.upscale(pre)
						}
					}

					if load < 10 {
						upscale = 0
						descale++

						if descale >= descale_rate {
							fmt.Println("trying to descale")
							descale = 0
							m.descale(pre)
						}
					}
				}

			}

		}
	}()

}

// func (m *manager) descale() {

// 	if m.active_servers <= int8(m.start) {
// 		return
// 	}
// 	m.logStr("descaling")
// 	se := m.servers[0]
// 	m.servers = m.servers[1:]

// 	se.cmd.Process.Kill()
// 	m.active_servers = m.active_servers - 1
// 	m.free_ports = append(m.free_ports, se.port)
// 	m.free_servers++

// }

// func (m *manager) monitor(max int32) {
// 	var pings int
// 	const sleepDuration = 1 * time.Second
// 	ticker := time.NewTicker(sleepDuration)
// 	defer ticker.Stop()

// 	for range ticker.C {
// 		time.Sleep(1 * time.Second)

// 		if m.active_servers == 0 {
// 			continue
// 		}

// 		rps := int32(m.total_req / int16(m.active_servers))
// 		if rps > max {

// 			m.gen_one()
// 		}

// 		if rps < max {
// 			pings++

// 			if pings > 10 {

// 				m.descale()
// 				pings = 0
// 			}
// 		}
// 	}
// }
