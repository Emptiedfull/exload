package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type State struct {
	rps  string
	mem  string
	srvs string
}

func createPayload(pS *State, nS State) []templ.Component {

	payload := make([]templ.Component, 0)
	if pS.rps != nS.rps {
		pS.rps = nS.rps
		payload = append(payload, update_status_item("rps", nS.rps))
	}
	if pS.mem != nS.mem {
		pS.mem = nS.mem
		payload = append(payload, update_status_item("mem", nS.mem))
	}
	if pS.srvs != nS.srvs {
		pS.srvs = nS.srvs
		payload = append(payload, update_status_item("srvs", nS.srvs))
	}

	return payload
}

func (m *manager) monitorDyno() {
	ticker := time.NewTicker(1 * time.Second)
	pS := State{}

	for range ticker.C {

		nS := State{}

		if m.conn["main_dashboard"] == nil {
			nS = State{}
			continue
		}
		fmt.Println("rick")
		rps := strconv.Itoa(TotalRps(m.UrlMap))
		mem := strconv.Itoa(totalMem(m.UrlMap))
		total, active := getTotalPorts(m)
		srvs := strconv.Itoa(active) + "/" + strconv.Itoa(total)

		nS = State{rps, mem, srvs}

		sendMass(m.conn["main_dashboard"], createPayload(&pS, nS))

	}

}

func monitor(m *manager, url string, w http.ResponseWriter, r *http.Request) {
	fmt.Println("monitor", url)
	switch {
	case url == "admin/":
		res := index()

		res.Render(context.Background(), w)
	case strings.HasPrefix(url, "/admin/ws"):
		wsHandler(url, w, r, m)
	default:
		res := index()
		res.Render(context.Background(), w)
	}

}

func send(conn *websocket.Conn, t templ.Component) {

	if conn == nil {
		fmt.Println("nil conn")
		return
	}

	var buf strings.Builder
	err := t.Render(context.Background(), &buf)
	if err != nil {
		fmt.Println("error sending", err)
	}

	conn.WriteMessage(websocket.TextMessage, []byte(buf.String()))

}

func sendMass(conn *websocket.Conn, tl []templ.Component) {
	var bufmain strings.Builder
	for _, t := range tl {
		var buf strings.Builder
		err := t.Render(context.Background(), &buf)
		if err != nil {
			fmt.Println("error rendering template", err)
			continue
		}
		fmt.Println(buf)
		bufmain.WriteString(buf.String())
	}

	if conn == nil {
		fmt.Println("nil conn")
		return
	}

	err := conn.WriteMessage(websocket.TextMessage, []byte(bufmain.String()))
	if err != nil {
		fmt.Println("error sending message", err)
	}

}

func wsHandler(url string, w http.ResponseWriter, r *http.Request, m *manager) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error upgrading")
		return
	}

	path := url[10:]
	m.conn[path] = conn

	for {

		_, _, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("closing conn", err)
			conn.Close()
			m.conn[path] = nil
			break
		}
	}
}
