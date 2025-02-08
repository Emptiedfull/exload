package main

import (
	"context"
	"fmt"
	"math/rand"
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

func monitor(conns map[string][]*client, url string, w http.ResponseWriter, r *http.Request) {

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
		penT := &penFormatted{
			name: "hi",
		}
		tb := make([]penFormatted, 0)
		tb = append(tb, *penT)

		res := penTableItems(tb)
		res.Render(context.Background(), w)
	case url == "/pen_dashboard":
		penT := []penFormatted{{
			name:   "Z1029",
			active: strconv.FormatFloat(10.2, 'f', 1, 64),
			max:    "82%",
			cmd:    "3 Years",
		}, {
			name:   "S2034",
			active: strconv.FormatFloat(9.6, 'f', 1, 64),
			max:    "97%",
			cmd:    "5 Years",
		}, {
			name:   "K9272",
			active: strconv.FormatFloat(11.2, 'f', 1, 64),
			max:    "72%",
			cmd:    "1 Year",
		}}

		res := pen_dashboard(penT)
		res.Render(context.Background(), w)
	case url == "/main_dashboard":
		res := main_dashboard()
		res.Render(context.Background(), w)
	case strings.HasPrefix(url, "/toggle"):
		t := strings.TrimPrefix(url, "/toggle")
		switch t {
		case "monitor":
			res := statusUpdate("monitor", con.Dynos.Monitor)
			res.Render(context.Background(), w)
		case "scale":
			res := statusUpdate("scale", con.Dynos.Scaler)
			res.Render(context.Background(), w)
		}

	case url == "/enableMonitor":

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Monitor enabled"))
	case url == "/disableMonitor":
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte("monitor disable"))
	case url == "/disableScale":
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte("scaler disable"))
	case url == "/enableScale":

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("scaler enabled"))
	case strings.HasPrefix(url, "/pen_chart/"):
		chartType := strings.TrimPrefix(url, "/pen_chart/")
		fmt.Println("ChartType:", chartType)
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

		wsHandler(conns, url, w, r)
	default:
		fmt.Println(url)
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 not found"))
	}

}

func monitorDyno(conns map[string][]*client, quit <-chan bool) {
	fmt.Println("monitor dyno started")
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if len(conns["main_dashboard"]) != 0 {

				gh := graphUpdate(conns)
				consR := rand.Intn(5) + 4
				totalR := 12
				ActiveR := 11

				rps := strconv.Itoa(gh.RPS)
				mem := strconv.Itoa(gh.MEM)
				cons := strconv.Itoa(consR)
				total, active := totalR, ActiveR
				srvs := strconv.Itoa(active) + "/" + strconv.Itoa(total)

				nS := State{rps, mem, srvs, cons}
				go sendUpdate(conns["main_dashboard"], nS)
			}

			pengraphUpdate(conns)
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

func pengraphUpdate(conns map[string][]*client) {

	if len(conns["pen_graph/Engine%20Torque"]) != 0 {
		fmt.Println("sending pen graph upare")
		pens := make(map[string]int, 0)
		// for pre, pen := range m.UrlMap {
		// 	pens[pre] = getRpsByPen(pen)
		// }
		pens["Z1023"] = rand.Intn(10) + 200
		pens["239T3"] = rand.Intn(10) + 200
		pens["K9272"] = rand.Intn(20) + 250

		for _, client := range conns["pen_graph/Engine%20Torque"] {
			client.conn.WriteJSON(pens)
		}
	}

	if len(conns["pen_graph/Engine%20Power"]) != 0 {
		fmt.Println("sending pen graph upare")
		pens := make(map[string]int, 0)
		// for pre, pen := range m.UrlMap {
		// 	pens[pre] = getRpsByPen(pen)
		// }
		pens["Z1023"] = rand.Intn(40) + 60
		pens["239T3"] = rand.Intn(40) + 60
		pens["K9272"] = rand.Intn(40) + 60

		for _, client := range conns["pen_graph/Engine%20Power"] {
			client.conn.WriteJSON(pens)
		}
	}

	if len(conns["pen_graph/Air%20Flow"]) != 0 {
		fmt.Println("sending pen graph upare")
		pens := make(map[string]int, 0)
		// for pre, pen := range m.UrlMap {
		// 	pens[pre] = getRpsByPen(pen)
		// }
		pens["Z1023"] = rand.Intn(80) + 120
		pens["239T3"] = rand.Intn(80) + 120
		pens["K9272"] = rand.Intn(80) + 120

		for _, client := range conns["pen_graph/Air%20Flow"] {
			client.conn.WriteJSON(pens)
		}
	}

}

func graphUpdate(conns map[string][]*client) graph_update {
	gh := graph_update{rand.Intn(20) + 8, rand.Intn(21) + 60}

	clients := conns["main_graph"]
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

func wsHandler(conns map[string][]*client, url string, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error upgrading")
		return
	}

	path := url[4:]

	conns[path] = append(conns[path], &client{conn, State{}})

	for {

		_, _, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("closing conn", err)
			conn.Close()
			conns[path] = nil
			break
		}
	}
}
