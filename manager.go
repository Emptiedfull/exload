package main

import (
	"crypto/sha256"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/gorilla/websocket"
)

type manager struct {
	host    string
	logFile *log.Logger
	file    *os.File
	UrlMap  map[string]*pen
	conn    map[string][]*client

	mu sync.RWMutex
}

type client struct {
	conn  *websocket.Conn
	state State
}

type pen struct {
	max_servers int8
	min_servers int8
	servers     []*server
	command     command
	con         int
	conMu       sync.RWMutex
}

type command struct {
	com  string
	args []string
}

func NewManager(con Config) *manager {

	logFile, err := os.OpenFile("logs/proxy.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("Error opening log file")
	}

	logger := log.New(logFile, "", log.LstdFlags)

	m := &manager{
		logFile: logger,
		host:    "0.0.0.0",
		file:    logFile,
		UrlMap:  make(map[string]*pen),
		mu:      sync.RWMutex{},
		conn:    make(map[string][]*client),
	}

	go m.Scaling_dyno(*con.Proxy_settings.Downscale_ping, *con.Proxy_settings.Upscale_ping, *con.Proxy_settings.scale_interval, int(*con.Proxy_settings.Max_load))

	m.UrlMap = m.setupUrlMap(con.ServerOptions)
	go m.monitorDyno()
	go m.rpsDyno()

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

		m.UrlMap[ser.Prefix] = &pen{servers: srvArr, command: command{com: ser.Command, args: ser.Args}, con: 0, max_servers: 4, min_servers: *ser.Startup_servers, conMu: sync.RWMutex{}}

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

func genSock() string {
	root := "/tmp/"
	currentTime := time.Now().String()
	hash := sha256.Sum256([]byte(currentTime))
	sockName := fmt.Sprintf("%x", hash)[:10] + ".sock"
	path := root + sockName
	return path
}

func (m *manager) create(sock string, cmd *exec.Cmd, done chan<- *server) {

	fmt.Println("starting server on", sock, cmd)

	// var cmd *exec.Cmd = exec.Command("uvicorn", "server:app", "--port", strconv.Itoa(int(port)), "--host", m.host)
	venvPath := "/venv"
	cmd.Env = append(os.Environ(), "VIRTUAL_ENV="+venvPath, "PATH="+venvPath+"/bin:"+os.Getenv("PATH"))

	logDir := "logs/server_logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		m.logErr("error creating log directory", err)
		return
	}

	logFile, err := os.OpenFile(logDir+"/"+sock[4:8]+".log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		m.logErr("error opening logs", err)
		return
	}
	defer logFile.Close()

	cmd.Stdout = logFile
	cmd.Stderr = logFile

	err = cmd.Start()
	if err != nil {
		m.logErr("error starting server", err)
		return
	}

	// s := &server{url, port, cmd, "/", 0, sync.RWMutex{}, sync.RWMutex{}}
	s := &server{sock: sock, cmd: cmd, prefix: "/", req: 0, reqMu: sync.RWMutex{}, mu: sync.RWMutex{}, conMu: sync.RWMutex{}, rpsMu: sync.RWMutex{}}

	ch := make(chan string)
	go wait_for_startup(s, ch)

	<-ch
	close(ch)

	m.logStr("Server started on", cmd.Process.Pid)

	done <- s
	close(done)

	err = cmd.Wait()
	for _, pen := range m.UrlMap {
		for i, srv := range pen.servers {
			if s == srv {
				fmt.Println("server removed")
				m.mu.Lock()
				pen.servers = append(pen.servers[:i], pen.servers[i+1:]...)
				m.mu.Unlock()
			}
		}
	}

	if err != nil {
		m.logErr("server closed with", err)
		return
	}
	fmt.Println("server ended")

}

