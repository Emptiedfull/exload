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
	cons string
}

func monitor(m *manager, url string, w http.ResponseWriter, r *http.Request) {

	switch {
	case url == "/":
		if con.Dynos.Monitor {
			res := index()
			res.Render(context.Background(), w)
		} else {
			res := monitor_err()
			res.Render(context.TODO(), w)

		}
	case url == "/pen_dashboard/getpentable":
		res := penTableItems(getPenFormatted(m))
		res.Render(context.Background(), w)
	case url == "/pen_dashboard":
		res := pen_dashboard(getPenFormatted(m))
		res.Render(context.Background(), w)
	case url == "/main_dashboard":
		res := main_dashboard()
		res.Render(context.Background(), w)
	case strings.HasPrefix(url, "/toggle"):
		t := strings.TrimPrefix(url, "/toggle")
		switch t {
		case "monitor":
			m.toggleMonitorDyno()
			res := statusUpdate("monitor", con.Dynos.Monitor)
			res.Render(context.Background(), w)
		case "scale":
			m.toggleScaleDyno()
			res := statusUpdate("scale", con.Dynos.Scaler)
			res.Render(context.Background(), w)
		}

	case url == "/enableMonitor":
		m.startMonitorDyno(false)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Monitor enabled"))
	case url == "/disableMonitor":
		m.EndMonitorDyno()
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte("monitor disable"))
	case url == "/disableScale":
		m.endScaleDyno()
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte("scaler disable"))
	case url == "/enableScale":
		m.startScaleDyno(false)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("scaler enabled"))
	case strings.HasPrefix(url, "/pen_chart/"):
		chartType := strings.TrimPrefix(url, "/pen_chart/")
		res := penChart(chartType)
		res.Render(context.Background(), w)
	case strings.HasPrefix(url, "/status/"):
		t := strings.TrimPrefix(url, "/status/")
		switch t {
		case "monitor":
			res := statusUpdate("monitor", con.Dynos.Monitor)
			res.Render(context.Background(), w)
		case "scale":
			res := statusUpdate("scale", con.Dynos.Scaler)
			res.Render(context.Background(), w)
		}

	case strings.HasPrefix(url, "/ws"):

		wsHandler(url, w, r, m)
	default:
		fmt.Println(url)
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 not found"))
	}

}

func (m *manager) monitorDyno(quit <-chan bool) {
	fmt.Println("monitor dyno started")
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if len(m.conn["main_dashboard"]) != 0 {

				gh := graphUpdate(m)

				rps := strconv.Itoa(gh.RPS)
				mem := strconv.Itoa(gh.MEM)
				cons := strconv.Itoa(getConns(m.UrlMap))
				total, active := getTotalPorts(m)
				srvs := strconv.Itoa(active) + "/" + strconv.Itoa(total)

				nS := State{rps, mem, srvs, cons}
				go sendUpdate(m.conn["main_dashboard"], nS)
			}

			pengraphUpdate(m)
		case <-quit:
			fmt.Println("Monitor dyno killed")
			return
		}
	}
}

type graph_update struct {
	RPS int `json:"rps"`
	MEM int `json:"mem"`
}

func pengraphUpdate(m *manager) {

	if len(m.conn["pen_graph/Requests"]) != 0 {
		pens := make(map[string]int, 0)
		for pre, pen := range m.UrlMap {
			pens[pre] = getRpsByPen(pen)
		}

		for _, client := range m.conn["pen_graph/Requests"] {
			client.conn.WriteJSON(pens)
		}
	}

	if len(m.conn["pen_graph/Memory"]) != 0 {
		pens := make(map[string]int, 0)
		for pre, pen := range m.UrlMap {
			pens[pre] = memInfo(pen) / 1024 / 1024
		}

		for _, client := range m.conn["pen_graph/Memory"] {
			client.conn.WriteJSON(pens)
		}
	}

	if len(m.conn["/admin/ws/pen_graph/Connections"]) != 0 {

		pens := make(map[string]int, 0)
		for pre, pen := range m.UrlMap {
			pens[pre] = int(pen.con.Load())
		}

		for _, client := range m.conn["pen_graph/Connections"] {
			client.conn.WriteJSON(pens)
		}
	}

}

func graphUpdate(m *manager) graph_update {
	gh := graph_update{TotalRps(m.UrlMap), totalMem(m.UrlMap)}

	clients := m.conn["main_graph"]
	if len(clients) == 0 {
		return gh
	}
	for _, client := range clients {
		client.conn.WriteJSON(gh)
	}

	return gh
}

func sendUpdate(clients []*client, state State) {
	for _, client := range clients {
		payload := createPayload(&client.state, state)
		if len(payload) == 0 {
			continue
		}
		sendMass(client.conn, payload)
	}
}

// func send(conn *websocket.Conn, t templ.Component) {

// 	if conn == nil {
// 		return
// 	}

// 	var buf strings.Builder
// 	err := t.Render(context.Background(), &buf)
// 	if err != nil {
// 		fmt.Println("error sending", err)
// 	}

// 	conn.WriteMessage(websocket.TextMessage, []byte(buf.String()))

// }

func sendMass(conn *websocket.Conn, tl []templ.Component) {
	var bufmain strings.Builder
	for _, t := range tl {
		var buf strings.Builder
		err := t.Render(context.Background(), &buf)
		if err != nil {
			fmt.Println("error rendering template", err)
			continue
		}
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

	path := url[4:]

	m.conn[path] = append(m.conn[path], &client{conn, State{}})

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
