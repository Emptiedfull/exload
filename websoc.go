package main

import (
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
)

func (m *manager) proxyWebSocket(w http.ResponseWriter, r *http.Request, url string) {
	parts := strings.SplitN(url, "/", -1)

	if len(parts) < 3 {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		fmt.Println(parts)
		return
	}
	path := "/" + strings.Join(parts[2:], "/")
	dest := "/" + parts[1]

	pen := m.UrlMap[dest]

	if pen == nil {
		http.Error(w, "Bad Url", http.StatusBadRequest)
		fmt.Println("pen not found", dest)
		return
	}

	var lcs *server
	min := int(^uint(0) >> 1)

	srvs := pen.servers
	if len(srvs) == 0 {
		http.Error(w, "No servers available", http.StatusFailedDependency)
		fmt.Println("no servers")
		return
	}

	for _, srv := range srvs {
		if int(srv.con.Load()) < min {
			lcs = srv
			min = int(srv.con.Load())
		}
	}

	lcs.con.Add(1)

	socPath := lcs.sock
	fmt.Println("sening to", socPath)

	UnConn, err := net.Dial("unix", socPath)
	if err != nil {
		http.Error(w, "bad server pls try again", http.StatusFailedDependency)
		fmt.Println("failed at 1", err)
		return
	}

	target := "ws://unix" + path

	dialer := websocket.Dialer{
		NetDial: func(network, addr string) (net.Conn, error) {
			return UnConn, nil
		},
	}

	serverConn, _, err := dialer.Dial(target, nil)
	if err != nil {
		http.Error(w, "bad server pls try again", http.StatusFailedDependency)
		fmt.Println("failed at 2", err, target)
		return
	}

	upgrader = websocket.Upgrader{}
	clientConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "cant connect to client", http.StatusBadRequest)
		fmt.Println("failed at 3")
		return
	}

	go relay(clientConn, serverConn, pen, lcs)
	go relay(serverConn, clientConn, pen, lcs)

}

func relay(dst, srv *websocket.Conn, p *pen, s *server) {
	defer dst.Close()
	defer srv.Close()

	p.con.Add(1)

	for {
		messageType, message, err := srv.ReadMessage()
		if err != nil {

			break
		}

		err = dst.WriteMessage(messageType, message)
		if err != nil {

			break
		}

	}

	p.con.Add(-1)
	s.con.Add(-1)

}
