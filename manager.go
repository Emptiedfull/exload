package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"sync"
)

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
	file           *os.File
	UrlMap         map[string][]*server
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
		start:          int(*con.Proxy_settings.Startup_servers),
		free_ports:     con.Proxy_settings.Free_ports,
		active_servers: 0,
		total_req:      0,
		total_servers:  total,
		free_servers:   int8(total),
		in_progress:    0,
		servers:        []*server{},
		host:           "0.0.0.0",
		file:           logFile,
		UrlMap:         make(map[string][]*server),
	}

	m.UrlMap = m.setupUrlMap(con.ServerOptions, int(*con.Proxy_settings.Startup_servers))
	fmt.Println(m.UrlMap)

	return m
}

func (m *manager) setupUrlMap(s map[string]ServerOption, start int) map[string][]*server {
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, ser := range s {
		fmt.Println(ser.Prefix, ser.Command, ser.Args)

		var srvArr = make([]*server, 0, start)
		for i := 0; i < start; i++ {
			wg.Add(1)
			go func(command string, args []string) {
				defer wg.Done()
				srv, err := m.gen_one(command, args)
				if srv == nil {
					m.logErr("error starting", err)
				} else {
					mu.Lock()
					srvArr = append(srvArr, srv)
					mu.Unlock()
				}
			}(ser.Command, ser.Args)
		}

		wg.Wait()

		m.UrlMap[ser.Prefix] = srvArr

	}

	return m.UrlMap
}

func (m *manager) logStr(s ...interface{}) {
	go m.logFile.Println(s...)
}

func (m *manager) logErr(s string, e error) {
	go m.logFile.Println(s, e)
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

	m.logStr("Server started on", port)

	m.in_progress = m.in_progress - 1
	m.active_servers++

	s := &server{url, port, 0, cmd, "/"}

	done <- s

	m.servers = append(m.servers, s)

	err = cmd.Wait()
	if err != nil {
		m.logErr("server closed with", err)
		return
	}

}

func (m *manager) gen_one(com string, args []string) (*server, error) {
	if m.active_servers < int8(m.total_servers) {
		m.logStr("upscaling")
		port, err := m.popPort()
		if err != nil {
			return nil, err
		}

		nargs := make([]string, len(args))

		for i, arg := range args {
			if arg == "<port>" {
				nargs[i] = strconv.Itoa(int(port))
			} else {
				nargs[i] = arg
			}
		}

		var cmd *exec.Cmd = exec.Command(com, nargs...)

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

	m.logStr("Proxying request ", url, " to ", lcs.url)

	lcs.request(url, w, r)

}

func (m *manager) descale() {

	if m.active_servers <= int8(m.start) {
		return
	}
	m.logStr("descaling")
	se := m.servers[0]
	m.servers = m.servers[1:]

	se.cmd.Process.Kill()
	m.active_servers = m.active_servers - 1
	m.free_ports = append(m.free_ports, se.port)
	m.free_servers++

}

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