func (m *manager) gen_one(comm command) (*server, error) {

	// port, err := m.popPort()
	// if err != nil {
	// 	return nil, err
	// }

	nargs := make([]string, len(comm.args))
	sock := genSock()

	for i, arg := range comm.args {
		if arg == "<sock>" {
			nargs[i] = sock
		} else {
			nargs[i] = arg
		}
	}

	var cmd *exec.Cmd = exec.Command(comm.com, nargs...)

	done := make(chan *server)

	go m.create(sock, cmd, done)

	return <-done, nil

}

// func (m *manager) popPort() (int32, error) {
// 	if len(m.free_ports) == 0 {
// 		return 0, fmt.Errorf("no free ports")
// 	}
// 	port := m.free_ports[0]
// 	m.free_ports = m.free_ports[1:]
// 	return port, nil
// }

func (m *manager) proxy(url string, w http.ResponseWriter, r *http.Request, cache *Cache) {

	hit, val := ReadThrough(r, cache)

	if hit {
		w.Write(val)

	} else {

		parts := strings.SplitN(url, "/", -1)

		if len(parts) < 3 {
			http.Error(w, "Invalid URL format", http.StatusBadRequest)
			fmt.Println(parts)
			return
		}
		path := "/" + strings.Join(parts[2:], "/")
		dest := "/" + parts[1]

		m.mu.RLock()
		defer m.mu.RUnlock()

		pen, ok := m.UrlMap[dest]
		if !ok || pen == nil {
			http.Error(w, "Destination not found", http.StatusNotFound)
			return
		}

		// pen.conMu.Lock()
		// pen.con++
		// pen.conMu.Unlock()

		srv_options := pen.servers
		if len(srv_options) == 0 {
			http.Error(w, "No servers available", http.StatusServiceUnavailable)
			return
		}

		var lcs *server

		var min int = int(^uint(0) >> 1)

		for _, srv := range srv_options {
			if srv.req < min {
				lcs = srv
				min = srv.req

			}
		}

		if lcs == nil {
			lcs = m.UrlMap["/api"].servers[0]
		}

		lcs.request(path, w, r, cache)

		// pen.conMu.Lock()
		// pen.con = pen.con - 1
		// pen.conMu.Unlock()
	}

}

func (m *manager) upscale(prefix string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	p := m.UrlMap[prefix]

	if len(p.servers) >= int(p.max_servers) {
		return
	}

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
	if len(p.servers) > int(p.min_servers) {
		srv := p.servers[0]
		srv.terminate()

		p.servers = p.servers[1:]
	} else {
		m.logErr("minimum server limit", fmt.Errorf("2"))
	}

}

func (m *manager) rpsDyno() {
	ticker := time.NewTicker(1 * time.Second)

	for range ticker.C {
		for _, pen := range m.UrlMap {
			for _, srv := range pen.servers {
				srv.rpsMu.Lock()
				srv.rps = srv.req
				srv.rpsMu.Unlock()

				srv.reqMu.Lock()
				srv.req = 0
				srv.reqMu.Unlock()
			}

		}
	}
}

func (m *manager) Scaling_dyno(d_pings int8, u_upings int8, interval int, m_load int) {
	ticker := time.NewTicker(time.Duration(interval) * time.Second)

	var descale int8 = 0
	var upscale int8 = 0

	var upscale_rate int8 = u_upings
	var descale_rate int8 = d_pings

	go func() {

		for range ticker.C {
			for pre, pen := range m.UrlMap {

				if len(pen.servers) > 0 {
					pen.conMu.RLock()
					load := getRpsByPen(pen) / len(pen.servers)

					pen.conMu.RUnlock()

					if len(pen.servers) < int(pen.min_servers) {
						go m.upscale(pre)
					}

					if load > m_load {
						descale = 0
						upscale++

						if upscale >= upscale_rate {

							upscale = 0
							go m.upscale(pre)
						}
					}

					if load < m_load {
						upscale = 0
						descale++

						if descale >= descale_rate {

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
