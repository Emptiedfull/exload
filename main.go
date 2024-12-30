package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/websocket"
)

func main() {

	logs, _ := os.ReadDir("./logs/server_logs")
	for _, file := range logs {
		path := "./logs/server_logs/" + file.Name()
		os.RemoveAll(path)
	}

	err := getConfig()

	if err != nil {
		fmt.Println("Error loading config")
		os.Exit(1)
	}
	c := NewCache(100)
	// c.put("/api/ping", make([]byte, 5), 1*time.Hour)
	m := NewManager()
	defer m.file.Close()

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static", fs))

	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {

		http.ServeFile(w, r, "static/favicon.ico")
	})

	http.HandleFunc("/admin/enable-monitor", func(w http.ResponseWriter, r *http.Request) {
		if m.monitorQuit != nil {
			m.monitorQuit <- true
			con.Dynos.Monitor = false
		}
	})

	http.HandleFunc("/admin/", func(w http.ResponseWriter, r *http.Request) {
		if con.Dynos.Monitor {
			monitor(m, r.URL.String(), w, r)
		} else {

			res := monitor_err()
			res.Render(context.TODO(), w)

		}

	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		url := r.URL.String()
		if websocket.IsWebSocketUpgrade(r) {
			m.proxyWebSocket(w, r, url)
		} else {
			m.proxy(url, w, r, c)
		}

	})

	pid := os.Getpid()
	fmt.Printf("The PID of this process is %d\n", pid)

	port := strconv.Itoa(int(*con.Proxy_port))
	fmt.Printf("Starting server on port %s\n", port)

	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		fmt.Printf("Error starting server: %v\n", err)
	}
}
